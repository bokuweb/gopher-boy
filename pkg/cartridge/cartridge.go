package cartridge

import (
	"fmt"
	"strings"

	"github.com/bokuweb/gopher-boy/pkg/types"
)

// Cartridge is GameBoy cartridge
type Cartridge struct {
	mbc     MBC
	Title   string
	ROM     []byte
	RAMSize int
}

/*
  CartridgeType is
  0x00: ROM ONLY
  0x01: ROM+MBC1
  0x02: ROM+MBC1+RAM
  0x03: ROM+MBC1+RAM+BATT
  0x05: ROM+MBC2
  0x06: ROM+MBC2+BATTERY
  0x08: ROM+RAM
  0x09: ROM+RAM+BATTERY
  0x0B: ROM+MMM01
  0x0C: ROM+MMM01+SRAM
  0x0D: ROM+MMM01+SRAM+BATT
  0x12: ROM+MBC3+RAM
  0x13: ROM+MBC3+RAM+BATT
  0x19: ROM+MBC5
  0x1A: ROM+MBC5+RAM
  0x1B: ROM+MBC5+RAM+BATT
  0x1C: ROM+MBC5+RUMBLE
  0x1D: ROM+MBC5+RUMBLE+SRAM
  0x1E: ROM+MBC5+RUMBLE+SRAM+BATT
  0x1F: Pocket Camera
  0xFD: Bandai TAMA5
  0xFE: Hudson HuC-3
*/
type CartridgeType byte

const (
	MBC_0                 CartridgeType = 0x00
	MBC_1                               = 0x01
	MBC_1_RAM                           = 0x02
	MBC_1_RAM_BATT                      = 0x03
	MBC_3_RAM_BATT                      = 0x13
	MBC_3_RAM_BATT_RTC                  = 0x10
	MBC_5                               = 0x19
	MBC_5_RAM                           = 0x1A
	MBC_5_RAM_BATT                      = 0x1B
	MBC_5_RUMBLE                        = 0x1C
	MBC_5_RAM_RUMBLE                    = 0x1D
	MBC_5_RAM_BATT_RUMBLE               = 0x1E
)

// NewCartridge is cartridge constructure
func NewCartridge(buf []byte) (*Cartridge, error) {
	title := strings.TrimSpace(string(buf[0x0134:0x0142]))
	// romSize := 0x8000 << buf[0x0148]
	ramSize := getRAMSize(buf[0x0149])
	cartridgeType := CartridgeType(buf[0x0147])
	fmt.Println("cartridge type is ", cartridgeType)
	var mbc MBC
	switch cartridgeType {
	case MBC_0:
		mbc = NewMBC0(buf[0x0000:0x8000])
	case MBC_1:
		mbc = NewMBC1(buf, ramSize, false)
	case MBC_1_RAM:
		mbc = NewMBC1(buf, ramSize, false)
	case MBC_1_RAM_BATT:
		mbc = NewMBC1(buf, ramSize, true)
	}

	return &Cartridge{
		mbc:     mbc,
		Title:   title,
		ROM:     buf,
		RAMSize: ramSize,
	}, nil
}

//RAM size:
// 0 - None
// 1 - 16kBit = 2kB = 1 bank
// 2 - 64kBit = 8kB = 1 bank
// 3 - 256kBit = 32kB = 4 banks
// 4 - 1MBit =128kB =16 banks
func getRAMSize(size byte) int {
	switch size {
	case 0x00:
		return 0
	case 0x01:
		return 2 * 1024
	case 0x02:
		return 8 * 1024
	case 0x03:
		return 32 * 1024
	case 0x04:
		return 128 * 1024
	}
	return 0
}

func (c *Cartridge) ReadByte(addr types.Word) byte {
	return c.mbc.Read(addr)
}

func (c *Cartridge) WriteByte(addr types.Word, data byte) {
	c.mbc.Write(addr, data)
}
