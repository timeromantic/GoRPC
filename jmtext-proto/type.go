package jmtextProto

import (
	"encoding/json"
	"github.com/pkg/errors"
)

type packet struct {
	Data      string                 `json:"data"`
	Signature string                 `json:"signature"`
	Context   map[string]interface{} `json:"CONTEXT"`
}

func (p *packet) getData() (*data, error) {
	var d = new(data)
	err := json.Unmarshal([]byte(p.Data), &d)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func (p *packet) getOwlContext() (map[string]interface{}, error) {
	var owlc = make(map[string]interface{})
	if _, ok := p.Context["owl_context"]; !ok {
		return owlc, nil
	}

	if str, ok := p.Context["owl_context"].(string); ok {
		err := json.Unmarshal([]byte(str), &owlc)
		if err != nil {
			return nil, errors.WithMessagef(err, "can't unmarshal owl context. %q", str)
		}

		return owlc, nil
	}

	return nil, errors.Errorf("owlc context is not string. %q", p.Context["owl_context"])
}

type Context struct {
	OwlContext map[string]string `json:"owl_context"`
}

type data struct {
	Version   string        `json:"version"`
	User      string        `json:"user"`
	Password  string        `json:"password"`
	Timestamp interface{}   `json:"timestamp"`
	Class     string        `json:"class"`
	Method    string        `json:"method"`
	Params    []interface{} `json:"params"`
}

type Exception struct {
	TraceID       string `json:"trace_id"`
	ServerIP      string `json:"server_ip"`
	Class         string `json:"class"`
	Message       string `json:"message"`
	Code          int    `json:"code"`
	File          string `json:"file"`
	Line          int    `json:"line"`
	TraceAsString string `json:"traceAsString"`
}

func pack(secretKey string, d *data, owlc map[string]interface{}) ([]byte, error) {
	dataBytes, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}

	owlcBytes, err := json.Marshal(owlc)
	if err != nil {
		return nil, err
	}

	sign := GetMD5ByByte(dataBytes, []byte(secretKey), '&')
	var p = &packet{
		Data:      string(dataBytes),
		Signature: sign,
		Context: map[string]interface{}{
			"owl_context": string(owlcBytes),
		},
	}

	packetBytes, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	return packetBytes, nil
}
