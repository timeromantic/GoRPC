package record

import (
	"github.com/zeast/logs"
	"net"
	"time"
)

type qpsReport struct {
	data               map[string]map[string]map[string]map[string]int //projectName=>ModuleName=>InterfaceName=>userName
	projectName        string
	qpsServiceUri      string
	dataReportInterval time.Duration //秒
}

//统计 qps 数据，并定期上报
func newQpsReport() *qpsReport {
	return &qpsReport{
		qpsServiceUri:      "udp://10.17.46.36:3000",
		dataReportInterval: time.Second,
	}
}

func (q *qpsReport) withProjectName(name string) *qpsReport {
	q.projectName = name
	return q
}

func (q *qpsReport) start() {
	reportTicker := time.NewTicker(q.dataReportInterval)
	for {
		select {
		case <-reportTicker.C:
			if err := q.report(); err != nil {
				logs.Errorf("report qps error. %s \n", err)
			}
		}
	}
}

func (q *qpsReport) add(m *StatsMessage) {
	if q.projectName == "" {
		return
	}

	if q.data == nil {
		q.data = make(map[string]map[string]map[string]map[string]int)
	}

	if q.data[q.projectName] == nil {
		q.data[q.projectName] = make(map[string]map[string]map[string]int)
	}

	if q.data[q.projectName][m.ModuleName] == nil {
		q.data[q.projectName][m.ModuleName] = make(map[string]map[string]int)
	}

	if q.data[q.projectName][m.ModuleName][m.InterfaceName] == nil {
		q.data[q.projectName][m.ModuleName][m.InterfaceName] = make(map[string]int)
	}

	q.data[q.projectName][m.ModuleName][m.InterfaceName][m.User]++
}

func (q *qpsReport) report() error {
	if q.data == nil {
		return nil
	}

	conn, err := net.Dial(splitNetUri(q.qpsServiceUri))
	if err != nil {
		return err
	}
	defer conn.Close()

	data := mustJsonMarshal(q.data)
	logger.Debugf("qps report data: %s", string(data))
	_, err = conn.Write(data)
	q.data = nil
	return err
}
