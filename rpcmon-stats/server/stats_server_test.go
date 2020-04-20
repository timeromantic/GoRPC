package server

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
	"time"
)

func TestGetFailedLog(t *testing.T) {
	tm, _ := time.ParseInLocation("2006-01-02 15:04:05", "2019-12-24 23:00:00", time.Local)
	fmt.Println(tm.Format("2006-01-02"))

	fmt.Println(time.Unix(tm.Unix(), 0).Format("2006-01-02 15:04:05"))

	b, err := getFailedLog("/tmp/logs/", "MemberLevel", "getRefundRate", 0, 0, "504", "", 0, 10)
	fmt.Println(string(b))
	fmt.Printf("%+v \n", err)
}

func TestServer(t *testing.T) {
	s := NewStatsServer()
	err := s.Start()
	assert.Nil(t, err)

	var body = &Body{
		ModuleName:    "xx1",
		InterfaceName: "xx2",
	}
	msg := &rpcmonMessage{
		version:  1,
		seriesId: 2,
		cmd:      cmdProvider,
		subCmd:   subCmdGetClientRaw,
		code:     0,
		packLen:  0,
		header:   [15]byte{},
		body:     mustJsonMarshal(body),
	}
	b, err := msg.marshal()
	assert.Nil(t, err)

	fmt.Println(len(b))

	conn, err := net.Dial("udp", "127.0.0.1:20204")
	assert.Nil(t, err)

	n, err := conn.Write(b)
	assert.Nil(t, err)
	assert.Equal(t, n, len(b))

	var retb = make([]byte, 1024)
	conn.Read(retb)
	fmt.Println(string(retb))
}

func TestSample(t *testing.T) {
	go f1(t)
	time.Sleep(time.Second)
	f2(t)
	f2(t)
	f2(t)
	time.Sleep(time.Second)

}

func f1(t *testing.T) {
	addr, err := net.ResolveUDPAddr("udp", ":20200")
	assert.Nil(t, err)

	conn, err := net.ListenUDP("udp", addr)
	assert.Nil(t, err)

	for {
		fmt.Println("listen err:", err)
		var b = make([]byte, 10)
		conn.Read(b)
		fmt.Println(b)

		rd := bytes.NewReader(b)

		var b1 = make([]byte, 1)
		n, err := rd.Read(b1)
		fmt.Println(n, err)
		fmt.Println(b1)

		n, err = rd.Read(b1)
		fmt.Println(n, err)
		fmt.Println(b1)
	}
}

func f2(t *testing.T) {
	conn, err := net.Dial("udp", "127.0.0.1:20200")
	assert.Nil(t, err)

	conn.Write([]byte("123456"))
}
