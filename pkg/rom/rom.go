package rom

import (
	"github.com/bokuweb/gopher-boy/pkg/types"
)

// ROM
type ROM struct {
	data []byte
}

// NewROM is ROM constructor
func NewROM(v []byte) *ROM {
	return &ROM{
		data: v,
	}
}

func (r *ROM) Read(addr types.Word) byte {
	return r.data[addr]
}
