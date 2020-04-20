package record

import (
	"github.com/zeast/logs"
	"net"
	"time"
)

type failedQpsReport struct {
	data               map[string]map[string]map[string]map[string]map[int64]int //projectName=>ModuleName=>InterfaceName=>userName=>Time
	projectName        string
	uri                string
	dataReportInterval time.Duration
}

// 统计失败的 qps 数据，并定期上报
func newFailedQpsReport() *failedQpsReport {
	return &failedQpsReport{
		uri:                "udp://10.17.46.36:3010",
		dataReportInterval: time.Second,
	}
}

func (q *failedQpsReport) withProjectName(name string) *failedQpsReport {
	q.projectName = name
	return q
}

func (q *failedQpsReport) start() {
	reportTicker := time.NewTicker(q.dataReportInterval)
	for {
		select {
		case <-reportTicker.C:
			if err := q.report(); err != nil {
				logs.Errorf("report failed qps error. %s \n", err)
			}
		}
	}
}

func (q *failedQpsReport) add(m *StatsMessage) {
	if m.Success {
		return
	}

	if q.projectName == "" {
		return
	}

	if q.data == nil {
		q.data = make(map[string]map[string]map[string]map[string]map[int64]int)
	}

	if q.data[q.projectName] == nil {
		q.data[q.projectName] = make(map[string]map[string]map[string]map[int64]int)
	}

	if q.data[q.projectName][m.ModuleName] == nil {
		q.data[q.projectName][m.ModuleName] = make(map[string]map[string]map[int64]int)
	}

	if q.data[q.projectName][m.ModuleName][m.InterfaceName] == nil {
		q.data[q.projectName][m.ModuleName][m.InterfaceName] = make(map[string]map[int64]int)
	}

	if q.data[q.projectName][m.ModuleName][m.InterfaceName][m.User] == nil {
		q.data[q.projectName][m.ModuleName][m.InterfaceName][m.User] = make(map[int64]int)
	}

	q.data[q.projectName][m.ModuleName][m.InterfaceName][m.User][m.Time.Unix()]++
}

func (q *failedQpsReport) report() error {
	if q.data == nil {
		return nil
	}

	conn, err := net.Dial(splitNetUri(q.uri))
	if err != nil {
		return err
	}
	defer conn.Close()

	data := mustJsonMarshal(q.data)

	logger.Debugf("failed qps report data: %s", string(data))
	_, err = conn.Write(data)
	q.data = nil
	return err
}
