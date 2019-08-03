package interrupt

import (
	"github.com/bokuweb/gopher-boy/pkg/interrupt"
	"github.com/bokuweb/gopher-boy/pkg/types"
)

// IRQFlag is
type IRQFlag = interrupt.IRQFlag

// Interrupt defined irq interface
type Interrupt interface {
	SetIRQ(f interrupt.IRQFlag)
	Enable()
	Disable()
	Enabled() bool
	Read(addr types.Word) byte
	Write(addr types.Word, data byte)
	HasIRQ() bool
	ResolveISRAddr() *types.Word
}
