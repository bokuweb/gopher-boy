package mocks

import (
	"github.com/bokuweb/gopher-boy/pkg/types"
)

type MockBus struct {
	MockMemory [0x10000]byte
}

func (b *MockBus) WriteByte(addr types.Word, data byte) {
	b.MockMemory[addr] = data
}

func (b *MockBus) WriteWord(addr types.Word, data types.Word) {
	b.MockMemory[addr] = byte(data & 0xFF)
	b.MockMemory[addr+1] = byte(data >> 8)
}

func (b *MockBus) ReadByte(addr types.Word) byte {
	return b.MockMemory[addr]
}

func (b *MockBus) ReadWord(addr types.Word) types.Word {
	upper := types.Word(b.MockMemory[addr]) << 8
	return upper + types.Word(b.MockMemory[addr+1])
}

func (b *MockBus) SetMemory(offset types.Word, data []byte) {
	for i, d := range data {
		b.MockMemory[offset+types.Word(i)] = d
	}
}
