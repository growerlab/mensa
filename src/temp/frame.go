package mensa

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

// ControlCode todo
type ControlCode uint8

// Error
var (
	ErrExceedFrameLen = errors.New("Exceeded Frame Length Limit")
)

// ControlCode
const (
	NOCONTROL     ControlCode = 0
	PING          ControlCode = 1
	CLOSEIO       ControlCode = 2
	EXITCODE      ControlCode = 3
	CONTROLRESULT ControlCode = 4
)

// ParseFrameHeader todo
func ParseFrameHeader(bf [4]byte) error {

	return nil
}

// StartFrame todo
type StartFrame struct {
	Argv []string
	Env  []string
}

// WriteStartFrame todo
// 2byte 1byte 1byte 2byte-argvlen argv... 2byte-envlen env...
func WriteStartFrame(w io.Writer, sf *StartFrame) error {
	buf := make([]byte, 65536)
	buf[2] = 0
	buf[3] = 0
	offset := 6
	ba := bytes.NewBuffer(buf[offset:])
	argvlen := 0
	for _, s := range sf.Argv {
		argvlen += len(s) + 1
		if argvlen+offset >= 65535 {
			return ErrExceedFrameLen
		}
		ba.WriteString(s)
		ba.WriteByte(0)
	}
	binary.LittleEndian.PutUint16(buf[offset-2:], uint16(argvlen))
	offset = ba.Len() + 8
	be := bytes.NewBuffer(buf[offset:])
	envlen := 0
	for _, s := range sf.Env {
		envlen += len(s) + 1
		if envlen+offset >= 65535 {
			return ErrExceedFrameLen
		}
		be.WriteString(s)
		be.WriteByte(0)
	}
	binary.LittleEndian.PutUint16(buf[offset-2:], uint16(envlen))
	total := offset + be.Len()
	binary.LittleEndian.PutUint16(buf, uint16(total))
	if _, err := w.Write(buf[0:total]); err != nil {
		return err
	}
	return nil
}
