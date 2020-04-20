package jmtextProto

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
)

func GetMD5ByByte(prefix, suffix []byte, separator byte) string {
	var data []byte
	data = append(prefix, separator)
	data = append(data, suffix...)

	h := md5.New()
	h.Write(data)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func CheckException(b []byte) error {
	if bytes.HasPrefix(b, []byte(`{"exception"`)) {
		return errors.New(string(b))
	}

	return nil
}
