package gb

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"

	"github.com/bokuweb/gopher-boy/pkg/constants"
	"github.com/bokuweb/gopher-boy/pkg/interrupt"
	"github.com/bokuweb/gopher-boy/pkg/pad"
	"github.com/bokuweb/gopher-boy/pkg/timer"

	"github.com/bokuweb/gopher-boy/pkg/bus"
	"github.com/bokuweb/gopher-boy/pkg/cartridge"
	"github.com/bokuweb/gopher-boy/pkg/cpu"
	"github.com/bokuweb/gopher-boy/pkg/gpu"
	"github.com/bokuweb/gopher-boy/pkg/interfaces/window"
	"github.com/bokuweb/gopher-boy/pkg/logger"
	"github.com/bokuweb/gopher-boy/pkg/ram"
	"github.com/bokuweb/gopher-boy/pkg/utils"
)

const (
	RomPathPrefix   = "../../roms/"
	ImagePathPrefix = "../../test/actual/"
)

// MockWindow is
type mockWindow struct {
	window.Window
}

func (m mockWindow) PollKey() {
}

func setup(file string) *GB {
	l := logger.NewLogger(logger.LogLevel("DEBUG"))
	buf, err := utils.LoadROM(file)
	if err != nil {
		panic(err)
	}
	cart, err := cartridge.NewCartridge(buf)
	if err != nil {
		panic(err)
	}

	vRAM := ram.NewRAM(0x2000)
	wRAM := ram.NewRAM(0x2000)
	hRAM := ram.NewRAM(0x80)
	oamRAM := ram.NewRAM(0xA0)
	gpu := gpu.NewGPU()
	t := timer.NewTimer()
	pad := pad.NewPad()
	irq := interrupt.NewInterrupt()
	b := bus.NewBus(l, cart, gpu, vRAM, wRAM, hRAM, oamRAM, t, irq, pad)
	gpu.Init(b, irq)
	win := mockWindow{}
	emu := NewGB(cpu.NewCPU(l, b, irq), gpu, t, irq, win)
	return emu
}

func set(img *image.RGBA, buf []byte) {
	imgData := make([]color.RGBA, constants.ScreenWidth*constants.ScreenHeight)
	i := 0
	for i*4 < len(buf) {
		y := constants.ScreenHeight - (i / constants.ScreenWidth) - 1
		imgData[y*constants.ScreenWidth+i%constants.ScreenWidth] = color.RGBA{buf[i*4], buf[i*4+1], buf[i*4+2], buf[i*4+3]}
		i++
	}

	rect := img.Rect
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			img.Set(x, rect.Max.Y-y, imgData[y*rect.Max.X+x])
		}
	}
}

func skipFrame(emu *GB, n int) []byte {
	var image []byte
	for i := 0; i < n; i++ {
		image = emu.Next()
	}
	return image
}

func TestROMs(t *testing.T) {
	tests := []struct {
		name  string
		path  string
		frame int
	}{
		{
			"hello_world",
			RomPathPrefix + "helloworld/hello.gb",
			100,
		},
		{
			"tictactoe",
			RomPathPrefix + "tictactoe/GB-TicTacToe.gb",
			100,
		},
		{
			"cpu_instr",
			RomPathPrefix + "cpu_instrs/cpu_instrs.gb",
			4000,
		},
		{
			"opus5",
			RomPathPrefix + "opus5/opus5.gb",
			100,
		},
		{
			"div_write",
			RomPathPrefix + "acceptance/timer/div_write.gb",
			1000,
		},
		{
			"tim00",
			RomPathPrefix + "acceptance/timer/tim00.gb",
			100,
		},
		{
			"tim01",
			RomPathPrefix + "acceptance/timer/tim01.gb",
			100,
		},
		{
			"tim10",
			RomPathPrefix + "acceptance/timer/tim10.gb",
			100,
		},
		{
			"tim11",
			RomPathPrefix + "acceptance/timer/tim10.gb",
			100,
		},
		{
			"bits_bank1",
			RomPathPrefix + "mbc1/bits_bank1.gb",
			100,
		},
		{
			"reg_f",
			RomPathPrefix + "acceptance/bits/reg_f.gb",
			100,
		},
		{
			"mem_oam",
			RomPathPrefix + "acceptance/bits/mem_oam.gb",
			100,
		},
		{
			"daa",
			RomPathPrefix + "acceptance/instr/daa.gb",
			100,
		},
		{
			"if_ie_registers",
			RomPathPrefix + "acceptance/if_ie_registers.gb",
			100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emu := setup(tt.path)
			buf := skipFrame(emu, tt.frame)
			file, err := os.Create(ImagePathPrefix + tt.name + ".png")
			defer file.Close()
			if err != nil {
				panic(err)
			}
			img := image.NewRGBA(image.Rect(0, 0, constants.ScreenWidth, constants.ScreenHeight))
			set(img, buf)
			if err := png.Encode(file, img); err != nil {
				panic(err)
			}
		})
	}
}
