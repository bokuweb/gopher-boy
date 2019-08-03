package pad

import "github.com/bokuweb/gopher-boy/pkg/pad"

// Pad defined keypad interface
type Pad interface {
	Press(b pad.Button)
	Release(b pad.Button)
	Read() byte
	Write(v byte)
}
