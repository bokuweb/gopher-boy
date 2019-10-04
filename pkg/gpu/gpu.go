package gpu

import (
	"image/color"

	"github.com/bokuweb/gopher-boy/pkg/constants"
	"github.com/bokuweb/gopher-boy/pkg/interfaces/bus"
	"github.com/bokuweb/gopher-boy/pkg/interfaces/interrupt"
	irq "github.com/bokuweb/gopher-boy/pkg/interrupt"
	"github.com/bokuweb/gopher-boy/pkg/types"
)

// CyclePerLine is gpu clock count per line
const CyclePerLine uint = 456

// LCDVBlankHeight means vblank height
const LCDVBlankHeight uint = 10

const spriteNum = 40

// GPU is
type GPU struct {
	bus             bus.Accessor
	irq             interrupt.Interrupt
	imageData       []byte
	mode            GPUMode
	clock           uint
	lcdc            byte
	stat            byte
	ly              uint
	lyc             byte
	scrollX         byte
	scrollY         byte
	windowX         byte
	windowY         byte
	bgPalette       byte
	objPalette0     byte
	objPalette1     byte
	disableDisplay  bool
	oamDMAStarted   bool
	oamDMAStartAddr types.Word
}

// GPUMode
type GPUMode = byte

const (
	// HBlankMode is period CPU can access the display RAM ($8000-$9FFF).
	HBlankMode GPUMode = iota
	// period and the CPU can access the display RAM ($8000-$9FFF).
	VBlankMode
	SearchingOAMMode
	TransferingData
)

// GPU register addresses
const (
	LCDC types.Word = 0x00
	STAT            = 0x01
	// Scroll Y (R/W)
	// 8 Bit value $00-$FF to scroll BG Y screen
	// position.
	SCROLLY = 0x02
	// Scroll X (R/W)
	// 8 Bit value $00-$FF to scroll BG X screen
	// position.
	SCROLLX = 0x03
	// LY Y-Coordinate (R)
	// The LY indicates the vertical line to which
	// the present data is transferred to the LCD
	// Driver. The LY can take on any value
	// between 0 through 153. The values between
	// 144 and 153 indicate the V-Blank period.
	// Writing will reset the counter.
	LY  = 0x04
	LYC = 0x05
	// BGP - BG & Window Palette Data (R/W)
	// Bit 7-6 - Data for Dot Data 11
	// (Normally darkest color)
	// Bit 5-4 - Data for Dot Data 10
	// Bit 3-2 - Data for Dot Data 01
	// Bit 1-0 - Data for Dot Data 00
	// (Normally lightest color)
	// This selects the shade of grays to use
	// for the background (BG) & window pixels.
	// Since each pixel uses 2 bits, the
	// corresponding shade will be selected from here.
	DMA  = 0x06
	BGP  = 0x07
	OBP0 = 0x08
	OBP1 = 0x09
	WX   = 0x0B
	WY   = 0x0A
)

const (
	TILEMAP0  types.Word = 0x9800
	TILEMAP1             = 0x9C00
	TILEDATA0            = 0x8800
	TILEDATA1            = 0x8000
	OAMSTART             = 0xFE00
)

// NewGPU is GPU constructor
func NewGPU() *GPU {
	return &GPU{
		imageData:       make([]byte, constants.ScreenWidth*constants.ScreenHeight*4),
		mode:            HBlankMode,
		clock:           0,
		lcdc:            0x91,
		ly:              0,
		scrollX:         0,
		scrollY:         0,
		oamDMAStarted:   false,
		oamDMAStartAddr: 0,
	}
}

// Init initialize GPU
func (g *GPU) Init(bus bus.Accessor, irq interrupt.Interrupt) {
	g.bus = bus
	g.irq = irq
}

// Step is run GPU
func (g *GPU) Step(cycles uint) {
	if g.bus == nil {
		panic("Please initialize gpu with Init, before running.")
	}
	g.updateMode()

	g.clock += cycles

	if !g.lcdEnabled() {
		return
	}
	if g.clock >= CyclePerLine {
		if g.ly == constants.ScreenHeight {
			g.buildSprites()
			g.irq.SetIRQ(irq.VerticalBlankFlag)
			if g.vBlankInterruptEnabled() {
				g.irq.SetIRQ(irq.LCDSFlag)
			}
		} else if g.ly >= constants.ScreenHeight+LCDVBlankHeight {
			g.ly = 0
			g.buildBGTile()
		} else if g.ly < constants.ScreenHeight {
			g.buildBGTile()
			if g.windowEnabled() {
				g.buildWindowTile()
			}
		}

		if g.ly == uint(g.lyc) {
			g.stat |= 0x04
			if g.coincidenceInterruptEnabled() {
				g.irq.SetIRQ(irq.LCDSFlag)
			}
		} else {
			g.stat &= 0xFB
		}
		g.ly++
		g.clock -= CyclePerLine
	}
}

func (g *GPU) lcdEnabled() bool {
	return (g.lcdc & 0x80) == 0x80
}

func (g *GPU) longSprite() bool {
	return (g.lcdc & 0x04) == 0x04
}

func (g *GPU) coincidenceInterruptEnabled() bool {
	return (g.stat & 0x40) == 0x40
}

func (g *GPU) vBlankInterruptEnabled() bool {
	return (g.stat & 0x10) == 0x10
}

func (g *GPU) hblankInterruptEnabled() bool {
	return (g.stat & 0x08) == 0x08
}

func (g *GPU) Read(addr types.Word) byte {
	switch addr {
	case LCDC:
		return g.lcdc
	case STAT:
		return g.stat&0xF8 | (byte(g.mode)) | 0x80
	case SCROLLX:
		return g.scrollX
	case SCROLLY:
		return g.scrollY
	case LY:
		return byte(g.ly)
	case BGP:
		return g.bgPalette
	case OBP0:
		return g.objPalette0
	case OBP1:
		return g.objPalette1
	case WX:
		return g.windowX
	case WY:
		return g.windowY
	}
	return 0x00
}

func (g *GPU) updateMode() {
	switch {
	case g.ly > constants.ScreenHeight:
		g.mode = VBlankMode
	case g.clock <= 80:
		g.mode = SearchingOAMMode
	case g.clock >= 167 && g.clock <= 291:
		g.mode = TransferingData
	default:
		g.mode = HBlankMode
		if g.hblankInterruptEnabled() {
			g.irq.SetIRQ(irq.LCDSFlag)
		}
	}

}

func (g *GPU) windowEnabled() bool {
	return g.lcdc&0x20 == 0x20
}

func (g *GPU) getWindowTilemapAddr() types.Word {
	if g.lcdc&0x40 == 0x40 {
		return TILEMAP1
	}
	return TILEMAP0
}

func (g *GPU) getBGTilemapAddr() types.Word {
	if g.lcdc&0x08 == 0x08 {
		return TILEMAP1
	}
	return TILEMAP0
}

func (g *GPU) getTileDataAddr() types.Word {
	if !g.tileData0Selected() {
		return TILEDATA1
	}
	return TILEDATA0
}

func (g *GPU) Write(addr types.Word, data byte) {
	switch addr {
	case LCDC:
		g.lcdc = data
	case STAT:
		// bit2-0 are flags
		g.stat = (g.stat & 0x07) | data
	case SCROLLX:
		g.scrollX = data
	case SCROLLY:
		g.scrollY = data
	case LY:
		g.ly = 0
	case LYC:
		g.lyc = data
	case BGP:
		g.bgPalette = data
	case OBP0:
		g.objPalette0 = data
	case OBP1:
		g.objPalette1 = data
	case DMA:
		g.oamDMAStarted = true
		g.oamDMAStartAddr = types.Word(data) * 0x100
	case WX:
		g.windowX = data
	case WY:
		g.windowY = data
	}
}

// GetImageData is image data getter
func (g *GPU) GetImageData() []byte {
	return g.imageData
}

func (g *GPU) DMAStarted() bool {
	return g.oamDMAStarted
}

func (g *GPU) Transfer() {
	for i := 0; i < 0xA0; i++ {
		data := g.bus.ReadByte(g.oamDMAStartAddr + types.Word(i))
		g.bus.WriteByte(OAMSTART+types.Word(i), data)
	}
	g.oamDMAStarted = false
}

func (g *GPU) buildSprites() {
	for i := 0; i < spriteNum; i++ {
		offsetY := int(g.bus.ReadByte(types.Word(OAMSTART+i*4))) - 16
		offsetX := int(g.bus.ReadByte(types.Word(OAMSTART+i*4+1))) - 8
		tileID := g.bus.ReadByte(types.Word(OAMSTART + i*4 + 2))
		config := types.Word(g.bus.ReadByte(types.Word(OAMSTART + i*4 + 3)))
		// aboveBG := config&0x80 == 0
		yFlip := config&0x40 != 0
		xFlip := config&0x20 != 0
		isPallette1 := config&0x10 != 0
		height := 8
		if g.longSprite() {
			height = 16
			// LSB is ignored (treated as 0) in 8x16 mode.)
			tileID = tileID & 0xFE
		}
		for x := 0; x < 8; x++ {
			for y := 0; y < height; y++ {
				if offsetX+x < 0 || offsetX+x >= constants.ScreenWidth {
					continue
				}
				if offsetY+y < 0 || offsetY+y >= constants.ScreenHeight {
					continue
				}

				paletteID := g.getSpritePaletteID(int(tileID), x, uint(y))
				adjustedX := x
				if xFlip {
					adjustedX = 7 - x
				}
				adjustedY := y
				if yFlip {
					adjustedY = 7 - y
				}
				var c byte
				if isPallette1 {
					c = (g.objPalette1 >> (paletteID * 2)) & 0x03
				} else {
					c = (g.objPalette0 >> (paletteID * 2)) & 0x03
				}
				if paletteID != 0 {
					rgba := g.getPalette(c)
					base := (uint(offsetY+adjustedY)*constants.ScreenWidth + uint(adjustedX+offsetX)) * 4
					g.imageData[base] = rgba.R
					g.imageData[base+1] = rgba.G
					g.imageData[base+2] = rgba.B
					g.imageData[base+3] = rgba.A
				}
			}
		}
	}
}

func (g *GPU) buildBGTile() {
	var tileID int
	for x := 0; x < constants.ScreenWidth; x++ {
		tileY := ((g.ly + uint(g.scrollY)) % 0x100) / 8 * 32
		tileID = g.getTileID(tileY, uint(x+int(g.scrollX))/8%32, g.getBGTilemapAddr())
		paletteID := g.getBGPaletteID(tileID, int(g.scrollX%8)+x, (g.ly+uint(g.scrollY))%8)
		rgba := g.getBGPalette(uint(paletteID))
		base := ((g.ly)*constants.ScreenWidth + uint(x)) * 4
		g.imageData[base] = rgba.R
		g.imageData[base+1] = rgba.G
		g.imageData[base+2] = rgba.B
		g.imageData[base+3] = rgba.A
	}
}

func (g *GPU) buildWindowTile() {
	var tileID int
	if (g.windowX < 0 && g.windowX >= 167) && (g.windowY < 0 && g.windowY >= 144) {
		return
	}
	if g.ly < uint(g.windowY) {
		return
	}
	for x := 0; x < constants.ScreenWidth; x++ {
		offsetX := uint(g.windowX - 7)
		if x < int(offsetX) {
			continue
		}
		tileY := (g.ly - uint(g.windowY)) / 8 * 32
		tileID = g.getTileID(tileY, uint(x-int(offsetX))/8, g.getWindowTilemapAddr())
		paletteID := g.getBGPaletteID(tileID, int(x-int(offsetX)), (g.ly-uint(g.windowY))%8)

		rgba := g.getBGPalette(uint(paletteID))
		base := ((g.ly)*constants.ScreenWidth + uint(x)) * 4
		g.imageData[base] = rgba.R
		g.imageData[base+1] = rgba.G
		g.imageData[base+2] = rgba.B
		g.imageData[base+3] = rgba.A
	}
}

func (g *GPU) tileData0Selected() bool {
	return g.lcdc&0x10 != 0x10
}

func (g *GPU) getSpritePaletteID(tileID int, x int, y uint) byte {
	x = x % 8
	addr := types.Word(tileID * 0x10)
	base := types.Word(TILEDATA1 + addr + types.Word(y*2))
	l1 := g.bus.ReadByte(base)
	l2 := g.bus.ReadByte(base + 1)
	paletteID := byte(0)
	if l1&(0x01<<(7-uint(x))) != 0 {
		paletteID = 1
	}
	if l2&(0x01<<(7-uint(x))) != 0 {
		paletteID += 2
	}
	return paletteID
}

func (g *GPU) getBGPaletteID(tileID int, x int, y uint) byte {
	x = x % 8
	var addr types.Word
	// In the first case, patterns are numbered with unsigned numbers from 0 to 255 (i.e.
	// 	pattern #0 lies at address $8000). In the second case,
	// 	patterns have signed numbers from -128 to 127 (i.e.
	// 	pattern #0 lies at address $9000). The Tile Data Table
	// 	address for the background can be selected via LCDC	register.
	if g.tileData0Selected() {
		addr = types.Word((int(int8(tileID)) + 128) * 0x10)
	} else {
		addr = types.Word(tileID * 0x10)
	}
	base := types.Word(g.getTileDataAddr() + addr + types.Word(y*2))
	l1 := g.bus.ReadByte(base)
	l2 := g.bus.ReadByte(base + 1)
	paletteID := byte(0)
	if l1&(0x01<<(7-uint(x))) != 0 {
		paletteID = 1
	}
	if l2&(0x01<<(7-uint(x))) != 0 {
		paletteID += 2
	}
	return paletteID
}

func (g *GPU) getTileID(tileY, lineOffset uint, offsetAddr types.Word) int {
	addr := types.Word(tileY) + types.Word(lineOffset) + offsetAddr
	id := byte(g.bus.ReadByte(addr))
	return int(id)

}

func (g *GPU) getBGPalette(n uint) color.RGBA {
	c := (g.bgPalette >> (n * 2)) & 0x03
	return g.getPalette(c)
}

func (g *GPU) getPalette(c byte) color.RGBA {
	switch c {
	case 0:
		return color.RGBA{175, 197, 160, 255}
	case 1:
		return color.RGBA{93, 147, 66, 255}
	case 2:
		return color.RGBA{22, 63, 48, 255}
	case 3:
		return color.RGBA{0, 40, 0, 255}
	}
	panic("unhandled color number detected.")
}
