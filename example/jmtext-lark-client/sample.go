package main

import (
	"fmt"
	jmtextLarkClient "gitlab.int.jumei.com/JMArch/go-rpc/jmtext-lark-client"
)

func main() {
	clientHelloWorld := jmtextLarkClient.NewClient().WithServiceName("hello-world").WithUser("test-user").WithClass("compute").MustValid()

	for i := 0; i < 10; i++ {
		b, o, err := clientHelloWorld.Call("sum", map[string]interface{}{"app_name": "hello_api"}, 1, 2, 3)
		fmt.Println(string(b), o, err)
	}
}
