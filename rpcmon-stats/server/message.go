package server

import (
	"bytes"
	"encoding/binary"
	"github.com/pkg/errors"
	"io"
	"reflect"
	"strconv"
)

const headerLen = 15

type rpcmonMessage struct {
	version  uint8
	seriesId uint16
	cmd      uint16
	subCmd   uint16
	code     int32
	packLen  uint32
	header   [headerLen]byte
	body     []byte
}

type Body struct {
	ModuleName         string      `json:"module"`
	InterfaceName      string      `json:"interface"`
	TimestampInterface interface{} `json:"Time"` // 接收的数据有时候是数字有时候是字符串
	Timestamp          int64       `json:"-"`    //存储以秒为单位的时间戳
	IP                 string      `json:"ip"`
}

func (b *Body) Verify() error {
	ty := reflect.TypeOf(b.TimestampInterface)
	switch ty.Kind() {
	case reflect.String:
		i, err := strconv.ParseInt(b.TimestampInterface.(string), 10, 64)
		if err != nil {
			return err
		}
		b.Timestamp = i

	case reflect.Float64:
		b.Timestamp = int64(b.TimestampInterface.(float64))

	case reflect.Float32:
		b.Timestamp = int64(b.TimestampInterface.(float32))

	case reflect.Int:
		b.Timestamp = int64(b.TimestampInterface.(int))

	case reflect.Int64:
		b.Timestamp = b.TimestampInterface.(int64)

	default:
		return errors.Errorf("verify body error. kind:%s, %+v", ty.Kind(), b)
	}

	if b.ModuleName == "PHPServer" {
		b.ModuleName = "GoServer"
	}

	return nil
}

func (m *rpcmonMessage) fill(r io.Reader) error {
	err := m.readHeaer(r)
	if err != nil {
		return err
	}

	err = m.readBody(r)
	return err
}

func (m *rpcmonMessage) readHeaer(r io.Reader) error {
	if _, err := io.ReadFull(r, m.header[:]); err != nil {
		return err
	}

	order := binary.LittleEndian
	rd := bytes.NewReader(m.header[:])
	var err error

	f := func(x interface{}) {
		if err != nil {
			return
		}

		err = binary.Read(rd, order, x)
	}

	f(&m.version)
	f(&m.seriesId)
	f(&m.cmd)
	f(&m.subCmd)
	f(&m.code)
	f(&m.packLen)

	return err
}

func (m *rpcmonMessage) readBody(r io.Reader) error {
	m.body = make([]byte, m.packLen-headerLen)

	_, err := io.ReadFull(r, m.body)

	return err
}

func (m *rpcmonMessage) marshal() ([]byte, error) {
	m.packLen = uint32(headerLen + len(m.body))

	buf := bytes.NewBuffer(nil)
	order := binary.LittleEndian

	var err error
	f := func(x interface{}) {
		if err != nil {
			return
		}

		err = binary.Write(buf, order, x)
	}

	f(m.version)
	f(m.seriesId)
	f(m.cmd)
	f(m.subCmd)
	f(m.code)
	f(m.packLen)
	if err != nil {
		return nil, err
	}

	_, err = buf.Write(m.body)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
