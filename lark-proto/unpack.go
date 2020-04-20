package larkProto

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"
)

type ReqOption struct {
	ClientIP string `json:"client_ip"`
}

type RespOption struct {
	ServerIP string `json:"server_ip"`
}

type LarkReqOption struct {
	TargetService  string        `json:"service_name"`
	TargetAddr     string        `json:"target_addr"`
	ProcessTimeout time.Duration `json:"recv_timeout"`
	SourceService  string        `json:"client_service"`
}

func UnPack(rd io.Reader) (uint16, [][]byte, error) {
	//读取 tag 和 length
	var header [6]byte
	_, err := io.ReadFull(rd, header[:])
	if err != nil {
		return 0, nil, err
	}

	tag := binary.BigEndian.Uint16(header[:lengthOfTag])

	l := binary.BigEndian.Uint32(header[lengthOfTag:lengthOfHeader])

	//读取 payload
	var payload = make([]byte, l)
	_, err = io.ReadFull(rd, payload)
	if err != nil {
		return 0, nil, err
	}

	//解析 value
	var value [][]byte
	var offset uint32
	for offset < l {
		ll := binary.BigEndian.Uint32(payload[offset : offset+lengthOflength])
		value = append(value, payload[offset+lengthOflength:offset+lengthOflength+ll])
		offset += lengthOflength + ll
	}

	return tag, value, nil
}

func UnPackReq(rd io.Reader) ([]byte, *ReqOption, error) {
	tag, segments, err := UnPack(rd)
	if err != nil {
		return nil, nil, err
	}

	if tag != 3000 {
		return nil, nil, fmt.Errorf("want tag %d, get tag %d", 3000, tag)
	}

	if len(segments) == 0 {
		return nil, nil, errors.New("lark segments is empty")
	}

	var req []byte
	if len(segments) >= 1 {
		req = segments[0]
	}

	var o = new(ReqOption)
	if len(segments) >= 2 {
		err := json.Unmarshal(segments[1], &o)
		if err != nil {
			return nil, nil, fmt.Errorf("unmarshal req option error. %w", err)
		}
	}

	return req, o, nil
}

func UnPackBizResp(rd io.Reader) ([]byte, *RespOption, error) {
	tag, segments, err := UnPack(rd)
	if err != nil {
		return nil, nil, err
	}

	if tag != 2000 {
		if len(segments) == 0 {
			return nil, nil, fmt.Errorf("want tag:%d, get tag:%d", 2000, tag)
		} else {
			return nil, nil, errors.New(string(segments[0]))
		}
	}

	if len(segments) == 0 {
		return nil, nil, errors.New("lark segments is empty")
	}

	var req []byte
	if len(segments) >= 1 {
		req = segments[0]
	}

	var o = new(RespOption)
	if len(segments) >= 2 {
		err := json.Unmarshal(segments[1], &o)
		if err != nil {
			return nil, nil, fmt.Errorf("unmarshal resp option error. %w", err)
		}
	}

	return req, o, nil
}
