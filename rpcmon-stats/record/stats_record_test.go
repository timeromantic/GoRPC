package record

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

var statsMsg = []*StatsMessage{
	&StatsMessage{
		ModuleName:    "testModuleName1",
		InterfaceName: "testInterface1",
		SourceIP:      "1.1.1.1",
		TargetIP:      "2.2.2.2",
		Success:       true,
		CostTime:      time.Millisecond * 200,
		User:          "testUser",
		Code:          110,
		Msg:           "test Msg 1",
	},
	&StatsMessage{
		ModuleName:    "testModuleName2",
		InterfaceName: "testInterface2",
		SourceIP:      "3.3.3.3",
		TargetIP:      "4.4.4.4",
		Success:       false,
		CostTime:      time.Millisecond * 200,
		User:          "testUser2",
		Code:          210,
		Msg:           "test Msg 2",
	},
}

func TestStatisticData(t *testing.T) {
	r := newStData()
	for _, msg := range statsMsg {
		r.add(msg)
	}

	errs := r.writeToFile(time.Now())
	for _, err := range errs {
		assert.Nil(t, err)
	}
}

func TestStatisticDataDetail(t *testing.T) {
	s := newDetailData()
	for _, msg := range statsMsg {
		s.add(msg)
	}

	errs := s.writeToFile(time.Now())
	for _, err := range errs {
		assert.Nil(t, err)
	}
}

func TestStatsCollect(t *testing.T) {
	os.RemoveAll("/tmp/stData/")
	defer os.RemoveAll("/tmp/stData/")

	os.RemoveAll("/tmp/detailData/")
	defer os.RemoveAll("/tmp/detailData/")

	os.RemoveAll("/tmp/userDetailData/")
	defer os.RemoveAll("/tmp/userDetailData/")

	sc := NewStatsRecord()
	sc.failedQpsReport.projectName = "test_project_name"
	sc.qpsReport.projectName = "test_project_name"
	sc.stData.filePath = "/tmp/stData/"
	sc.detailData.filePath = "/tmp/detailData/"
	sc.userDetailData.filePath = "/tmp/userDetailData/"
	sc.qpsReport.qpsServiceUri = "udp://127.0.0.1:3000"
	sc.failedQpsReport.uri = "udp://127.0.0.1:3000"

	sc.Start()

	for _, msg := range statsMsg {
		sc.Write(msg)
	}

	time.Sleep(time.Second * 3)

	sc.Stop()
}
