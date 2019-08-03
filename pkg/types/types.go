package types

import "image/color"

// Register is generic register
type Register = byte

// RGB display rgb color
type RGB struct {
	Red   byte
	Green byte
	Blue  byte
}

// Word for gb
type Word uint16

type ImageData = []color.RGBA

type Bit int

const (
	Bit0 Bit = 0x01
	Bit1 Bit = 0x02
	Bit2 Bit = 0x04
	Bit3 Bit = 0x08
	Bit4 Bit = 0x10
	Bit5 Bit = 0x20
	Bit6 Bit = 0x40
	Bit7 Bit = 0x80
)
