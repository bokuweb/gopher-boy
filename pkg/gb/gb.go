package gb

import (
	"image/color"
	"time"

	"github.com/bokuweb/gopher-boy/pkg/constants"
	"github.com/bokuweb/gopher-boy/pkg/cpu"
	"github.com/bokuweb/gopher-boy/pkg/gpu"
	"github.com/bokuweb/gopher-boy/pkg/interfaces/window"
	"github.com/bokuweb/gopher-boy/pkg/interrupt"
	"github.com/bokuweb/gopher-boy/pkg/timer"
)

// CyclesPerFrame is cpu clock num for 1frame.
const CyclesPerFrame = 70224

// GB is gameboy emulator struct
type GB struct {
	currentCycle uint
	cpu          *cpu.CPU
	gpu          *gpu.GPU
	timer        *timer.Timer
	irq          *interrupt.Interrupt
	win          window.Window
}

// NewGB is gb initializer
func NewGB(cpu *cpu.CPU, gpu *gpu.GPU, timer *timer.Timer, irq *interrupt.Interrupt, win window.Window) *GB {
	return &GB{
		currentCycle: 0,
		cpu:          cpu,
		gpu:          gpu,
		timer:        timer,
		irq:          irq,
		win:          win,
	}
}

// Start is
func (g *GB) Start() {
	t := time.NewTicker(16 * time.Millisecond)
	for {
		select {
		case <-t.C:
			buf := g.Next()
			imgData := make([]color.RGBA, constants.ScreenWidth*constants.ScreenHeight)
			i := 0
			for i*4 < len(buf) {
				y := constants.ScreenHeight - (i / constants.ScreenWidth) - 1
				imgData[y*constants.ScreenWidth+i%constants.ScreenWidth] = color.RGBA{buf[i*4], buf[i*4+1], buf[i*4+2], buf[i*4+3]}
				i++
			}
			g.win.Render(imgData)
		}
	}
	t.Stop()
}
func (g *GB) Next() []byte {
	for {
		var cycles uint
		if g.gpu.DMAStarted() {
			g.gpu.Transfer()
			// https://github.com/Gekkio/mooneye-gb/blob/master/docs/accuracy.markdown#how-many-cycles-does-oam-dma-take
			cycles = 162
		} else {
			cycles = g.cpu.Step()
		}
		g.gpu.Step(cycles * 4)
		if overflowed := g.timer.Update(cycles); overflowed {
			g.irq.SetIRQ(interrupt.TimerOverflowFlag)
		}
		g.currentCycle += cycles * 4
		if g.currentCycle >= CyclesPerFrame {
			g.win.PollKey()
			g.currentCycle -= CyclesPerFrame
			return g.gpu.GetImageData()
		}
	}
}
