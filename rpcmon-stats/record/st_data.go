package record

import (
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"
)

type stData struct {
	sync.Mutex
	data                 map[string]map[string]*OneRecord
	filePath             string
	expTime              time.Duration //文件的过期时间
	dataFlushInterval    int           //秒
	defaultModuleName    string
	defaultInterfaceName string
}

type OneRecord struct {
	code         map[int]int //对应错误码的次数
	sucCount     int         //成功的请求次数
	sucCostTime  float64     //成功的请求，耗时时间，单位秒
	failCount    int         //失败的请求次数
	failCostTime float64     //失败的请求，耗时时间，单位秒
	time         time.Time
}

func newStData() *stData {
	s := &stData{
		filePath:             "/home/logs/go-rpc-server/stats/st/",
		expTime:              time.Hour * 24 * 14,
		dataFlushInterval:    60,
		defaultModuleName:    defaultModuleName,
		defaultInterfaceName: defaultInterfaceName,
	}

	return s
}

func (s *stData) withFilePath(p string) *stData {
	s.filePath = p
	return s
}

func (s *stData) start() {
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

func (s *stData) add(m *StatsMessage) {
	s.Lock()
	defer s.Unlock()

	if s.data == nil {
		s.data = make(map[string]map[string]*OneRecord)
	}

	if s.data[m.ModuleName] == nil {
		s.data[m.ModuleName] = make(map[string]*OneRecord)
	}

	if s.data[m.ModuleName][m.InterfaceName] == nil {
		s.data[m.ModuleName][m.InterfaceName] = &OneRecord{
			code: make(map[int]int),
		}
	}

	if s.data[s.defaultModuleName] == nil {
		s.data[s.defaultModuleName] = make(map[string]*OneRecord)
	}

	if s.data[s.defaultModuleName][s.defaultInterfaceName] == nil {
		s.data[s.defaultModuleName][s.defaultInterfaceName] = &OneRecord{
			code: make(map[int]int),
		}
	}

	re := s.data[m.ModuleName][m.InterfaceName]
	re.code[m.Code]++

	de := s.data[s.defaultModuleName][s.defaultInterfaceName]
	de.code[m.Code]++

	if m.Success {
		re.sucCount++
		re.sucCostTime += m.CostTime.Seconds()

		de.sucCount++
		de.sucCostTime += m.CostTime.Seconds()
	} else {
		re.failCount++
		re.failCostTime += m.CostTime.Seconds()

		de.failCount++
		de.failCostTime += m.CostTime.Seconds()
	}

}

func (s *stData) writeToFile(wt time.Time) []error {
	s.Lock()
	defer s.Unlock()

	var errs []error
	for moduleName, v := range s.data {
		for interfaceName, re := range v {
			if e := s.writeToOneFile(moduleName, interfaceName, re, wt); e != nil {
				errs = append(errs, e)
			}
		}
	}

	s.data = nil
	return errs
}

func (s *stData) writeToOneFile(moduleName string, interfaceName string, re *OneRecord, wt time.Time) error {
	text := fmt.Sprintf("%s\t%d\t%d\t%f\t%d\t%f\t%s\n", LocalIPByHostsFile(), wt.Unix(), re.sucCount, re.sucCostTime, re.failCount, re.failCostTime, mustJsonMarshal(re.code))
	path := fmt.Sprintf("%s/%s", s.filePath, moduleName)
	name := fmt.Sprintf("%s|%s", interfaceName, wt.Format("2006-01-02"))

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
