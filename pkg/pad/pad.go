package pad

// This is the matrix layout for register $FF00:
//              P14      P15
//               |        |
//  P10-------O-Right----O-A
//               |        |
//  P11-------O-Left-----O-B
//               |        |
//  P12-------O-Up-------O-Select
//              |         |
//  P13-------O-Down-----O-Start
//

type Pad struct {
	// Bit 7 - Not used
	// Bit 6 - Not used
	// Bit 5 - P15 out port
	// Bit 4 - P14 out port
	// Bit 3 - P13 in port
	// Bit 2 - P12 in port
	// Bit 1 - P11 in port
	// Bit 0 - P10 in port
	reg   byte
	state Button
}

type Button byte

const (
	// A is the A button on the GameBoy.
	A Button = 0x01
	// B is the B button on the GameBoy.
	B Button = 0x02
	// Select is the select button on the GameBoy.
	Select Button = 0x04
	// Start is the start button on the GameBoy.
	Start Button = 0x08
	// Right is the right pad direction on the GameBoy.
	Right Button = 0x10
	// Left is the left pad direction on the GameBoy.
	Left Button = 0x20
	// Up is the up pad direction on the GameBoy.
	Up Button = 0x40
	// Down is the down pad direction on the GameBoy.
	Down Button = 0x80
)

// NewPad constructs pad peripheral.
func NewPad() *Pad {
	return &Pad{
		reg: 0x3F,
	}
}

func (pad *Pad) Read() byte {
	if pad.isP14On() {
		return pad.reg & ^byte(pad.state>>4)
	}
	if pad.isP15On() {
		return pad.reg & ^byte(pad.state&0x0F)
	}
	return pad.reg | 0x0f
}

func (pad *Pad) Write(data byte) {
	pad.reg = (pad.reg & 0xCF) | (data & 0x30)
}

func (pad *Pad) isP14On() bool {
	return pad.reg&0x10 == 0
}

func (pad *Pad) isP15On() bool {
	return pad.reg&0x20 == 0
}

func (pad *Pad) Press(button Button) {
	pad.state |= button
}

func (pad *Pad) Release(button Button) {
	pad.state &= ^button
}
