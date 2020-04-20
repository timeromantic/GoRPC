package jmthriftProto

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"git.apache.org/thrift.git/lib/go/thrift"
	larkProto "gitlab.int.jumei.com/JMArch/go-rpc/lark-proto"
	"io"
	"time"
)

/*
在 framed transport 的基础上添加了对 lark 协议的支持
*/

type LarkFramedTransportServer struct {
	transport thrift.TTransport
	frameSize uint32 //Current remaining size of the frame. if ==0 read next frame header
	buffer    [4]byte
	maxLength uint32
	bizBuf    *bytes.Buffer
	unpacked  bool
	reqOpt    *larkProto.ReqOption
	beginTime time.Time     //请求开始的数据
	costTime  time.Duration //一次请求的耗时时间
}

func NewLarkFramedTransportServer() *LarkFramedTransportServer {
	return &LarkFramedTransportServer{
		maxLength: thrift.DEFAULT_MAX_LENGTH,
		bizBuf:    new(bytes.Buffer),
	}
}

func (t *LarkFramedTransportServer) WithTransport(trans thrift.TTransport) *LarkFramedTransportServer {
	t.transport = trans
	return t
}

func (p *LarkFramedTransportServer) ReqOption() *larkProto.ReqOption {
	return p.reqOpt
}

func (p *LarkFramedTransportServer) Open() error {
	return p.transport.Open()
}

func (p *LarkFramedTransportServer) IsOpen() bool {
	return p.transport.IsOpen()
}

func (p *LarkFramedTransportServer) Close() error {
	return p.transport.Close()
}

func (p *LarkFramedTransportServer) Read(buf []byte) (l int, err error) {
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

func (p *LarkFramedTransportServer) ReadByte() (c byte, err error) {
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

func (p *LarkFramedTransportServer) Write(buf []byte) (int, error) {
	p.unpacked = false
	n, err := p.bizBuf.Write(buf)
	return n, thrift.NewTTransportExceptionFromError(err)
}

func (p *LarkFramedTransportServer) WriteByte(c byte) error {
	p.unpacked = false
	return p.bizBuf.WriteByte(c)
}

func (p *LarkFramedTransportServer) WriteString(s string) (n int, err error) {
	p.unpacked = false
	return p.bizBuf.WriteString(s)
}

func (p *LarkFramedTransportServer) Flush() error {
	p.unpacked = false
	p.costTime = time.Since(p.beginTime)

	size := p.bizBuf.Len()
	var buf = make([]byte, 4+size)
	binary.BigEndian.PutUint32(buf[:4], uint32(size))
	copy(buf[4:], p.bizBuf.Bytes())

	var b []byte
	b = larkProto.PackBizResp(buf)

	if _, err := p.transport.Write(b); err != nil {
		return thrift.NewTTransportExceptionFromError(err)
	}

	return thrift.NewTTransportExceptionFromError(p.transport.Flush())
}

func (p *LarkFramedTransportServer) readFrameHeader() (uint32, error) {
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

func (p *LarkFramedTransportServer) RemainingBytes() (num_bytes uint64) {
	return uint64(p.frameSize)
}

func (p *LarkFramedTransportServer) unpack() error {
	if p.unpacked {
		return nil
	}

	p.unpacked = true

	b, o, err := larkProto.UnPackReq(p.transport)
	if err != nil {
		return err
	}

	p.reqOpt = o

	p.bizBuf.Reset()
	p.bizBuf.Write(b)

	p.beginTime = time.Now()
	return nil
}

func (p *LarkFramedTransportServer) CostTime() time.Duration {
	return p.costTime
}
