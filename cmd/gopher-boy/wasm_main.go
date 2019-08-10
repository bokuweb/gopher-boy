// +build wasm

package main

import (
	"errors"
	// "image/color"
	"log"
	"syscall/js"
 
	"github.com/bokuweb/gopher-boy/pkg/interrupt"
	"github.com/bokuweb/gopher-boy/pkg/logger"
	"github.com/bokuweb/gopher-boy/pkg/pad"
	"github.com/bokuweb/gopher-boy/pkg/window"
 
	"github.com/bokuweb/gopher-boy/pkg/gpu"
	"github.com/bokuweb/gopher-boy/pkg/timer"
 
	"github.com/bokuweb/gopher-boy/pkg/cpu"
	"github.com/bokuweb/gopher-boy/pkg/gb"
	"github.com/bokuweb/gopher-boy/pkg/ram"

	"github.com/bokuweb/gopher-boy/pkg/bus"
	"github.com/bokuweb/gopher-boy/pkg/cartridge"
)

func newGB(this js.Value, args []js.Value) interface{} {
	buf := []byte{}
	for i := 0; i < args[0].Get("length").Int(); i++ {
		buf = append(buf, byte(args[0].Index(i).Int()))
	}
	l := logger.NewLogger(logger.LogLevel("INFO"))
	cart, err := cartridge.NewCartridge(buf)
	if err != nil {
		log.Fatalf("ERROR: %v", errors.New("Failed to create cartridge"))
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

	win := window.NewWindow(pad)
	emu := gb.NewGB(cpu.NewCPU(l, b, irq), gpu, t, irq, win)

	this.Set("next", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		img := emu.Next()
		// buf := []byte{}
		// for _, color := range img {
		// 	buf = append(buf, color.R)
		// 	buf = append(buf, color.G)
		// 	buf = append(buf, color.B)
		// 	buf = append(buf, 255)
		// }
		return js.TypedArrayOf(img)
	}))
	return this
}

func main() {
	 w := js.Global()

	 w.Set("GB", js.FuncOf(newGB))
	 select {}
}
