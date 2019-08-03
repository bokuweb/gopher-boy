package bus

import (
	"testing"

	"github.com/bokuweb/gopher-boy/pkg/cartridge"
	"github.com/bokuweb/gopher-boy/pkg/gpu"
	"github.com/bokuweb/gopher-boy/pkg/interrupt"
	"github.com/bokuweb/gopher-boy/pkg/logger"
	"github.com/bokuweb/gopher-boy/pkg/pad"
	"github.com/bokuweb/gopher-boy/pkg/ram"
	"github.com/bokuweb/gopher-boy/pkg/timer"
	"github.com/bokuweb/gopher-boy/pkg/types"
	"github.com/stretchr/testify/assert"
)

func setup() (*Bus, *ram.RAM, *ram.RAM) {
	buf := make([]byte, 0x8000)
	cart, _ := cartridge.NewCartridge(buf)
	vRAM := ram.NewRAM(0x2000)
	wRAM := ram.NewRAM(0x2000)
	hRAM := ram.NewRAM(0x80)
	oamRAM := ram.NewRAM(0xA0)
	gpu := gpu.NewGPU()
	pad := pad.NewPad()
	l := logger.NewLogger(logger.LogLevel("Debug"))
	t := timer.NewTimer()
	irq := interrupt.NewInterrupt()
	return NewBus(l, cart, gpu, vRAM, wRAM, hRAM, oamRAM, t, irq, pad), wRAM, hRAM
}

func TestWRAMReadWrite(t *testing.T) {
	assert := assert.New(t)
	b, wRAM, _ := setup()
	b.WriteByte(0xC000, 0xA5)
	b.WriteWord(0xC100, 0xDEAD)
	assert.Equal(byte(0xA5), wRAM.Read(0x0000))
	assert.Equal(byte(0xA5), b.ReadByte(0xC000))
	assert.Equal(byte(0xAD), wRAM.Read(0x0100))
	assert.Equal(byte(0xDE), wRAM.Read(0x0101))
	assert.Equal(types.Word(0xDEAD), b.ReadWord(0xC100))
}

func TestShadowWRAMReadWrite(t *testing.T) {
	assert := assert.New(t)
	b, wRAM, _ := setup()
	b.WriteByte(0xE000, 0xA5)
	b.WriteWord(0xE100, 0xDEAD)
	assert.Equal(byte(0xA5), wRAM.Read(0x0000))
	assert.Equal(byte(0xA5), b.ReadByte(0xE000))
	assert.Equal(byte(0xAD), wRAM.Read(0x0100))
	assert.Equal(byte(0xDE), wRAM.Read(0x0101))
	assert.Equal(types.Word(0xDEAD), b.ReadWord(0xE100))
}

func TestZRAMReadWrite(t *testing.T) {
	assert := assert.New(t)
	b, _, hRAM := setup()
	b.WriteByte(0xFF80, 0xA5)
	b.WriteWord(0xFF90, 0xDEAD)
	assert.Equal(byte(0xA5), hRAM.Read(0x0000))
	assert.Equal(types.Word(0xDEAD), b.ReadWord(0xFF90))
}
