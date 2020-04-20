package jmthriftProto

import (
	"encoding/json"
	"git.apache.org/thrift.git/lib/go/thrift"
)

var __parrotContextPrefix = "$$$_1_PC$$$$$$$$__1_&_8_#"

func packContext(context map[string]string, owlc map[string]interface{}) []byte {
	buf := thrift.NewTMemoryBuffer()
	prot := thrift.NewTBinaryProtocolTransport(buf)
	prot.WriteMessageBegin(__parrotContextPrefix, thrift.CALL, 0)
	prot.WriteMapBegin(thrift.STRING, thrift.STRING, len(context)+1)
	for k, v := range context {
		prot.WriteString(k)
		prot.WriteString(v)
	}

	owlcBytes, _ := json.Marshal(owlc)
	prot.WriteString("owl_context")
	prot.WriteString(string(owlcBytes))

	prot.WriteMapEnd()
	prot.WriteMessageEnd()

	return buf.Bytes()
}
