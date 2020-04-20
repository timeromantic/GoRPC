package jmtextLarkServer

import (
	"errors"
	"fmt"
	"github.com/zeast/logs"
	jmtextProto "gitlab.int.jumei.com/JMArch/go-rpc/jmtext-proto"
	"gitlab.int.jumei.com/JMArch/go-rpc/lark-proto"
	"gitlab.int.jumei.com/JMArch/go-rpc/lark-registry"
	rpcmonStats "gitlab.int.jumei.com/JMArch/go-rpc/rpcmon-stats"
	"gitlab.int.jumei.com/JMArch/go-rpc/rpcmon-stats/record"
	"net"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"time"
)

type HandlerFunc func([]interface{}, *larkProto.ReqOption, map[string]interface{}) (interface{}, error)

type Server struct {
	serviceName   string //lark 暴露到外部的信息
	servicePort   int
	serviceIP     string
	rproxyNetwork string //用来接收来自 lark 的数据
	rproxyNetaddr string
	handler       map[string]map[string]HandlerFunc //class => method => func
	reuseConn     bool                              //true 的话是使用长连接
	reg           *larkRegistry.Registry
	logger        *logs.Logger
	rpcmonStats   *rpcmonStats.Stats
	rpcmonEnable  bool
}

func NewServer() *Server {
	s := &Server{
		reuseConn:    true,
		serviceIP:    localIPByHostsFile(),
		handler:      make(map[string]map[string]HandlerFunc),
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

func (s *Server) WithRpcmonStats(stats *rpcmonStats.Stats) *Server {
	s.rpcmonStats = stats
	return s
}

func (s *Server) WithRpcmonEnable(enable bool) *Server {
	s.rpcmonEnable = enable
	return s
}

func (s *Server) WithLogger(l *logs.Logger) *Server {
	if l != nil {
		s.logger = l
	}
	return s
}

func (s *Server) RegisterFunc(class string, method string, fn HandlerFunc) {
	if s.handler[class] == nil {
		s.handler[class] = make(map[string]HandlerFunc)
	}

	s.handler[class][method] = fn
}

func (s *Server) LarkRegistry() *larkRegistry.Registry {
	return s.reg
}

func (s *Server) Start() error {
	if _, err := s.valid(); err != nil {
		return err
	}

	l, err := net.Listen(s.rproxyNetwork, s.rproxyNetaddr)
	if err != nil {
		return err
	}

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				if nerr, ok := err.(net.Error); ok {
					if nerr.Temporary() {
						s.logger.Warnf("listen %s %s temporary error. %s", s.rproxyNetwork, s.rproxyNetaddr, nerr.Error())
					}
				} else {
					s.logger.Errorf("listen %s %s error. %s", s.rproxyNetwork, s.rproxyNetaddr, err.Error())
				}
				time.Sleep(time.Millisecond)
			}

			go s.handleConn(conn)
		}
	}()

	return nil
}

func (s *Server) handleConn(conn net.Conn) {
	defer func() {
		conn.Close()
	}()

	for {
		err := s.handleConnOnce(conn)
		if err != nil {
			s.logger.Errorf("%s", err.Error())

			ex := jmtextProto.Exception{
				ServerIP:      localIPByHostsFile(),
				TraceAsString: err.Error(),
			}

			b, err := jmtextProto.MarshalResp(ex)
			if err != nil {
				s.logger.Errorf("%s", err.Error())
				break
			}

			b = larkProto.PackBizResp(b)
			conn.Write(b)
			break
		}

		if !s.reuseConn {
			break
		}
	}

}

func (s *Server) handleConnOnce(conn net.Conn) (err error) {
	defer func() {
		if x := recover(); x != nil {
			err = fmt.Errorf("recover panic:%s; %s", x, string(debug.Stack()))
		}
	}()

	//解 lark 包
	req, o, err := larkProto.UnPackReq(conn)
	if err != nil {
		err = fmt.Errorf("lark proto unpack req error.%w", err)
		return
	}

	begin := time.Now()
	var moduleName string
	var interfaceName string
	defer func() {
		if s.rpcmonStats != nil {
			m := &record.StatsMessage{
				ModuleName:    moduleName,
				InterfaceName: interfaceName,
				SourceIP:      o.ClientIP,
				User:          "",
				TargetIP:      localIPByHostsFile(),
				CostTime:      time.Since(begin),
			}

			if err != nil {
				m.Success = false

				if rerr, ok := err.(*rpcmonStats.RpcmonErr); ok {
					m.Code = rerr.Code
					m.Msg = rerr.Msg
				} else {
					m.Code = 1000
					m.Msg = err.Error()
				}
			} else {
				m.Success = true
			}

			s.rpcmonStats.Write(m)
		}
	}()

	//解 jmtext 包
	class, method, params, owlContext, err := jmtextProto.UnmarshalReq(req)
	if err != nil {
		return fmt.Errorf("jmtext proto unmarshal packet error. %w", err)
	}

	moduleName = class
	interfaceName = method

	//查找处理函数
	fn := s.handlerFn(class, method)
	if fn == nil {
		return fmt.Errorf("can not find handler, class:%s, method:%s", class, method)
	}

	//调用处理函数
	resp, err := fn(params, o, owlContext)
	if resp == nil {
		//不需要返回数据，不代表出错
		return nil
	}

	//打包 jmtext
	b, err := jmtextProto.MarshalResp(resp)
	if err != nil {
		return fmt.Errorf("jmtext proto marhsal resp error.%w", err)
	}

	//打包 lark
	b = larkProto.PackBizResp(b)

	//写入返回的数据
	_, err = conn.Write(b)
	if err != nil {
		return fmt.Errorf("lark proto pack biz error.%w", err)
	}

	return nil
}

func (s *Server) handlerFn(class, method string) HandlerFunc {
	_, ok := s.handler[class]
	if !ok {
		return nil
	}

	fn, ok := s.handler[class][method]
	if !ok {
		return nil
	}
	return fn
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
