package jmthriftProto

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/pkg/errors"
	larkProto "gitlab.int.jumei.com/JMArch/go-rpc/lark-proto"
	"io"
)

/*
在 framed transport 的基础上添加了对 lark 协议的支持
*/

const DEFAULT_MAX_LENGTH = 16384000

type LarkFramedTransportClient struct {
	hostport  string
	transport thrift.TTransport
	frameSize uint32 //Current remaining size of the frame. if ==0 read next frame header
	buffer    [4]byte
	maxLength uint32

	bizBuf   *bytes.Buffer
	reqOpt   *larkProto.LarkReqOption
	respOpt  *larkProto.RespOption
	unpacked bool
	role     string
}

func NewLarkFramedTransportClient() *LarkFramedTransportClient {
	return &LarkFramedTransportClient{
		hostport:  "127.0.0.1:12311",
		maxLength: thrift.DEFAULT_MAX_LENGTH,
		bizBuf:    new(bytes.Buffer),
		reqOpt:    new(larkProto.LarkReqOption),
		respOpt:   new(larkProto.RespOption),
	}
}

func (t *LarkFramedTransportClient) WithHostport(hostport string) *LarkFramedTransportClient {
	t.hostport = hostport
	return t
}

func (t *LarkFramedTransportClient) WithLarkReqOption(opt *larkProto.LarkReqOption) *LarkFramedTransportClient {
	t.reqOpt = opt
	return t
}

func (t *LarkFramedTransportClient) WithTransport(trans thrift.TTransport) *LarkFramedTransportClient {
	t.transport = trans
	return t
}

func (t *LarkFramedTransportClient) RespOption() *larkProto.RespOption {
	return t.respOpt
}

func (t *LarkFramedTransportClient) Valid() (*LarkFramedTransportClient, error) {
	if t.transport == nil {
		trans, err := thrift.NewTSocket(t.hostport)
		if err != nil {
			return t, errors.WithStack(err)
		}

		if err := trans.Open(); err != nil {
			return t, errors.WithStack(err)
		}
		t.transport = trans
	}

	if t.reqOpt == nil || t.reqOpt.TargetService == "" {
		return t, errors.New("target srevice is empty")
	}
	return t, nil
}

func (t *LarkFramedTransportClient) MustValid() *LarkFramedTransportClient {
	if _, err := t.Valid(); err != nil {
		panic(err)
	}

	return t
}

func (p *LarkFramedTransportClient) Open() error {
	return p.transport.Open()
}

func (p *LarkFramedTransportClient) IsOpen() bool {
	return p.transport.IsOpen()
}

func (p *LarkFramedTransportClient) Close() error {
	return p.transport.Close()
}

func (p *LarkFramedTransportClient) Read(buf []byte) (l int, err error) {
	if err := p.unpack(); err != nil {
		return 0, err
	}

	if p.frameSize == 0 {
		p.frameSize, err = p.readFrameHeader()
		if err != nil {
			return
		}
	}
	if p.frameSize < uint32(len(buf)) {
		frameSize := p.frameSize
		tmp := make([]byte, p.frameSize)
		l, err = p.Read(tmp)
		copy(buf, tmp)
		if err == nil {
			err = thrift.NewTTransportExceptionFromError(fmt.Errorf("Not enough frame size %d to read %d bytes", frameSize, len(buf)))
			return
		}
	}
	got, err := p.bizBuf.Read(buf)
	p.frameSize = p.frameSize - uint32(got)
	//sanity check
	if p.frameSize < 0 {
		return 0, thrift.NewTTransportException(thrift.UNKNOWN_TRANSPORT_EXCEPTION, "Negative frame size")
	}
	return got, thrift.NewTTransportExceptionFromError(err)
}

func (p *LarkFramedTransportClient) ReadByte() (c byte, err error) {
	if err := p.unpack(); err != nil {
		return 0, err
	}

	if p.frameSize == 0 {
		p.frameSize, err = p.readFrameHeader()
		if err != nil {
			return
		}
	}
	if p.frameSize < 1 {
		return 0, thrift.NewTTransportExceptionFromError(fmt.Errorf("Not enough frame size %d to read %d bytes", p.frameSize, 1))
	}
	c, err = p.bizBuf.ReadByte()
	if err == nil {
		p.frameSize--
	}
	return
}

func (p *LarkFramedTransportClient) Write(buf []byte) (int, error) {
	p.unpacked = false
	n, err := p.bizBuf.Write(buf)
	return n, thrift.NewTTransportExceptionFromError(err)
}

func (p *LarkFramedTransportClient) WriteByte(c byte) error {
	p.unpacked = false
	return p.bizBuf.WriteByte(c)
}

func (p *LarkFramedTransportClient) WriteString(s string) (n int, err error) {
	p.unpacked = false
	return p.bizBuf.WriteString(s)
}

func (p *LarkFramedTransportClient) Flush() error {
	p.unpacked = false

	size := p.bizBuf.Len()
	var buf = make([]byte, 4+size)
	binary.BigEndian.PutUint32(buf[:4], uint32(size))
	copy(buf[4:], p.bizBuf.Bytes())

	b, err := larkProto.PackBizReq(buf, p.reqOpt)
	if err != nil {
		return thrift.NewTTransportExceptionFromError(err)
	}

	if _, err := p.transport.Write(b); err != nil {
		return thrift.NewTTransportExceptionFromError(err)
	}

	return thrift.NewTTransportExceptionFromError(p.transport.Flush())
}

func (p *LarkFramedTransportClient) readFrameHeader() (uint32, error) {
	buf := p.buffer[:4]
	if _, err := io.ReadFull(p.bizBuf, buf); err != nil {
		return 0, err
	}
	size := binary.BigEndian.Uint32(buf)
	if size < 0 || size > p.maxLength {
		return 0, thrift.NewTTransportException(thrift.UNKNOWN_TRANSPORT_EXCEPTION, fmt.Sprintf("Incorrect frame size (%d)", size))
	}
	return size, nil
}

func (p *LarkFramedTransportClient) RemainingBytes() (num_bytes uint64) {
	return uint64(p.frameSize)
}

func (p *LarkFramedTransportClient) unpack() error {
	if p.unpacked {
		return nil
	}

	p.unpacked = true

	b, o, err := larkProto.UnPackBizResp(p.transport)
	if err != nil {
		return err
	}

	p.respOpt = o
	p.bizBuf.Reset()
	p.bizBuf.Write(b)

	return nil
}
