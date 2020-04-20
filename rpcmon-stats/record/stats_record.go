package record

import (
	"github.com/zeast/logs"
	"path/filepath"
	"time"
)

// StatsRecord 收集数据，并定期刷盘，上报
type StatsRecord struct {
	stData          *stData
	detailData      *detailData
	userDetailData  *userDetailData
	failedLogData   *failedLogData
	qpsReport       *qpsReport
	failedQpsReport *failedQpsReport
	failedLogReport *failedLogReport
	messageCh       chan *StatsMessage
	closeCh         chan struct{}
}

type StatsMessage struct {
	ModuleName    string
	InterfaceName string
	SourceIP      string
	TargetIP      string
	Success       bool
	CostTime      time.Duration
	User          string
	Time          time.Time
	Code          int
	Msg           string
}

func NewStatsRecord() *StatsRecord {
	var sc = &StatsRecord{
		stData:          newStData(),
		detailData:      newDetailData(),
		userDetailData:  newUserDetailData(),
		failedLogData:   newFailedLogData(),
		qpsReport:       newQpsReport(),
		failedQpsReport: newFailedQpsReport(),
		failedLogReport: newFaildLogReport(),
		messageCh:       make(chan *StatsMessage, 100),
		closeCh:         make(chan struct{}),
	}

	return sc
}

func (s *StatsRecord) Start() {
	go s.stData.start()
	go s.detailData.start()
	go s.userDetailData.start()
	go s.failedLogData.start()

	go s.qpsReport.start()
	go s.failedQpsReport.start()
	go s.failedLogReport.start()

	go s.handle()
}

func (s *StatsRecord) Stop() {
	close(s.closeCh)
}

func (s *StatsRecord) Write(m *StatsMessage) {
	select {
	case s.messageCh <- m:
	default:
		logs.Errorf("stats collect. busy, drop rpcmonMessage %v", m)
	}
}

func (s *StatsRecord) handle() {
	for {
		select {
		case m := <-s.messageCh:
			s.stData.add(m)
			s.detailData.add(m)
			s.userDetailData.add(m)
			s.failedLogData.add(m)

			s.failedQpsReport.add(m)
			s.qpsReport.add(m)
			s.failedLogReport.add(m)

		case <-s.closeCh:
			return
		}
	}
}

func (s *StatsRecord) WithProjectName(name string) {
	s.qpsReport.withProjectName(name)
	s.failedQpsReport.withProjectName(name)
	s.failedLogReport.withProjectName(name)
}

func (s *StatsRecord) WithPrefixPath(p string) {
	s.stData.withFilePath(filepath.Join(p, "st"))
	s.detailData.withFilePath(filepath.Join(p, "detail"))
	s.userDetailData.withFilePath(filepath.Join(p, "userdetail"))
	s.failedLogData.withFilePath(filepath.Join(p, "log"))
}
