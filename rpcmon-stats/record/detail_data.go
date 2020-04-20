package record

import (
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"
)

type detailData struct {
	sync.Mutex
	data                 map[string]map[string]map[string][2]int // ModuleName=>InterfaceName=>['ip'=>[suc_count, fail_count],'ip'=>[suc_count, fail_count], ...]
	filePath             string
	expTime              time.Duration //文件的过期时间
	dataFlushInterval    int           //秒
	defaultModuleName    string
	defaultInterfaceName string
}

func newDetailData() *detailData {
	return &detailData{
		filePath:             "/home/logs/go-rpc-server/stats/detail/",
		expTime:              time.Hour * 24 * 14,
		dataFlushInterval:    60,
		defaultModuleName:    defaultModuleName,
		defaultInterfaceName: defaultInterfaceName,
	}
}

func (s *detailData) withFilePath(p string) *detailData {
	s.filePath = p
	return s
}

func (s *detailData) start() {
	now := time.Now()

	d := now.Second() % s.dataFlushInterval
	if d == 0 {
		d = s.dataFlushInterval
	}
	flushTicker := time.NewTicker(time.Second * time.Duration(s.dataFlushInterval-d))
	firstFlush := true

	clearFileTicker := time.NewTicker(time.Hour)
	for {
		select {
		case <-clearFileTicker.C:
			clearFile(s.filePath, s.expTime)

		case n := <-flushTicker.C:
			s.writeToFile(n)

			if firstFlush {
				flushTicker.Stop()
				flushTicker = time.NewTicker(time.Second * time.Duration(s.dataFlushInterval))
				firstFlush = false
			}
		}
	}
}

func (s *detailData) add(m *StatsMessage) {
	if s.data == nil {
		s.data = make(map[string]map[string]map[string][2]int)
	}

	if s.data[m.ModuleName] == nil {
		s.data[m.ModuleName] = make(map[string]map[string][2]int)
	}

	if s.data[m.ModuleName][m.InterfaceName] == nil {
		s.data[m.ModuleName][m.InterfaceName] = make(map[string][2]int)
	}

	if s.data[s.defaultModuleName] == nil {
		s.data[s.defaultModuleName] = make(map[string]map[string][2]int)
	}

	if s.data[s.defaultModuleName][s.defaultInterfaceName] == nil {
		s.data[s.defaultModuleName][s.defaultInterfaceName] = make(map[string][2]int)
	}

	r := s.data[m.ModuleName][m.InterfaceName][m.SourceIP]
	d := s.data[s.defaultModuleName][s.defaultInterfaceName][m.SourceIP]
	if m.Success {
		r[0]++
		d[0]++
	} else {
		r[1]++
		d[1]++
	}
	s.data[m.ModuleName][m.InterfaceName][m.SourceIP] = r
	s.data[s.defaultModuleName][s.defaultInterfaceName][m.SourceIP] = d
}

func (s *detailData) writeToFile(wt time.Time) []error {
	var errs []error

	for moduleName, v := range s.data {
		for interfaceName, d := range v {
			if e := s.writeToOneFile(moduleName, interfaceName, d, wt); e != nil {
				errs = append(errs, e)
			}
		}
	}

	s.data = nil

	return errs
}

func (s *detailData) writeToOneFile(moduleName string, interfaceName string, d map[string][2]int, wt time.Time) error {
	text := fmt.Sprintf("%d\t%s\n", wt.Unix(), mustJsonMarshal(d))
	path := fmt.Sprintf("%s/%s", s.filePath, moduleName)
	name := fmt.Sprintf("%s-detail|%s", interfaceName, wt.Format("2006-01-02"))

	syscall.Umask(0)
	os.MkdirAll(path, 0777)

	file, err := os.OpenFile(fmt.Sprintf("%s/%s", path, name), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(text)
	return err
}
