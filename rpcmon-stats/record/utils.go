package record

import (
	"encoding/json"
	"fmt"
	"github.com/zeast/logs"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Logger interface {
	Debugf(string, ...interface{})
	Infof(string, ...interface{})
	Warnf(string, ...interface{})
	Errorf(string, ...interface{})
}

var logger Logger

func init() {
	logger = logs.NewLogger(os.Stdout)
}

func SetLogger(l Logger) {
	logger = l
}

func mustJsonMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return data
}

func splitNetUri(uri string) (string, string) {
	tmp := strings.Split(uri, "://")
	return tmp[0], tmp[1]
}

//清除指定目录下，
func clearFile(root string, expTime time.Duration) {
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && info.ModTime().Before(time.Now().Add(-expTime)) {
			fmt.Println(path, info)
			os.Remove(filepath.Join(root, info.Name()))
		}
		return nil
	})
}

func LocalIPByHostsFile() string {
	name, err := os.Hostname()
	if err != nil {
		return "127.0.0.1"
	}

	addr, err := net.LookupHost(name)
	if err != nil {
		return "127.0.0.1"
	}

	if len(addr) == 0 {
		return "127.0.0.1"
	}

	return addr[0]
}
