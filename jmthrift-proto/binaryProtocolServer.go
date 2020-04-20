package jmthriftProto

import (
	"encoding/json"
	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/pkg/errors"
	"reflect"
)

type BinaryProtocolServer struct {
	*thrift.TBinaryProtocol
	seqid      int32 //因为 PHPserver 在发生异常的时候， seqid 会返回 0，所以这里是为了适应这种情况
	trans      thrift.TTransport
	handler    interface{}
	methodName string
}

func NewBinaryProtocolServer(trans thrift.TTransport, handler interface{}) *BinaryProtocolServer {
	return &BinaryProtocolServer{
		handler:         handler,
		trans:           trans,
		TBinaryProtocol: thrift.NewTBinaryProtocolTransport(trans),
	}
}

func (p *BinaryProtocolServer) MethodName() string {
	return p.methodName
}

func (p *BinaryProtocolServer) WriteMessageBegin(name string, kind thrift.TMessageType, seqid int32) error {
	//提取 owl context
	var owlc = make(map[string]interface{})
	if p.handler != nil {
		if v := reflect.ValueOf(p.handler).Elem().FieldByName("Owlc"); v.IsValid() {
			reflect.ValueOf(&owlc).Elem().Set(v)
		}
	}

	//组装 context
	owlcBytes, _ := json.Marshal(owlc)
	context := map[string]string{
		"methodName":  name,
		"owl_context": string(owlcBytes),
	}

	//将 context thrift 序列化
	buf := thrift.NewTMemoryBuffer()
	prot := thrift.NewTBinaryProtocolTransport(buf)
	prot.WriteMessageBegin(__parrotContextPrefix, thrift.CALL, 0)
	prot.WriteMapBegin(thrift.STRING, thrift.STRING, len(context))
	for k, v := range context {
		prot.WriteString(k)
		prot.WriteString(v)
	}
	prot.WriteMapEnd()
	prot.WriteMessageEnd()

	//将 jmthrift 多的部分先写进去
	b := buf.Bytes()
	if _, err := p.trans.Write(b); err != nil {
		return errors.WithStack(err)
	}

	p.seqid = seqid

	//原生 thrift 调用
	if err := p.TBinaryProtocol.WriteMessageBegin(name, kind, seqid); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (p *BinaryProtocolServer) ReadMessageBegin() (string, thrift.TMessageType, int32, error) {
	name, _, _, err := p.TBinaryProtocol.ReadMessageBegin()
	if err != nil {
		return "", thrift.EXCEPTION, 0, errors.WithStack(err)
	}

	if name != __parrotContextPrefix {
		return "", thrift.EXCEPTION, 0, errors.Errorf("parrot context invalid. get: %s, want: %s", name, __parrotContextPrefix)

	}

	_, _, size, err := p.TBinaryProtocol.ReadMapBegin()
	if err != nil {
		return "", thrift.EXCEPTION, 0, errors.WithStack(err)
	}

	var context = make(map[string]string)
	for i := 0; i < size; i++ {
		k, err := p.TBinaryProtocol.ReadString()
		if err != nil {
			return "", thrift.EXCEPTION, 0, errors.WithStack(err)
		}

		v, err := p.TBinaryProtocol.ReadString()
		if err != nil {
			return "", thrift.EXCEPTION, 0, errors.WithStack(err)
		}

		context[k] = v
	}
	p.TBinaryProtocol.ReadMapEnd()
	p.TBinaryProtocol.ReadMessageEnd()

	//提取 owl context
	owlc, err := getOwlContext(context)
	if err != nil {
		return "", thrift.EXCEPTION, 0, errors.WithStack(err)
	}

	//设置 owl context
	if p.handler != nil {
		if v := reflect.ValueOf(p.handler).Elem().FieldByName("Owlc"); v.IsValid() {
			v.Set(reflect.ValueOf(owlc))
		}
	}

	p.methodName = context["methodName"]

	r1, r2, r3, err := p.TBinaryProtocol.ReadMessageBegin()
	if err != nil {
		return "", thrift.EXCEPTION, 0, errors.WithStack(err)
	}

	if r2 == thrift.EXCEPTION {
		error0 := thrift.NewTApplicationException(thrift.UNKNOWN_APPLICATION_EXCEPTION, "Unknown Exception")
		error1, err := error0.Read(p.TBinaryProtocol)
		if err != nil {
			return r1, r2, r3, err
		}
		if err = p.TBinaryProtocol.ReadMessageEnd(); err != nil {
			return r1, r2, r3, err
		}
		return r1, r2, r3, error1
	}

	return r1, r2, p.seqid, err
}

func getOwlContext(context map[string]string) (map[string]interface{}, error) {
	var owlc = make(map[string]interface{})
	if str, ok := context["owl_context"]; !ok {
		return owlc, nil
	} else {
		if err := json.Unmarshal([]byte(str), &owlc); err != nil {
			return nil, err
		}

		if owlc == nil {
			owlc = make(map[string]interface{})
		}

		return owlc, nil
	}

}
