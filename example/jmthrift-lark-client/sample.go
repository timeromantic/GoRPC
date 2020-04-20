package main

import (
	"fmt"
	"gitlab.int.jumei.com/JMArch/go-rpc/example/trusteeship"
	jmthriftProto "gitlab.int.jumei.com/JMArch/go-rpc/jmthrift-proto"
	larkProto "gitlab.int.jumei.com/JMArch/go-rpc/lark-proto"
	"time"
)

func main() {
	trans := jmthriftProto.NewLarkFramedTransportClient().WithLarkReqOption(&larkProto.LarkReqOption{
		TargetService: "TrusteeshipTest",
	}).MustValid()
	prot := jmthriftProto.NewBinaryProtocolClient(trans)

	client := trusteeship.NewTrusteeshipDataClientProtocol(trans, prot, prot)

	for {

		r, err := client.IsExist(123456)
		if err != nil {
			fmt.Println(err)
			//return
		}
		fmt.Println(r)
		fmt.Println(prot.GetOwlContext())

		//设置自定义的 owl context
		prot.SetOwlContext(map[string]interface{}{"k1": "v1"})
		r, err = client.GetDecryptPhoneNumber(123456)
		if err != nil {
			fmt.Printf("error:%s, remote_ip:%s", err, trans.RespOption().ServerIP)
			//return
		}

		fmt.Println(r)
		fmt.Println(prot.GetOwlContext())
		time.Sleep(time.Second)
	}
}
