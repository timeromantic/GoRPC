package record

import (
	"fmt"
	"github.com/zeast/logs"
	"net"
	"strconv"
	"strings"
	"time"
)

type failedLogReport struct {
	data               map[string]map[string]map[string]*failedLog //time=>ModuleName+InterfaceName=>codeMsg=>oneFailed
	projectName        string
	uri                string
	dataReportInterval time.Duration
	maxReportSize      int //上报的时候，如果超过这个参数，就 send 一次
}

type failedLog struct {
	count int    //相同 code 的错误计数
	log   string //错误的详细信息
}

func newFaildLogReport() *failedLogReport {
	return &failedLogReport{
		uri:                "udp://10.17.46.36:3020",
		dataReportInterval: time.Second,
		maxReportSize:      65507,
	}
}

func (d *failedLogReport) withProjectName(name string) *failedLogReport {
	d.projectName = name
	return d
}

func (q *failedLogReport) start() {
	reportTicker := time.NewTicker(q.dataReportInterval)
	for {
		select {
		case <-reportTicker.C:
			if err := q.report(); err != nil {
				logs.Errorf("report failed log error. %v", err)
			}
		}
	}
}

func (q *failedLogReport) add(m *StatsMessage) {
	if q.projectName == "" || m.Success {
		return
	}

	if q.data == nil {
		q.data = make(map[string]map[string]map[string]*failedLog)
	}

	t := strconv.Itoa(m.Time.Second())

	if q.data[t] == nil {
		q.data[t] = make(map[string]map[string]*failedLog)
	}

	pmi := q.projectName + "::" + m.ModuleName + "::" + m.InterfaceName
	if q.data[t][pmi] == nil {
		q.data[t][pmi] = make(map[string]*failedLog)
	}

	codeStr := strconv.Itoa(m.Code)
	if q.data[t][pmi][codeStr] == nil {
		logStr := fmt.Sprintf("%s\t%s::%s::%s\tCODE:%d\tMSG:%s\tsource_ip:%s\ttarget_ip:%s",
			m.Time.Format("2006"),
			q.projectName,
			m.ModuleName,
			m.InterfaceName,
			m.Code,
			m.Msg,
			m.SourceIP,
			m.TargetIP)
		q.data[t][pmi][codeStr] = &failedLog{
			count: 1,
			log:   logStr,
		}
	} else {
		q.data[t][pmi][codeStr].count++
	}
}

func (q *failedLogReport) report() error {
	if q.data == nil {
		return nil
	}

	conn, err := net.Dial(splitNetUri(q.uri))
	if err != nil {
		return err
	}
	defer conn.Close()

	strb := new(strings.Builder)
	for _, pmiMap := range q.data {
		for _, codeMap := range pmiMap {
			for _, tmp := range codeMap {
				strb.WriteString(fmt.Sprintf("%s\tCOUNT:%d\n", tmp.log, tmp.count))

				if strb.Len() >= q.maxReportSize {
					if _, err := conn.Write([]byte(strb.String())); err != nil {
						return err
					}
					strb.Reset()
				}

			}
		}
	}

	if strb.Len() > 0 {
		logger.Debugf("failed log report data: %s", strb.String())
		if _, err := conn.Write([]byte(strb.String())); err != nil {
			return err
		}
	}

	strb.Reset()
	q.data = nil

	return nil
}
