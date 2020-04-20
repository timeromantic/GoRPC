package server

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFill(t *testing.T) {
	msg1 := &rpcmonMessage{
		version:  1,
		seriesId: 2,
		cmd:      3,
		subCmd:   4,
		code:     5,
		packLen:  6,
		body:     []byte("hello"),
	}

	b, err := msg1.marshal()
	assert.Nil(t, err)

	rd := bytes.NewReader(b)

	msg2 := new(rpcmonMessage)
	err = msg2.fill(rd)
	assert.Nil(t, err)

	msg1.header = msg2.header

	assert.Equal(t, msg1, msg2)

}
