package bus

import (
	"github.com/bokuweb/gopher-boy/pkg/interfaces/pad"
	"github.com/bokuweb/gopher-boy/pkg/interrupt"

	"github.com/bokuweb/gopher-boy/pkg/cartridge"
	"github.com/bokuweb/gopher-boy/pkg/gpu"
	"github.com/bokuweb/gopher-boy/pkg/interfaces/logger"
	"github.com/bokuweb/gopher-boy/pkg/ram"
	"github.com/bokuweb/gopher-boy/pkg/serial"
	"github.com/bokuweb/gopher-boy/pkg/timer"
	"github.com/bokuweb/gopher-boy/pkg/types"
	"github.com/bokuweb/gopher-boy/pkg/utils"
)

const (
	// DMGStatusReg is DMA status register
	DMGStatusReg types.Word = 0xFF50
)

// Bus is gb bus
type Bus struct {
	logger    logger.Logger
	bootmode  bool
	cartridge *cartridge.Cartridge
	gpu       *gpu.GPU
	vRAM      *ram.RAM
	wRAM      *ram.RAM
	hRAM      *ram.RAM
	oamRAM    *ram.RAM
	timer     *timer.Timer
	irq       *interrupt.Interrupt
	pad       pad.Pad
}

/* --------------------------+
| Interrupt Enable Register  |
------------------------------  0xFFFF
| Internal RAM               |
------------------------------  0xFF80
| Empty but unusable for I/O |
------------------------------  0xFF4C
| I/O ports                  |
------------------------------  0xFF00
| Empty but unusable for I/O |
------------------------------  0xFEA0
| Sprite Attrib Memory (OAM) |
------------------------------  0xFE00
| Echo of 8kB Internal RAM   |
------------------------------  0xE000
| 8kB Internal RAM           |
------------------------------  0xC000
| 8kB switchable RAM bank    |
------------------------------  0xA000
| 8kB Video RAM              |
------------------------------  0x8000 --+
| 16kB switchable ROM bank   |           |
------------------------------  0x4000   | =  32kB Cartrigbe
| 16kB ROM bank #0           |           |
------------------------------  0x0000 --+   */

// NewBus is bus constructor
func NewBus(
	logger logger.Logger,
	cartridge *cartridge.Cartridge,
	gpu *gpu.GPU,
	vram *ram.RAM,
	wram *ram.RAM,
	hRAM *ram.RAM,
	oamRAM *ram.RAM,
	timer *timer.Timer,
	irq *interrupt.Interrupt,
	pad pad.Pad) *Bus {
	return &Bus{
		logger:    logger,
		bootmode:  true,
		cartridge: cartridge,
		gpu:       gpu,
		vRAM:      vram,
		wRAM:      wram,
		hRAM:      hRAM,
		oamRAM:    oamRAM,
		timer:     timer,
		irq:       irq,
		pad:       pad,
	}
}

// ReadByte is byte data reader from bus
func (b *Bus) ReadByte(addr types.Word) byte {

	switch {
	case addr >= 0x0000 && addr <= 0x7FFF:
		if b.bootmode && addr < 0x0100 {
			return BIOS[addr]
		}
		if addr == 0x0100 {
			b.bootmode = false
		}
		return b.cartridge.ReadByte(addr)
	// Video RAM
	case addr >= 0x8000 && addr <= 0x9FFF:
		return b.vRAM.Read(addr - 0x8000)
	case addr >= 0xA000 && addr <= 0xBFFF:
		return b.cartridge.ReadByte(addr)
	// Working RAM
	case addr >= 0xC000 && addr <= 0xDFFF:
		return b.wRAM.Read(addr - 0xC000)
	// Shadow
	case addr >= 0xE000 && addr <= 0xFDFF:
		return b.wRAM.Read(addr - 0xE000)
	// OAM
	case addr >= 0xFE00 && addr <= 0xFE9F:
		return b.oamRAM.Read(addr - 0xFE00)
	// Pad
	case addr == 0xFF00:
		return b.pad.Read()
	// Timer
	case addr >= 0xFF04 && addr <= 0xFF07:
		return b.timer.Read(addr - 0xFF00)
	// IF
	case addr == 0xFF0F:
		return b.irq.Read(addr - 0xFF00)
	// GPU
	case addr >= 0xFF40 && addr <= 0xFF7F:
		return b.gpu.Read(addr - 0xFF40)
	// Zero page RAM
	case addr >= 0xFF80 && addr <= 0xFFFE:
		return b.hRAM.Read(addr - 0xFF80)
	// IE
	case addr == 0xFFFF:
		return b.irq.Read(addr - 0xFF00)
	default:
	}
	return 0
}

// ReadWord is word data reader from bus
func (b *Bus) ReadWord(addr types.Word) types.Word {
	l := b.ReadByte(addr)
	u := b.ReadByte(addr + 1)
	return utils.Bytes2Word(u, l)
}

// WriteByte is byte data writer to bus
func (b *Bus) WriteByte(addr types.Word, data byte) {
	switch {
	case addr >= 0x0000 && addr <= 0x7FFF:
		b.cartridge.WriteByte(addr, data)
	// Video RAM
	case addr >= 0x8000 && addr <= 0x9FFF:
		b.vRAM.Write(addr-0x8000, data)
	case addr >= 0xA000 && addr <= 0xBFFF:
		b.cartridge.WriteByte(addr, data)
	// Working RAM
	case addr >= 0xC000 && addr <= 0xDFFF:
		b.wRAM.Write(addr-0xC000, data)
	// Shadow
	case addr >= 0xE000 && addr <= 0xFDFF:
		b.wRAM.Write(addr-0xE000, data)
	// OAM
	case addr >= 0xFE00 && addr <= 0xFE9F:
		b.oamRAM.Write(addr-0xFE00, data)
	// Pad
	case addr == 0xFF00:
		b.pad.Write(data)
	// Serial
	case addr == 0xFF01:
		serial.Send(data)
	// Timer
	case addr >= 0xFF04 && addr <= 0xFF07:
		b.timer.Write(addr-0xFF00, data)
	// IF
	case addr == 0xFF0F:
		b.irq.Write(addr-0xFF00, data)
	// GPU
	case addr >= 0xFF40 && addr <= 0xFF7F:
		b.gpu.Write(addr-0xFF40, data)
	//Zero page RAM
	case addr >= 0xFF80 && addr <= 0xFFFE:
		b.hRAM.Write(addr-0xFF80, data)
	// IE
	case addr == 0xFFFF:
		b.irq.Write(addr-0xFF00, data)
	default:
		// fmt.Printf("Error: You can not write 0x%X, this area is invalid or unimplemented area.\n", addr)
	}
}

// WriteWord is word data writer to bus
func (b *Bus) WriteWord(addr types.Word, data types.Word) {
	upper, lower := utils.Word2Bytes(data)
	b.WriteByte(addr, lower)
	b.WriteByte(addr+1, upper)
}

// BIOS is
var BIOS = []byte{ /* Removed*/ }
