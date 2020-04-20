package larkProto

import (
	"encoding/binary"
	"encoding/json"
)

const (
	lengthOfTag    = 2
	lengthOflength = 4
	lengthOfHeader = lengthOfTag + lengthOflength
)

func Pack(tag uint16, segments ...[]byte) []byte {
	var totalLen int
	for _, segment := range segments {
		totalLen += lengthOflength + len(segment)
	}

	var data = make([]byte, totalLen+lengthOfHeader)

	binary.BigEndian.PutUint16(data[:lengthOfTag], tag)
	binary.BigEndian.PutUint32(data[lengthOfTag:lengthOfHeader], uint32(totalLen))

	var offset = lengthOfHeader
	for _, segment := range segments {
		binary.BigEndian.PutUint32(data[offset:offset+lengthOflength], uint32(len(segment)))
		copy(data[offset+lengthOflength:offset+lengthOflength+len(segment)], segment)
		offset += lengthOflength + len(segment)
	}
	return data[:offset]
}

func PackBizReq(data []byte, o *LarkReqOption) ([]byte, error) {
	b, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}

	return Pack(1000, data, nil, b), nil
}

func PackBizResp(data []byte) []byte {
	return Pack(4000, data)
}

type ServiceRegistInfo struct {
	Ip          string `json:"ip"`
	Port        int    `json:"port"`
	ServiceName string `json:"service"`
	WorkNum     int    `json:"worker_num"`
	CgiPass     string `json:"cgi_pass"`
	Kind        string `json:"kind"` //用于部分逻辑的区别处理，现在有重试逻辑
}

//rc = register center
func PackRegisterReq(rinfo []ServiceRegistInfo, rclist map[string][]string) ([]byte, error) {
	s1, err := json.Marshal(rinfo)
	if err != nil {
		return nil, err
	}

	s2, err := json.Marshal(rclist)
	if err != nil {
		return nil, err
	}

	return Pack(5000, s1, s2), nil
}

func PackUnRegisterReq(rinfo []ServiceRegistInfo, rclist map[string][]string) ([]byte, error) {
	s1, err := json.Marshal(rinfo)
	if err != nil {
		return nil, err
	}

	s2, err := json.Marshal(rclist)
	if err != nil {
		return nil, err
	}

	return Pack(5010, s1, s2), nil
}
