package main

import (
	"fmt"
	"git.apache.org/thrift.git/lib/go/thrift"
	"gitlab.int.jumei.com/JMArch/go-rpc/example/trusteeship"
	"gitlab.int.jumei.com/JMArch/go-rpc/jmthrift-proto"
)

func main() {
	transport, err := thrift.NewTSocket("127.0.0.1:9898")
	if err != nil {
		panic(err)
	}
	defer transport.Close()

	err = transport.Open()
	if err != nil {
		panic(err)
	}

	trans := jmthriftProto.NewFramedTransport(transport)
	prot := jmthriftProto.NewBinaryProtocol(jmthriftProto.NewOwlContext(), trans)

	client := trusteeship.NewTrusteeshipDataClientProtocol(trans, prot, prot)
	r, err := client.IsExist(123456)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(r)

	r, err = client.GetDecryptPhoneNumber(123456)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(r)
}
