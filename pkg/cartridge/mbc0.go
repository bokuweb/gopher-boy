package cartridge

import (
	"github.com/bokuweb/gopher-boy/pkg/rom"
	"github.com/bokuweb/gopher-boy/pkg/types"
)

// MBC0 ROM ONLY
// This is a 32kB (256kb) ROM and occupies 0000-7FFF.
type MBC0 struct {
	rom *rom.ROM
}

// NewMBC0 constracts MBC0
func NewMBC0(data []byte) *MBC0 {
	m := new(MBC0)
	m.rom = rom.NewROM(data)
	return m
}

func (m *MBC0) Write(addr types.Word, value byte) {
}

func (m *MBC0) Read(addr types.Word) byte {
	return m.rom.Read(uint32(addr))
}

func (m *MBC0) switchROMBank(bank int) {
	// nop
}

func (m *MBC0) switchRAMBank(bank int) {
	// nop
}
