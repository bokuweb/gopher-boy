package main

import (
	"errors"
	"log"
	"os"

	"github.com/bokuweb/gopher-boy/pkg/interrupt"
	"github.com/bokuweb/gopher-boy/pkg/pad"

	"github.com/bokuweb/gopher-boy/pkg/gpu"
	"github.com/bokuweb/gopher-boy/pkg/timer"
	"github.com/bokuweb/gopher-boy/pkg/utils"

	"github.com/bokuweb/gopher-boy/pkg/cpu"
	"github.com/bokuweb/gopher-boy/pkg/gb"
	"github.com/bokuweb/gopher-boy/pkg/logger"
	"github.com/bokuweb/gopher-boy/pkg/ram"

	"github.com/bokuweb/gopher-boy/pkg/bus"
	"github.com/bokuweb/gopher-boy/pkg/cartridge"
	"github.com/bokuweb/gopher-boy/pkg/window"
)

func main() {
	level := "Debug"
	if os.Getenv("LEVEL") != "" {
		level = os.Getenv("LEVEL")
	}
	l := logger.NewLogger(logger.LogLevel(level))
	if len(os.Args) != 2 {
		log.Fatalf("ERROR: %v", errors.New("Please specify the ROM"))
	}
	file := os.Args[1]
	log.Println(file)
	buf, err := utils.LoadROM(file)
	if err != nil {
		log.Fatalf("ERROR: %v", errors.New("Failed to load ROM"))
	}
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
	win.Run(func() {
		win.Init()
		emu.Start()
	})
}
