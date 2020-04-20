package jmthriftProto

import (
	"git.apache.org/thrift.git/lib/go/thrift"
)

type TProcessorFactory interface {
	GetProcessor(trans thrift.TTransport) thrift.TProcessor
}

type tProcessorFactory struct {
	proFunc func() thrift.TProcessor
}

func NewTProcessorFactory(proFunc func() thrift.TProcessor) TProcessorFactory {
	return &tProcessorFactory{proFunc: proFunc}
}

func (p *tProcessorFactory) GetProcessor(trans thrift.TTransport) thrift.TProcessor {
	return p.proFunc()
}
