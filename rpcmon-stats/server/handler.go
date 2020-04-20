package server

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func (s *StatsServer) subCmdGetStAndModulesHandler(b []byte) ([]byte, error) {
	var body = new(Body)
	if err := json.Unmarshal(b, &body); err != nil {
		return nil, errors.WithStack(err)
	}

	if err := body.Verify(); err != nil {
		return nil, errors.WithStack(err)
	}

	modules, err := getModule(s.stDataPath, body.ModuleName)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	statistic, err := getStatistic(s.stDataPath, body.ModuleName, body.InterfaceName, body.Timestamp)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return mustJsonMarshal(map[string]interface{}{
		"modules":   modules,
		"statistic": string(statistic),
	}), nil
}

func (s *StatsServer) subCmdGetLogsHandler(b []byte) ([]byte, error) {
	var tmp = struct {
		ModuleName        string `json:"module"`
		InterfaceName     string `json:"interface"`
		StartTimestampStr string `json:"start_time"`
		EndTimestampStr   string `json:"end_time"`
		Code              string `json:"code"`
		Msg               string `json:"msg"`
		Pointer           int64  `json:"pointer"`
		Count             int    `json:"count"`
	}{
		Count: 10,
	}

	if err := json.Unmarshal(b, &tmp); err != nil {
		return nil, errors.WithStack(err)
	}
	var startTimestamp int
	if tmp.StartTimestampStr != "" {
		i, err := strconv.Atoi(tmp.StartTimestampStr)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		startTimestamp = i
	}

	var endTimestamp int
	if tmp.EndTimestampStr != "" {
		i, err := strconv.Atoi(tmp.EndTimestampStr)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		endTimestamp = i
	}

	data, err := getFailedLog(s.failedLogDataPath, tmp.ModuleName, tmp.InterfaceName, startTimestamp, endTimestamp, tmp.Code, tmp.Msg, tmp.Pointer, tmp.Count)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return data, nil
}

func (s *StatsServer) subCmdGetClientSummaryHandler(b []byte) ([]byte, error) {
	var body = new(Body)
	if err := json.Unmarshal(b, &body); err != nil {
		return nil, errors.WithStack(err)
	}

	if err := body.Verify(); err != nil {
		return nil, errors.WithStack(err)
	}

	data, err := getClientSummary(s.detailDataPath, body.ModuleName, body.InterfaceName, body.Timestamp)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return data, nil
}

func (s *StatsServer) subCmdGetClientUserSummaryHandler(b []byte) ([]byte, error) {
	var body = new(Body)
	if err := json.Unmarshal(b, &body); err != nil {
		return nil, errors.WithStack(err)
	}

	if err := body.Verify(); err != nil {
		return nil, errors.WithStack(err)
	}

	data, err := getClientSummary(s.userDetailDataPath, body.ModuleName, body.InterfaceName, body.Timestamp)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return data, nil
}

func (s *StatsServer) subCmdGetClientDetailHandler(b []byte) ([]byte, error) {
	var body = new(Body)
	if err := json.Unmarshal(b, &body); err != nil {
		return nil, errors.WithStack(err)
	}

	if err := body.Verify(); err != nil {
		return nil, errors.WithStack(err)
	}

	data, err := getClientDetail(s.detailDataPath, body.ModuleName, body.InterfaceName, body.IP, body.Timestamp)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return data, nil
}

func (s *StatsServer) subCmdGetClientUserDetailHandler(b []byte) ([]byte, error) {
	var body = new(Body)
	if err := json.Unmarshal(b, &body); err != nil {
		return nil, errors.WithStack(err)
	}

	if err := body.Verify(); err != nil {
		return nil, errors.WithStack(err)
	}

	data, err := getClientDetail(s.userDetailDataPath, body.ModuleName, body.InterfaceName, body.IP, body.Timestamp)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return data, nil
}

func (s *StatsServer) subCmdGetClientRawHandler(b []byte) ([]byte, error) {
	var body = new(Body)
	if err := json.Unmarshal(b, &body); err != nil {
		return nil, errors.WithStack(err)
	}

	if err := body.Verify(); err != nil {
		return nil, errors.WithStack(err)
	}

	data, err := getClientRaw(s.detailDataPath, body.ModuleName, body.InterfaceName, body.Timestamp)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return data, nil
}

func (s *StatsServer) subCmdGetClientUserRawHandler(b []byte) ([]byte, error) {
	var body = new(Body)
	if err := json.Unmarshal(b, &body); err != nil {
		return nil, errors.WithStack(err)
	}

	if err := body.Verify(); err != nil {
		return nil, errors.WithStack(err)
	}

	data, err := getClientRaw(s.userDetailDataPath, body.ModuleName, body.InterfaceName, body.Timestamp)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return data, nil
}

// 从目录中获取模块的名称和指定模块下的接口名称
func getModule(root string, moduleName string) (map[string]map[string]string, error) {
	matches, err := filepath.Glob(filepath.Join(root, "/*"))
	if err != nil {
		return nil, err
	}

	var data = make(map[string]map[string]string)
	for _, match := range matches {
		dir, file := filepath.Split(match)

		data[file] = make(map[string]string)

		if file != moduleName {
			continue
		}

		err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			tmp := strings.Split(info.Name(), "|")
			data[file][tmp[0]] = tmp[0]
			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	return data, nil
}

func getStatistic(base, moduleName, interfaceName string, timestamp int64) ([]byte, error) {
	if moduleName == "" || interfaceName == "" {
		return nil, nil
	}

	filename := filepath.Join(base, moduleName, interfaceName+"|"+time.Unix(timestamp, 0).Format("2006-01-02"))

	_, err := os.Stat(filename)
	if err != nil && os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return ioutil.ReadFile(filename)
}

func getClientSummary(base, moduleName, interfaceName string, timestamp int64) ([]byte, error) {
	if moduleName == "" || interfaceName == "" {
		return nil, nil
	}

	filedata, err := readAllDetailContent(base, moduleName, interfaceName, timestamp)
	if err != nil {
		return nil, err
	}

	// [ip:[suc_cnt,fail_cnt], ip:[suc_cnt,fail_cnt], ..]
	data := make(map[string][2]int)

	for _, line := range bytes.Split(filedata, []byte("\n")) {
		if len(line) == 0 {
			continue
		}

		tmp := make(map[string][2]int)
		err := json.Unmarshal(bytes.Split(line, []byte("\t"))[1], &tmp)
		if err != nil {
			return nil, err
		}

		for ip, counts := range tmp {
			d := data[ip]
			d[0] += counts[0]
			d[1] += counts[1]
			data[ip] = d
		}
	}

	return mustJsonMarshal(map[string]interface{}{
		"statistic": data,
	}), nil
}

// 查找指定 module, interface, ip, timeStr 下的一个时间点对应多少正确和错误请求
func getClientDetail(base, moduleName, interfaceName, ip string, timestamp int64) ([]byte, error) {
	if ip == "" {
		return nil, fmt.Errorf("ip is empty. module:%s, interface:%s, ip:%s", moduleName, interfaceName, ip)
	}

	filedata, err := readAllDetailContent(base, moduleName, interfaceName, timestamp)
	if err != nil {
		return nil, err
	}

	// [time:[suc_cnt,fail_cnt], time:[suc_cnt,fail_cnt], ..]
	data := make(map[string][2]int)

	for _, line := range bytes.Split(filedata, []byte("\n")) {
		tmp := make(map[string][2]int)
		lineSplit := bytes.Split(line, []byte("\t"))
		err := json.Unmarshal(bytes.Split(lineSplit[1], []byte("\t"))[1], &tmp)
		if err != nil {
			return nil, err
		}

		for cip, counts := range tmp {
			if cip == ip {
				data[string(lineSplit[0])] = [2]int{counts[0], counts[1]}
			}
		}
	}

	return mustJsonMarshal(map[string]interface{}{
		"statistic": data,
	}), nil
}

func getClientRaw(base, moduleName, interfaceName string, timestamp int64) ([]byte, error) {
	data, err := readAllDetailContent(base, moduleName, interfaceName, timestamp)
	if err != nil {
		return nil, err
	}

	return mustJsonMarshal(map[string]interface{}{
		"statistic": data,
	}), nil
}

func readAllDetailContent(base, moduleName, interfaceName string, timestamp int64) ([]byte, error) {
	if moduleName == "" || interfaceName == "" {
		return nil, fmt.Errorf("module name or interface name is empty. module:%s, interface:%s", moduleName, interfaceName)
	}

	filename := filepath.Join(base, moduleName, interfaceName+"-detail|"+time.Unix(timestamp, 0).Format("2006-01-02"))
	return ioutil.ReadFile(filename)
}

func getFailedLog(path string, moduleName string, interfaceName string, startTimestamp int, endTimestamp int, code string, msg string, pointer int64, count int) (retData []byte, retErr error) {
	defer func() {
		if x := recover(); x != nil {
			retErr = errors.Errorf("%+v", x)
		}
	}()

	var fileName string
	if startTimestamp == 0 {
		fileName = time.Now().Format("2006-01-02")
	} else {
		fileName = time.Unix(int64(startTimestamp), 0).Format("2006-01-02")
	}

	file, err := os.Open(filepath.Join(path, fileName))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	stat, err := file.Stat()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if startTimestamp != 0 && pointer == 0 && stat.Size() > 2048000 {
		var startPointer int64
		var endPointer int64 = stat.Size()

		var bsize = 4096
		var b = make([]byte, bsize)
		for startPointer < endPointer {
			mid := (startPointer + endPointer) / 2

			ret, err := file.Seek(int64(mid), 0)

			n, err := file.Read(b)
			if err != nil && err != io.EOF {
				return nil, errors.WithStack(err)
			}
			b = b[:n]

			i := bytes.IndexByte(b, '\n')
			if i == -1 {
				return nil, errors.Errorf("failed log msg is too long, max is %d", bsize)
			}

			//19 是时间字符串的长度
			if i+19 > len(b) {
				break
			}

			t, err := time.ParseInLocation("2006-01-02 15:04:05", string(b[i+1:i+20]), time.Local)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			if t.Second() < startTimestamp {
				startPointer = ret
			} else if t.Second() > startTimestamp {
				endPointer = ret
			} else {
				pointer = ret
				break
			}
		}
	}

	var pattern = `^(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})\t`
	if moduleName != "" {
		pattern += moduleName + "::"
	} else {
		pattern += ".*::"
	}

	if interfaceName != "" {
		pattern += interfaceName + `\t`
	} else {
		pattern += `.*\t`
	}

	if code != "" {
		pattern += "CODE:" + code + `\t`
	} else {
		pattern += `CODE:\d+\t`
	}

	if msg != "" {
		pattern += "MSG:" + msg
	}

	pattern += ".*"

	if _, err := file.Seek(0, 0); err != nil {
		return nil, errors.WithStack(err)
	}

	rd := bufio.NewReader(file)
	data := new(bytes.Buffer)
	for i := 0; i < count; i++ {
		line, isprefix, err := rd.ReadLine()
		if err != nil && err != io.EOF {
			return nil, errors.WithStack(err)
		}

		if len(line) == 0 {
			break
		}

		pointer += int64(len(line))

		if isprefix {
			return nil, errors.Errorf("invalid failed log msg: %s", string(line))
		}

		sli := regexp.MustCompile(pattern).FindStringSubmatch(string(line))
		if len(sli) >= 2 {
			tm, err := time.ParseInLocation("2006-01-02 15:04:05", sli[1], time.Local)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			if startTimestamp != 0 && tm.Unix() < int64(startTimestamp) {
				continue
			}

			if endTimestamp != 0 && tm.Unix() > int64(endTimestamp) {
				break
			}

			data.Write(line)
			data.WriteByte('\n')
		}
	}

	return mustJsonMarshal(map[string]interface{}{
		"pointer": pointer,
		"data":    data.String(),
	}), nil
}
