package record

import (
	"bytes"
	"fmt"
	"github.com/zeast/logs"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

type failedLogData struct {
	sync.Mutex
	dataBuf           *bytes.Buffer
	maxBufSize        int
	filePath          string
	expTime           time.Duration //文件的过期时间
	dataFlushInterval int           //秒
}

func newFailedLogData() *failedLogData {
	return &failedLogData{
		filePath:          "/home/logs/go-rpc-server/stats/log/",
		expTime:           time.Hour * 24 * 14,
		dataBuf:           new(bytes.Buffer),
		maxBufSize:        524288,
		dataFlushInterval: 60,
	}
}

func (r *failedLogData) withFilePath(p string) *failedLogData {
	r.filePath = p
	return r
}

func (r *failedLogData) start() {
	now := time.Now()

	d := now.Second() % r.dataFlushInterval
	if d == 0 {
		d = r.dataFlushInterval
	}
	flushTicker := time.NewTicker(time.Second * time.Duration(r.dataFlushInterval-d))
	firstFlush := true

	clearFileTicker := time.NewTicker(time.Hour)
	for {
		select {
		case <-clearFileTicker.C:
			clearFile(r.filePath, r.expTime)

		case n := <-flushTicker.C:
			r.writeToFile(n)

			if firstFlush {
				flushTicker.Stop()
				flushTicker = time.NewTicker(time.Second * time.Duration(r.dataFlushInterval))
				firstFlush = false
			}
		}
	}
}

func (r *failedLogData) add(m *StatsMessage) {
	if m.Success {
		return
	}

	r.Lock()
	defer r.Unlock()

	logStr := fmt.Sprintf("%s\t%s::%s\tCODE:%d\tMSG:%s\tsource_ip:%s\ttarget_ip:%s\n",
		m.Time.Format("2006-01-02 15:04:05"),
		m.ModuleName,
		m.InterfaceName,
		m.Code,
		m.Msg,
		m.SourceIP,
		m.TargetIP)

	r.dataBuf.WriteString(logStr)
	if r.dataBuf.Len() >= r.maxBufSize {
		r.writeToFileWithLock()
	}
}

func (r *failedLogData) writeToFile(wt time.Time) {
	r.Lock()
	defer r.Unlock()
	r.writeToFileWithLock()

}

func (r *failedLogData) writeToFileWithLock() {
	if r.dataBuf.Len() == 0 {
		return
	}

	syscall.Umask(0)
	os.MkdirAll(r.filePath, 0777)

	f, err := os.OpenFile(filepath.Join(r.filePath, time.Now().Format("2006-01-02")), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		logs.Errorf("stats failed log record error. %s", err)
		return
	}
	defer f.Close()

	_, err = f.Write(r.dataBuf.Bytes())
	if err != nil {
		logs.Errorf("stats failed log record error. %s", err)
		return
	}

	r.dataBuf.Reset()
	return
}
