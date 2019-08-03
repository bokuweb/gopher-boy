package interrupt

import (
	"github.com/bokuweb/gopher-boy/pkg/types"
)

var (
	// VerticalBlankISRAddr is Vertical Blank Interrupt Start Address
	VerticalBlankISRAddr types.Word = 0x0040
	// LCDCStatusTriggersISRAddr is LCDC Status Interrupt Start Address
	LCDCStatusTriggersISRAddr types.Word = 0x0048
	// TimeroverflowISRAddr is Timer Overflow Interrupt Start Address
	TimeroverflowISRAddr types.Word = 0x0050
	// SerialTransferISRAddr is Serial Transfer Completion Interrupt
	SerialTransferISRAddr types.Word = 0x0058
	// JoypadPressISRAddr is High-to-Low of P10-P13 Interrupt
	JoypadPressISRAddr types.Word = 0x0060
)

// IRQFlag is
type IRQFlag = byte

const (
	VerticalBlankFlag  IRQFlag = 0x01
	LCDSFlag           IRQFlag = 0x02
	TimerOverflowFlag  IRQFlag = 0x04
	SerialTransferFlag IRQFlag = 0x08
	JoypadPressFlag    IRQFlag = 0x10
)

const (
	RegisterOffset = 0xFF00
	// IF - Interrupt Flag (R/W)
	// Bit 4: Transition from High to Low of Pin
	// number P10-P13
	// Bit 3: Serial I/O transfer complete
	// Bit 2: Timer Overflow
	// Bit 1: LCDC (see STAT)
	// Bit 0: V-Blank
	// The priority and jump address for the above 5
	// 	interrupts are:
	// Interrupt Priority Start Address
	// V-Blank 1 $0040
	// LCDC Status 2 $0048 - Modes 0, 1, 2
	// LYC=LY coincide
	// (selectable)
	// Timer Overflow 3 $0050
	// Serial Transfer 4 $0058 - when transfer
	// is complete
	// Hi-Lo of P10-P13 5 $0060
	// * When more than 1 interrupts occur at the same
	// time only the interrupt with the highest priority
	// can be acknowledged. When an interrupt is used a
	// '0' should be stored in the IF register before the
	// IE register is set.
	IF = 0x0F
	// IE - Interrupt Enable (R/W)
	// Bit 4:
	// Transition from High to Low of Pin number P10-P13.
	// Serial I/O transfer complete
	// Timer Overflow
	// Bit 3:
	// Bit 2:
	// Bit 1:
	// Bit 0: V-Blank
	// 0: disable
	// 1: enable
	IE = 0xFF
)

const (
	InterruptFlagAddr       = RegisterOffset + IF
	InterruptEnableFlagAddr = RegisterOffset + IE
)

// Interrupt has 2 registers to manage
type Interrupt struct {
	IF      byte
	IE      byte
	enabled bool
}

// NewInterrupt constructs irq peripheral.
func NewInterrupt() *Interrupt {
	return &Interrupt{
		IF:      0x00,
		IE:      0x00,
		enabled: false,
	}
}

// SetIRQ set flag
func (irq *Interrupt) SetIRQ(f IRQFlag) {
	irq.IF |= f
}

func (irq *Interrupt) Read(addr types.Word) byte {
	switch addr {
	case IE:
		return irq.IE
	case IF:
		return irq.IF | 0xE0
	}
	panic("Illegal access detected.")
}

func (irq *Interrupt) Write(addr types.Word, data byte) {
	switch addr {
	case IE:
		irq.IE = data
		return
	case IF:
		irq.IF = data
		return
	}
	panic("Illegal access detected.")
}

func (irq *Interrupt) HasIRQ() bool {
	i := irq.IF & irq.IE
	return i != 0x00
}

func (irq *Interrupt) Enable() {
	irq.enabled = true
}

func (irq *Interrupt) Enabled() bool {
	return irq.enabled
}

func (irq *Interrupt) Disable() {
	irq.enabled = false
}

func (irq *Interrupt) ResolveISRAddr() *types.Word {
	i := irq.IF & irq.IE
	switch {
	case i&VerticalBlankFlag != 0:
		irq.IF &= ^VerticalBlankFlag
		return &VerticalBlankISRAddr
	case i&LCDSFlag != 0:
		irq.IF &= ^LCDSFlag
		return &LCDCStatusTriggersISRAddr
	case i&TimerOverflowFlag != 0:
		irq.IF &= ^TimerOverflowFlag
		return &TimeroverflowISRAddr
	case i&SerialTransferFlag != 0:
		irq.IF &= ^SerialTransferFlag
		return &SerialTransferISRAddr
	case i&JoypadPressFlag != 0:
		irq.IF &= ^JoypadPressFlag
		return &JoypadPressISRAddr
	}
	return nil
}
