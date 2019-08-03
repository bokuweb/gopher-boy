package timer

import (
	"github.com/bokuweb/gopher-boy/pkg/types"
)

const (
	// TimerRegisterOffset is register offset address
	TimerRegisterOffset types.Word = 0xFF00
	// DIV - Divider Register (R/W)
	// This register is incremented 16384 (~16779 on SGB) times a second.
	// Writing any value sets it to $00.
	DIV = 0x04
	// TIMA - Timer counter (R/W)
	// This timer is incremented by a clock frequency specified by the TAC register ($FF07).
	// The timer generates an interrupt when it overflows.
	TIMA = 0x05
	// TMA - Timer Modulo (R/W)
	// When the TIMA overflows, this data will be loaded.
	TMA = 0x06
	// TAC - Timer Control (R/W)
	// Bit 2 - Timer Stop
	//         0: Stop Timer
	//         1: Start Timer
	// Bits 1+0 - Input Clock Select
	//         00: 4.096 KHz (~4.194 KHz SGB)
	//         01: 262.144 Khz (~268.4 KHz SGB)
	//         10: 65.536 KHz (~67.11 KHz SGB)
	//         11: 16.384 KHz (~16.78 KHz SGB)
	TAC = 0x07
)

// Timer has 4 registers.
type Timer struct {
	internalCounter uint16
	TIMA            byte
	TAC             byte
	TMA             byte
}

// NewTimer constructs timer peripheral.
func NewTimer() *Timer {
	return &Timer{
		// 4.194304MHz / 256 = 16.384KHz
		internalCounter: 0,
		TIMA:            0x00,
		TAC:             0x00,
		TMA:             0x00,
	}
}

// Update timer counter registers
// If timer is overflowed return true
func (timer *Timer) Update(cycles uint) bool {
	r := false
	for cycles > 0 {
		cycles--
		old := timer.internalCounter
		timer.internalCounter += 4

		if !timer.isStarted() {
			continue
		}
		if !timer.hasFallingEdgeDetected(old, timer.internalCounter) {
			continue
		}
		timer.TIMA++
		if timer.TIMA == 0 {
			timer.TIMA = timer.TMA
			r = true
		}
	}
	return r
}

func (timer *Timer) Read(addr types.Word) byte {
	switch addr {
	case DIV:
		return byte(timer.internalCounter >> 8)
	case TIMA:
		return timer.TIMA
	case TMA:
		return timer.TMA
	case TAC:
		return timer.TAC
	}
	panic("Illegal access detected.")
}

func (timer *Timer) Write(addr types.Word, data byte) {
	switch addr {
	case DIV:
		// Writing any value sets it to $00.
		// When writing to DIV, the whole counter is reseted, so the timer is also affected.
		// When writing to DIV, if the current output is '1' and timer is enabled
		// as the new value after reseting DIV will be '0', the falling edge detector will detect a falling edge and TIMA will increase.
		if timer.hasFallingEdgeDetected(timer.internalCounter, 0) {
			timer.TIMA++
		}
		timer.internalCounter = 0
	case TIMA:
		timer.TIMA = data
	case TMA:
		timer.TMA = data
	case TAC:
		timer.TAC = data
	}
}

func (timer *Timer) isStarted() bool {
	return timer.TAC&0x04 == 0x04
}

func (timer *Timer) hasFallingEdgeDetected(old, new uint16) bool {
	mask := uint16(1 << timer.getMaskBit())
	return ((old & mask) != 0) && ((new & mask) == 0)
}

func (timer *Timer) getMaskBit() uint {
	switch timer.TAC & 0x03 {
	case 0x00:
		return 9
	case 0x01:
		return 3
	case 0x02:
		return 5
	case 0x03:
		return 7
	}
	return 0
}
