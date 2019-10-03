package rom

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

func (r *ROM) Read(addr uint32) byte {
	return r.data[addr]
}
