package cpu

import (
	"testing"

	"github.com/bokuweb/gopher-boy/pkg/interrupt"
	"github.com/bokuweb/gopher-boy/pkg/logger"
	"github.com/bokuweb/gopher-boy/pkg/mocks"
	"github.com/bokuweb/gopher-boy/pkg/types"
	"github.com/stretchr/testify/assert"
)

func setupCPU(offset types.Word, data []byte) (*CPU, *mocks.MockBus) {
	b := mocks.MockBus{}
	b.SetMemory(offset, data)
	irq := interrupt.NewInterrupt()
	l := logger.NewLogger(logger.LogLevel("Debug"))
	return NewCPU(l, &b, irq), &b
}

func TestNOP(t *testing.T) {
	cpu, _ := setupCPU(0, []byte{0})
	cpu.Step()
}

func TestLDn_nn(t *testing.T) {
	assert := assert.New(t)
	cpu, _ := setupCPU(0, []byte{0x01, 0xDE, 0xAD})
	cpu.PC = 0x00
	cpu.Step()
	assert.Equal(byte(0xAD), cpu.Regs.B, "should B equals 0xad")
	assert.Equal(byte(0xDE), cpu.Regs.C, "should C equals 0xde")
}

func TestLDrr_r(t *testing.T) {
	assert := assert.New(t)
	cpu, bus := setupCPU(0, []byte{0x02})
	cpu.PC = 0x00
	cpu.Regs.A = 0xA5
	cpu.Regs.B = 0x10
	cpu.Regs.C = 0x20
	cpu.Step()
	assert.Equal(byte(0xA5), bus.MockMemory[0x1020], "should memory equals 0xa5")
}

func TestIncrr(t *testing.T) {
	assert := assert.New(t)
	cpu, _ := setupCPU(0, []byte{0x03})
	cpu.PC = 0x00
	cpu.Regs.B = 0x10
	cpu.Regs.C = 0x20
	cpu.Step()
	assert.Equal(byte(0x10), cpu.Regs.B, "should not B incremented")
	assert.Equal(byte(0x21), cpu.Regs.C, "should C incremented")
}

func TestIncB(t *testing.T) {
	assert := assert.New(t)
	cpu, _ := setupCPU(0, []byte{0x04})
	cpu.PC = 0x00
	cpu.Regs.B = 0x10
	cpu.Step()
	assert.Equal(byte(0x11), cpu.Regs.B, "should B incremented")
}

func TestDecB(t *testing.T) {
	assert := assert.New(t)
	cpu, _ := setupCPU(0, []byte{0x05})
	cpu.PC = 0x00
	cpu.Regs.B = 0x10
	cpu.Step()
	assert.Equal(byte(0x0F), cpu.Regs.B, "should B decremented")
}

func TestLDnn_n(t *testing.T) {
	assert := assert.New(t)
	cpu, _ := setupCPU(0, []byte{0x06, 0xA5})
	cpu.PC = 0x00
	cpu.Step()
	assert.Equal(cpu.Regs.B, byte(0xA5), "should B equals 0xa5")
}
