package main

import (
	"fmt"
	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/zeast/logs"
	"gitlab.int.jumei.com/JMArch/go-rpc/example/trusteeship"
	jmthriftLarkServer "gitlab.int.jumei.com/JMArch/go-rpc/jmthrift-lark-server"
	rpcmonStats "gitlab.int.jumei.com/JMArch/go-rpc/rpcmon-stats"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
)

type TrusteeshipServer struct {
	Owlc map[string]interface{} //如果想使用 owlc context, 必须添加这个成员变量，不想使用的话，则可以不加
}

func (s *TrusteeshipServer) IsExist(data float64) (r map[string]string, err error) {
	fmt.Printf("owl context in func isExist %v \n", s.Owlc) //owlc 会被设置成本次 client 请求过来的 owl context
	s.Owlc["hello"] = "world"                               //业务可以修改自己的 owl context, 在函数返回的时候，context 会回传给 client
	return map[string]string{
		"a": "b",
	}, nil
}

func (s *TrusteeshipServer) EncryptData(data string) (r map[string]string, err error) {
	return map[string]string{
		"a1": "b1",
	}, nil
}
func (s *TrusteeshipServer) GetDecryptData(dataId float64, appId string, timestamp float64, token string) (r map[string]string, err error) {
	return map[string]string{
		"a2": "b2",
	}, nil
}
func (s *TrusteeshipServer) EncryptDataBatchSimple(dataArr []string) (r map[string]string, err error) {
	return map[string]string{
		"a3": "b3",
	}, nil
}

func (s *TrusteeshipServer) GetDecryptDataBatchSimple(dataIdArr []float64, appId string, timestamp float64, token string) (r map[string]string, err error) {
	return map[string]string{
		"a4": "b4",
	}, nil
}

func (s *TrusteeshipServer) GetDecryptPhoneNumber(dataId float64) (r map[string]string, err error) {
	fmt.Printf("owl context in func GetDecryptPhoneNumber %v \n", s.Owlc) //owlc 会被设置成本次 client 请求过来的 owl context
	return map[string]string{
			"a5": "b5",
		}, &rpcmonStats.RpcmonErr{ //返回此类型的错误，如果开启 rpcmon 相关功能，会被 rpcmon 搜集
			Code: 5555,
			Msg:  "get decrypt phone number error",
		}
}

func main() {
	go func() {
		fmt.Println("pprof")
		logs.Fatal(http.ListenAndServe(":16666", nil))
	}()

	f := func() (thrift.TProcessor, interface{}) {
		handler := new(TrusteeshipServer)
		processor := trusteeship.NewTrusteeshipDataProcessor(handler)
		return processor, handler
	}

	stats := rpcmonStats.NewStats()
	stats.WithProjectName("TrusteeshipTest")
	err := stats.Start()
	if err != nil {
		panic(err)
	}

	s := jmthriftLarkServer.NewServer().
		WithServiceName("TrusteeshipTest").
		WithServicePort(9898).
		WithNewProcessorFunc(f).
		WithRpcmonStats(stats)

	err = s.Serve()
	if err != nil {
		panic(err)
	}

	fmt.Println(s.LarkRegistry().Register(nil))

	fmt.Printf("当前注册了的注册中心信息: %+v \n", s.LarkRegistry().RegistryCenter())
	fmt.Printf("所有可以注册的中心信息: %+v \n", s.LarkRegistry().AllRegistryCenter())

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
	s.LarkRegistry().UnRegister(nil)

}
