package jmtextProto

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"strings"
	"time"
)

func MarshalReq(user string, secretKey string, class string, method string, params []interface{}, owlContext map[string]interface{}) ([]byte, error) {
	var d = &data{
		Version:   "2.0",
		User:      user,
		Password:  GetMD5ByByte([]byte(user), []byte(secretKey), ':'),
		Timestamp: time.Now().Unix(),
		Class:     "RpcClient_" + class,
		Method:    method,
		Params:    params,
	}

	packetBytes, err := pack(secretKey, d, owlContext)
	if err != nil {
		return nil, err
	}

	command := "RPC"
	b := fmt.Sprintf("%d\n%s\n%d\n%s\n", len(command), command, len(string(packetBytes)), string(packetBytes))
	return []byte(b), nil
}

// UnmarshalReq return class, method, params, owl context
func UnmarshalReq(b []byte) (string, string, []interface{}, map[string]interface{}, error) {
	vs := bytes.SplitN(b, []byte{'\n'}, 4)
	if len(vs) != 4 {
		return "", "", nil, nil, errors.Errorf("jmtext unmarshal req error. %v, %s", b, string(b))
	}

	var p = new(packet)
	err := json.Unmarshal(vs[3], &p)
	if err != nil {
		return "", "", nil, nil, errors.WithMessagef(err, "jmtext unmarshal req packet error. %q", vs[3])
	}

	data, err := p.getData()
	if err != nil {
		return "", "", nil, nil, errors.WithMessagef(err, "jmtext unmarshal req packet error. %q", vs[3])
	}

	owlc, err := p.getOwlContext()
	if err != nil {
		return "", "", nil, nil, errors.WithMessagef(err, "jmtext unmarshal req packet error. %q", vs[3])
	}

	if !strings.HasPrefix(data.Class, "RpcClient_") {
		return "", "", nil, nil, errors.Errorf("jmtext unmarshal req packet error. class name must prefix by RpcClient_, get %s ", data.Class)
	}

	class := strings.TrimPrefix(data.Class, "RpcClient_")

	return class, data.Method, data.Params, owlc, nil
}
