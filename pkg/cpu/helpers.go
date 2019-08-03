package cpu

import (
	"github.com/bokuweb/gopher-boy/pkg/types"
	"github.com/bokuweb/gopher-boy/pkg/utils"
)

func (cpu *CPU) inc(value byte) byte {
	incremented := value + 0x01
	cpu.clearFlag(N)
	if incremented == 0 {
		cpu.setFlag(Z)
	} else {
		cpu.clearFlag(Z)
	}
	if (incremented^0x01^value)&0x10 == 0x10 {
		cpu.setFlag(H)
	} else {
		cpu.clearFlag(H)
	}
	return incremented
}

func (cpu *CPU) dec(value byte) byte {
	decremented := value - 1
	cpu.setFlag(N)
	cpu.applyZeroBy(decremented)
	if (decremented^0x01^value)&0x10 == 0x10 {
		cpu.setFlag(H)
	} else {
		cpu.clearFlag(H)
	}
	return decremented
}

func (cpu *CPU) clearFlag(flag flags) {
	switch flag {
	case Z:
		cpu.Regs.F = cpu.clearBit(7, cpu.Regs.F)
	case N:
		cpu.Regs.F = cpu.clearBit(6, cpu.Regs.F)
	case H:
		cpu.Regs.F = cpu.clearBit(5, cpu.Regs.F)
	case C:
		cpu.Regs.F = cpu.clearBit(4, cpu.Regs.F)
	}
	cpu.Regs.F &= 0xF0
}

func (cpu *CPU) setFlag(flag flags) {
	switch flag {
	case Z:
		cpu.Regs.F = cpu.setBit(7, cpu.Regs.F)
	case N:
		cpu.Regs.F = cpu.setBit(6, cpu.Regs.F)
	case H:
		cpu.Regs.F = cpu.setBit(5, cpu.Regs.F)
	case C:
		cpu.Regs.F = cpu.setBit(4, cpu.Regs.F)
	}
	cpu.Regs.F &= 0xF0
}

func (cpu *CPU) isSet(flag flags) bool {
	switch flag {
	case Z:
		return cpu.Regs.F&0x80 != 0
	case N:
		return cpu.Regs.F&0x40 != 0
	case H:
		return cpu.Regs.F&0x20 != 0
	case C:
		return cpu.Regs.F&0x10 != 0
	}
	return false
}

func (cpu *CPU) setBit(bit byte, a byte) byte {
	return a | (1 << uint(bit))
}

func (cpu *CPU) clearBit(bit byte, a byte) byte {
	return a & ^(1 << uint(bit))
}

func (cpu *CPU) addBytes(a, b byte) byte {
	computed := a + b
	cpu.clearFlag(N)
	if computed == 0x00 {
		cpu.setFlag(Z)
	} else {
		cpu.clearFlag(Z)
	}
	if (computed^b^a)&0x10 == 0x10 {
		cpu.setFlag(H)
	} else {
		cpu.clearFlag(H)
	}
	if computed < a {
		cpu.setFlag(C)
	} else {
		cpu.clearFlag(C)
	}
	return computed
}

func (cpu *CPU) addWords(a, b types.Word) types.Word {
	computed := a + b
	if computed < a {
		cpu.setFlag(C)
	} else {
		cpu.clearFlag(C)
	}
	if (computed^b^a)&0x1000 == 0x1000 {
		cpu.setFlag(H)
	} else {
		cpu.clearFlag(H)
	}

	cpu.clearFlag(N)

	return computed
}

func (cpu *CPU) subBytes(a, b byte) byte {
	cpu.setFlag(N)
	if a&0xF < b&0xF {
		cpu.setFlag(H)
	} else {
		cpu.clearFlag(H)
	}
	if a&0xFF < b&0xFF {
		cpu.setFlag(C)
	} else {
		cpu.clearFlag(C)
	}
	computed := a - b
	if computed == 0x00 {
		cpu.setFlag(Z)
	} else {
		cpu.clearFlag(Z)
	}
	return computed
}

func (cpu *CPU) getHL() types.Word {
	return utils.Bytes2Word(cpu.Regs.H, cpu.Regs.L)
}

func (cpu *CPU) toHLRegs(v types.Word) {
	cpu.Regs.H, cpu.Regs.L = utils.Word2Bytes(v)
}

func (cpu *CPU) and(a, b byte) byte {
	cpu.setFlag(H)
	cpu.clearFlag(N)
	cpu.clearFlag(C)
	computed := a & b
	if computed == 0x00 {
		cpu.setFlag(Z)
	} else {
		cpu.clearFlag(Z)
	}
	return computed
}

func (cpu *CPU) xor(a, b byte) byte {
	cpu.clearFlag(H)
	cpu.clearFlag(N)
	cpu.clearFlag(C)
	computed := a ^ b
	if computed == 0x00 {
		cpu.setFlag(Z)
	} else {
		cpu.clearFlag(Z)
	}
	return computed
}

func (cpu *CPU) or(a, b byte) byte {
	cpu.clearFlag(H)
	cpu.clearFlag(N)
	cpu.clearFlag(C)
	computed := a | b
	if computed == 0x00 {
		cpu.setFlag(Z)
	} else {
		cpu.clearFlag(Z)
	}
	return computed
}

func (cpu *CPU) pop() byte {
	b := cpu.bus.ReadByte(cpu.SP)
	cpu.SP++
	return b
}

func (cpu *CPU) push(v byte) {
	cpu.SP--
	cpu.bus.WriteByte(cpu.SP, v)
}

func (cpu *CPU) pushPC() {
	upper, lower := utils.Word2Bytes(cpu.PC)
	cpu.push(upper)
	cpu.push(lower)
}

func (cpu *CPU) pop2PC() {
	lower := cpu.pop()
	upper := cpu.pop()
	cpu.PC = utils.Bytes2Word(upper, lower)
}

func (cpu *CPU) applyZeroBy(computed byte) {
	if computed == 0 {
		cpu.setFlag(Z)
	} else {
		cpu.clearFlag(Z)
	}
}
