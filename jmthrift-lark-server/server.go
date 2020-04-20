package jmthriftLarkServer

import (
	"errors"
	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/zeast/logs"
	jmthriftProto "gitlab.int.jumei.com/JMArch/go-rpc/jmthrift-proto"
	"gitlab.int.jumei.com/JMArch/go-rpc/lark-proto"
	"gitlab.int.jumei.com/JMArch/go-rpc/lark-registry"
	rpcmonStats "gitlab.int.jumei.com/JMArch/go-rpc/rpcmon-stats"
	"gitlab.int.jumei.com/JMArch/go-rpc/rpcmon-stats/record"
	"log"
	"net"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"sync"
	"time"
)

type Server struct {
	serviceName     string //提供给调用方用的服务名称
	servicePort     int    //服务的端口
	serviceIP       string //服务的 IP，一般来说是本机 IP，内网 IP
	rproxyNetwork   string //用来接收来自 lark 的数据
	rproxyNetaddr   string
	reg             *larkRegistry.Registry
	processor       thrift.TProcessor
	proFunc         func() (thrift.TProcessor, interface{})
	quit            chan struct{}
	serverTransport thrift.TServerTransport
	rpcmonStats     *rpcmonStats.Stats
	rpcmonEnable    bool
	logger          *logs.Logger
}

func NewServer() *Server {
	s := &Server{
		serviceIP:    localIPByHostsFile(),
		quit:         make(chan struct{}),
		rpcmonEnable: true,
	}

	l := logs.NewLogger(os.Stdout)
	l.SetLogLevel(logs.LevelError)
	s.logger = l

	return s
}

func (s *Server) valid() (*Server, error) {
	if s.serviceName == "" {
		return nil, errors.New("must have a service name")
	}

	if s.serviceIP == "" {
		return nil, errors.New("must have a service ip")
	}

	if s.servicePort == 0 {
		return nil, errors.New("must have a service port")
	}

	if s.rproxyNetwork == "" && s.rproxyNetaddr == "" {
		s.rproxyNetwork = "tcp4"
		s.rproxyNetaddr = "127.0.0.1:" + strconv.Itoa(s.servicePort+1)
	}

	if s.rpcmonEnable && s.rpcmonStats == nil {
		s.rpcmonStats = rpcmonStats.NewStats()
		s.rpcmonStats.Start()
	}

	if s.reg == nil {
		s.reg = larkRegistry.NewRegistry(larkProto.ServiceRegistInfo{
			Ip:          s.serviceIP,
			Port:        s.servicePort,
			ServiceName: s.serviceName,
			WorkNum:     runtime.GOMAXPROCS(-1) * 100,
			CgiPass:     s.rproxyNetwork + "://" + s.rproxyNetaddr,
		})
	}

	return s, nil
}

func (s *Server) WithLarkRegistry(reg *larkRegistry.Registry) *Server {
	s.reg = reg
	return s
}

func (s *Server) WithServicePort(port int) *Server {
	s.servicePort = port
	return s
}

func (s *Server) WithServiceIP(ip string) *Server {
	s.serviceIP = ip
	return s
}

func (s *Server) WithServiceName(name string) *Server {
	s.serviceName = name
	return s
}

func (s *Server) WithLogger(l *logs.Logger) *Server {
	if l != nil {
		s.logger = l
	}

	return s
}

func (s *Server) WithProcessor(processor thrift.TProcessor) *Server {
	s.processor = processor
	return s
}

func (s *Server) WithNewProcessorFunc(proFunc func() (thrift.TProcessor, interface{})) *Server {
	s.proFunc = proFunc
	return s
}

func (s *Server) WithRpcmonStats(stats *rpcmonStats.Stats) *Server {
	s.rpcmonStats = stats
	return s
}

func (s *Server) WithRpcmonEnable(enable bool) *Server {
	s.rpcmonEnable = enable
	return s
}

func (s *Server) Serve() error {
	if _, err := s.valid(); err != nil {
		return err
	}

	serverSocket, err := thrift.NewTServerSocket(s.rproxyNetaddr)
	if err != nil {
		return err
	}

	s.serverTransport = serverSocket

	go func() {
		s.logger.Errorf("%s", s.serve())
	}()

	return nil
}

func localIPByHostsFile() string {
	name, err := os.Hostname()
	if err != nil {
		return ""
	}

	addr, err := net.LookupHost(name)
	if err != nil {
		return ""
	}

	if len(addr) == 0 {
		return ""
	}

	return addr[0]
}

func (s *Server) listen() error {
	return s.serverTransport.Listen()
}

func (s *Server) acceptLoop() error {
	for {
		client, err := s.serverTransport.Accept()
		if err != nil {
			select {
			case <-s.quit:
				return nil
			default:
			}
			return err
		}
		if client != nil {
			go func() {
				if err := s.processRequests(client); err != nil {
					log.Println("error processing request:", err)
				}
			}()
		}
	}
}

func (s *Server) serve() error {
	err := s.listen()
	if err != nil {
		return err
	}
	s.acceptLoop()
	return nil
}

var stopOnce sync.Once

func (s *Server) Stop() error {
	q := func() {
		s.quit <- struct{}{}
		s.serverTransport.Interrupt()
	}
	stopOnce.Do(q)
	return nil
}

func (s *Server) LarkRegistry() *larkRegistry.Registry {
	return s.reg
}

func (s *Server) processRequests(client thrift.TTransport) error {
	trans := jmthriftProto.NewLarkFramedTransportServer().WithTransport(client)
	processor, handler := s.proFunc()
	prot := jmthriftProto.NewBinaryProtocolServer(trans, handler)

	defer func() {
		if e := recover(); e != nil {
			log.Printf("panic in processor: %s: %s", e, string(debug.Stack()))
		}
	}()

	if trans != nil {
		defer trans.Close()
	}

	for {
		ok, err := processor.Process(prot, prot)
		if err, ok := err.(thrift.TTransportException); ok && err.TypeId() == thrift.END_OF_FILE {
			return nil
		} else if err != nil {
			log.Printf("error processing request: %s", err)
			return err
		}
		if err, ok := err.(thrift.TApplicationException); ok && err.TypeId() == thrift.UNKNOWN_METHOD {
			continue
		}
		if !ok {
			break
		}

		if s.rpcmonStats != nil {
			m := &record.StatsMessage{
				ModuleName:    s.serviceName,
				InterfaceName: prot.MethodName(),
				SourceIP:      trans.ReqOption().ClientIP,
				TargetIP:      s.serviceIP,
				Success:       true,
				CostTime:      trans.CostTime(),
				User:          "go-rpc",
				Time:          time.Now(),
				Code:          0,
				Msg:           "",
			}

			if e, ok := err.(*rpcmonStats.RpcmonErr); ok {
				m.Success = false
				m.Msg = e.Msg
				m.Code = e.Code
			}

			s.rpcmonStats.Write(m)
		}
	}
	return nil
}
