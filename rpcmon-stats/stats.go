package rpcmonStats

import (
	"encoding/json"
	"gitlab.int.jumei.com/JMArch/go-rpc/rpcmon-stats/record"
	"gitlab.int.jumei.com/JMArch/go-rpc/rpcmon-stats/server"
)

type RpcmonErr struct {
	Code int
	Msg  string
}

func (e *RpcmonErr) Error() string {
	v, _ := json.Marshal(e)
	return string(v)
}

type Stats struct {
	record *record.StatsRecord
	server *server.StatsServer
}

func NewStats() *Stats {
	c := &Stats{
		record: record.NewStatsRecord(),
		server: server.NewStatsServer(),
	}

	return c
}

func (s *Stats) Start() error {
	s.record.Start()
	return s.server.Start()
}

func (s *Stats) Stop() {
	s.record.Stop()
	s.server.Stop()
}

func (s *Stats) Write(m *record.StatsMessage) {
	s.record.Write(m)
}

func (s *Stats) WithProjectName(name string) *Stats {
	s.record.WithProjectName(name)
	return s
}

func (s *Stats) WithRecordPrefixPath(p string) *Stats {
	s.record.WithPrefixPath(p)
	s.server.WithPrefixPath(p)
	return s
}
