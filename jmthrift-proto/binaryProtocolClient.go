package jmthriftProto

import (
	"encoding/json"
	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/pkg/errors"
)

type BinaryProtocolClient struct {
	*thrift.TBinaryProtocol
	seqid      int32 //因为 PHPserver 在发生异常的时候， seqid 会返回 0，所以这里是为了适应这种情况
	context    map[string]string
	owlc       map[string]interface{}
	owlcSetted bool
	trans      thrift.TTransport
}

func NewBinaryProtocolClient(trans thrift.TTransport) *BinaryProtocolClient {
	return &BinaryProtocolClient{
		trans:           trans,
		TBinaryProtocol: thrift.NewTBinaryProtocolTransport(trans),
	}
}

func (p *BinaryProtocolClient) GetOwlContext() (map[string]interface{}, error) {
	if p.context == nil {
		return map[string]interface{}{}, nil
	}

	str := p.context["owl_context"]
	if len(str) == 0 {
		return map[string]interface{}{}, nil
	}

	var owlc = make(map[string]interface{})
	err := json.Unmarshal([]byte(str), &owlc)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return owlc, nil
}

func (p *BinaryProtocolClient) SetOwlContext(owlc map[string]interface{}) {
	if owlc == nil {
		return
	}

	p.owlc = owlc
	p.owlcSetted = true
	return
}

func (p *BinaryProtocolClient) WriteMessageBegin(name string, kind thrift.TMessageType, seqid int32) error {
	if !p.owlcSetted {
		p.owlc = make(map[string]interface{})
	}
	p.owlcSetted = false

	b := packContext(map[string]string{"methodName": name}, p.owlc)
	if _, err := p.trans.Write(b); err != nil {
		return errors.WithStack(err)
	}

	p.seqid = seqid
	if err := p.TBinaryProtocol.WriteMessageBegin(name, kind, seqid); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (p *BinaryProtocolClient) ReadMessageBegin() (string, thrift.TMessageType, int32, error) {
	p.context = make(map[string]string)
	p.owlcSetted = false

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

	for i := 0; i < size; i++ {
		k, err := p.TBinaryProtocol.ReadString()
		if err != nil {
			return "", thrift.EXCEPTION, 0, errors.WithStack(err)
		}

		v, err := p.TBinaryProtocol.ReadString()
		if err != nil {
			return "", thrift.EXCEPTION, 0, errors.WithStack(err)
		}

		p.context[k] = v
	}
	p.TBinaryProtocol.ReadMapEnd()
	p.TBinaryProtocol.ReadMessageEnd()

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
