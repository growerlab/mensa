package mensa

// ControlCode todo
type ControlCode uint8

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
