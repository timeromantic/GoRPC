package server

import (
	"github.com/pkg/errors"
	"github.com/zeast/logs"
	"io"
	"net"
	"path/filepath"
	"sync/atomic"
	"time"
)

const (
	cmdReportIP       = 11
	cmdReportIPResult = 12
	cmdProvider       = 107

	subCmdGetSt                = 201
	subCmdGetModules           = 202
	subCmdGetLogs              = 203
	subCmdGetStAndModules      = 204
	subCmdGetClientSummary     = 205
	subCmdGetClientDetail      = 206
	subCmdGetClientRaw         = 207
	subCmdGetClientUserSummary = 208
	subCmdGetClientUserDetail  = 209
	subCmdGetClientUserRaw     = 210
)

// StatsServer 提供 一个 udp 接口，用来接收来自 rpcmon 中心节点的命令，根据命令返回对应的数据
type StatsServer struct {
	l                  net.Listener
	netAddr            string
	failedLogDataPath  string
	stDataPath         string
	detailDataPath     string
	userDetailDataPath string
	closed             int32 //0 表示关闭，1 表示启动
}

func NewStatsServer() *StatsServer {
	return &StatsServer{
		failedLogDataPath:  "/home/logs/go-rpc-server/stats/log/",
		stDataPath:         "/home/logs/go-rpc-server/stats/st/",
		detailDataPath:     "/home/logs/go-rpc-server/stats/detail/",
		userDetailDataPath: "/home/logs/go-rpc-server/stats/userdetail",
		netAddr:            ":20204",
	}
}

func (s *StatsServer) WithPrefixPath(p string) *StatsServer {
	s.failedLogDataPath = filepath.Join(p, "log")
	s.stDataPath = filepath.Join(p, "st")
	s.detailDataPath = filepath.Join(p, "detail")
	s.userDetailDataPath = filepath.Join(p, "userdetail")
	return s
}

func (s *StatsServer) Start() error {
	if !atomic.CompareAndSwapInt32(&s.closed, 0, 1) {
		return nil
	}

	l, err := net.Listen("tcp4", s.netAddr)
	if err != nil {
		return errors.WithStack(err)
	}

	s.l = l

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				logs.Errorf("accept error, %s", err)
				time.Sleep(time.Millisecond * 100)
				continue
			}

			conn.SetDeadline(time.Now().Add(time.Second * 3))
			ret, err := s.handle(conn)
			if err != nil {
				logs.Errorf("stats server handle done. ret:%q, err:%s", string(ret), err)
			} else {
				logs.Debugf("stats server handle done. ret:%q, err:%s", string(ret), err)
			}

			if ret != nil {
				conn.Write(ret)
			}

			conn.Close()
		}
	}()

	return nil
}

func (s *StatsServer) Stop() {
	if atomic.CompareAndSwapInt32(&s.closed, 1, 0) {
		s.l.Close()
	}
}

func (s *StatsServer) handle(r io.Reader) ([]byte, error) {
	defer PrintPanicStack()

	msg := new(rpcmonMessage)
	if err := msg.fill(r); err != nil {
		return nil, errors.WithStack(err)
	}

	logs.Debugf("stats server handle. cmd:%d, subCmd:%d, body:%s", msg.cmd, msg.subCmd, string(msg.body))

	var ret []byte
	var err error
	switch msg.cmd {
	case cmdReportIP:
		err = errors.Errorf("can not handle this subcmd:%d", msg.subCmd)

	case cmdProvider:
		switch msg.subCmd {
		case subCmdGetStAndModules:
			ret, err = s.subCmdGetStAndModulesHandler(msg.body)

		case subCmdGetLogs:
			ret, err = s.subCmdGetLogsHandler(msg.body)

		case subCmdGetClientSummary:
			ret, err = s.subCmdGetClientSummaryHandler(msg.body)

		case subCmdGetClientUserSummary:
			ret, err = s.subCmdGetClientUserSummaryHandler(msg.body)

		case subCmdGetClientDetail:
			ret, err = s.subCmdGetClientDetailHandler(msg.body)

		case subCmdGetClientUserDetail:
			ret, err = s.subCmdGetClientUserDetailHandler(msg.body)

		case subCmdGetClientRaw:
			ret, err = s.subCmdGetClientRawHandler(msg.body)

		case subCmdGetClientUserRaw:
			ret, err = s.subCmdGetClientUserRawHandler(msg.body)

		default:
			err = errors.Errorf("can not handle this subcmd:%d", msg.subCmd)
		}

	default:
		err = errors.Errorf("can not handle this cmd:%d", msg.cmd)
	}

	if err != nil {
		return nil, err
	}

	msg.body = ret

	return msg.marshal()
}
