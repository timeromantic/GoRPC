package jmtextProto

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

func MarshalResp(resp interface{}) ([]byte, error) {
	b, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}

	l := strconv.Itoa(len(b))

	buf := make([]byte, len(l)+1+len(b)+1)
	n := copy(buf, l)
	n += copy(buf[n:], "\n")
	n += copy(buf[n:], b)
	copy(buf[n:], "\n")

	return buf, nil
}

func UnmarshalResp(b []byte) ([]byte, error) {
	vs := bytes.SplitN(b, []byte{'\n'}, 2)
	if len(vs) != 2 {
		return nil, fmt.Errorf("jmtext unmarshal resp error. %v, %s", b, string(b))
	}

	return vs[1], nil
}
