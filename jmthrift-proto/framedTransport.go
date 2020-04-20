package jmthriftProto

import "git.apache.org/thrift.git/lib/go/thrift"

/*
在原生的 framed transport 上强制使用短连接，因为 lark 在收到非 lark 传输过来的协议的时候，总是短连接
*/

type FramedTransportFactory struct{}

// Return a wrapped instance of the base Transport.
func (p *FramedTransportFactory) GetTransport(trans thrift.TTransport) thrift.TTransport {
	return NewFramedTransport(trans)
}

func NewFramedTransportFactory() *FramedTransportFactory {
	return &FramedTransportFactory{}
}

type FramedTransport struct {
	*thrift.TFramedTransport
	afterFlush bool //在调用 flush 函数之后，设置为 true，表示一次调用完结，下次写之前需要重现建立连接
}

func NewFramedTransport(trans thrift.TTransport) thrift.TTransport {
	return &FramedTransport{
		TFramedTransport: thrift.NewTFramedTransport(trans),
	}
}

func (t *FramedTransport) Flush() error {
	err := t.TFramedTransport.Flush()
	if err != nil {
		return err
	}

	t.afterFlush = true

	return err
}

func (t *FramedTransport) Write(buf []byte) (int, error) {
	t.reOpen()
	return t.TFramedTransport.Write(buf)

}

func (t *FramedTransport) WriteByte(c byte) error {
	t.reOpen()
	return t.TFramedTransport.WriteByte(c)

}

func (t *FramedTransport) WriteString(s string) (int, error) {
	t.reOpen()
	return t.TFramedTransport.WriteString(s)

}

func (t *FramedTransport) reOpen() {
	if !t.afterFlush {
		return
	}
	t.TFramedTransport.Close()
	t.TFramedTransport.Open()
	t.afterFlush = false
}
