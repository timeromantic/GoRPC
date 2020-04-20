package main

import (
	"encoding/json"
	"fmt"
	"gitlab.int.jumei.com/JMArch/go-rpc/jmtext-lark-server"
	"gitlab.int.jumei.com/JMArch/go-rpc/lark-proto"
	"github.com/golang/protobuf/proto"
	"log"
	"os"
	"os/signal"
	"syscall"
	"./Proto"
)

func main() {
	// 声明一个名为 hello-world 的服务，服务端口为 9394
	s := jmtextLarkServer.NewServer().
		WithServiceName("hello-world").
		WithServicePort(9394)

	// 注册一个 class name  为 compute
	// method name 为 sum 的函数
	s.RegisterFunc("compute", "sum", sum)

	// 开始监听设定的socket
	err := s.Start()
	if err != nil {
		fmt.Println(err)
		return
	}

	// 注册服务
	s.LarkRegistry().Register([]string{"dev-cd"})

	// 此处程序可以 catch 一些系统信号
	// 让程序 hold 在这个地方
	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	<-sc

	// 在程序关闭的时候，别忘了调用 Unregister
	s.LarkRegistry().UnRegister([]string{"dev-cd"})
}

// sum 实现了多个数字相加的功能，并将结果返回
func sum(b []interface{}, o *larkProto.ReqOption, owlc map[string]interface{}) (interface{}, error) {
	fmt.Printf("收到的 owl context:%+v \n", owlc)
	var total int

	for _, n := range b {
		total += int(n.(float64))
	}

	return total, nil
}


// 接收buf数据
func decodeResponse(buf []byte, o *larkProto.ReqOption, owlc map[string]interface{}) (interface{}, error)  {
	// 对已经序列化的数据进行反序列化
	var Response Proto.RTAResponse_2_0
	// 对流数据进行反序列
	err := proto.Unmarshal(buf, &Response)
	if err != nil{
		log.Fatalln("UnMashal data error:", err)
	}
	// 编码json返回数据
	jsons, _:= json.Marshal(Response)
	fmt.Println(string(jsons))
}
