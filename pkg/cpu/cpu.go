package cpu

import (
	"github.com/bokuweb/gopher-boy/pkg/interfaces/bus"
	"github.com/bokuweb/gopher-boy/pkg/interfaces/interrupt"
	"github.com/bokuweb/gopher-boy/pkg/interfaces/logger"
	"github.com/bokuweb/gopher-boy/pkg/types"
	"github.com/bokuweb/gopher-boy/pkg/utils"
)

// Registers is generic cpu registers
type Registers struct {
	A types.Register
	B types.Register
	C types.Register
	D types.Register
	E types.Register
	H types.Register
	L types.Register
	F types.Register
}

type flags int

const (
	// C is carry flag
	C flags = iota + 1
	// H is half carry flag
	H
	// N is negative flag
	N
	// Z is zero flag
	Z
)

// CPU is cpu state struct
type CPU struct {
	logger  logger.Logger
	PC      types.Word
	SP      types.Word
	Regs    Registers
	bus     bus.Accessor
	irq     interrupt.Interrupt
	stopped bool
	halted  bool
}

type Cycle = uint

// NewCPU is CPU constructor
func NewCPU(logger logger.Logger, bus bus.Accessor, irq interrupt.Interrupt) *CPU {
	cpu := &CPU{
		logger: logger,
		PC:     0x100, // INFO: Skip
		SP:     0xFFFE,
		Regs: Registers{
			A: 0x11,
			B: 0x00,
			C: 0x00,
			D: 0xFF,
			E: 0x56,
			F: 0x80,
			H: 0x00,
			L: 0x0D,
		},
		bus:     bus,
		irq:     irq,
		stopped: false,
		halted:  false,
	}
	return cpu
}

func (cpu *CPU) fetch() byte {
	d := cpu.bus.ReadByte(cpu.PC)
	cpu.PC++
	return d
}

// Step execute an instruction
func (cpu *CPU) Step() Cycle {

	if cpu.halted {
		if cpu.irq.HasIRQ() {
			cpu.halted = false
		}
		return 0x01
	}
	if hasIRQ := cpu.resolveIRQ(); hasIRQ {
		return 0x01
	}
	opcode := cpu.fetch()
	var inst *inst
	if opcode == 0xCB {
		next := cpu.fetch()
		inst = cbPrefixedInstructions[next]
	} else {
		inst = instructions[opcode]
	}

	operands := cpu.fetchOperands(inst.OperandsSize)
	inst.Execute(cpu, operands)
	return inst.Cycles
}

func (cpu *CPU) fetchOperands(size uint) []byte {
	operands := []byte{}
	switch size {
	case 1:
		operands = append(operands, cpu.fetch())
	case 2:
		operands = append(operands, cpu.fetch())
		operands = append(operands, cpu.fetch())
	}
	return operands
}

type inst struct {
	Opcode       byte
	Description  string
	OperandsSize uint
	Cycles       uint
	Execute      func(cpu *CPU, operands []byte)
}

// EMPTY is empty instruction
var EMPTY = &inst{0xFF, "EMPTY", 0, 1, func(cpu *CPU, operands []byte) {
}}

var cbPrefixedInstructions = []*inst{
	&inst{0x0, "RLC B", 0, 2, func(cpu *CPU, operands []byte) { cpu.rlc_n(&cpu.Regs.B) }},
	&inst{0x1, "RLC C", 0, 2, func(cpu *CPU, operands []byte) { cpu.rlc_n(&cpu.Regs.C) }},
	&inst{0x2, "RLC D", 0, 2, func(cpu *CPU, operands []byte) { cpu.rlc_n(&cpu.Regs.D) }},
	&inst{0x3, "RLC E", 0, 2, func(cpu *CPU, operands []byte) { cpu.rlc_n(&cpu.Regs.E) }},
	&inst{0x4, "RLC H", 0, 2, func(cpu *CPU, operands []byte) { cpu.rlc_n(&cpu.Regs.H) }},
	&inst{0x5, "RLC L", 0, 2, func(cpu *CPU, operands []byte) { cpu.rlc_n(&cpu.Regs.L) }},
	&inst{0x6, "RLC (HL)", 0, 4, func(cpu *CPU, operands []byte) { cpu.rlc_hl() }},
	&inst{0x7, "RLC A", 0, 2, func(cpu *CPU, operands []byte) { cpu.rlc_n(&cpu.Regs.A) }},
	&inst{0x8, "RRC B", 0, 2, func(cpu *CPU, operands []byte) { cpu.rrc_n(&cpu.Regs.B) }},
	&inst{0x9, "RRC C", 0, 2, func(cpu *CPU, operands []byte) { cpu.rrc_n(&cpu.Regs.C) }},
	&inst{0xA, "RRC D", 0, 2, func(cpu *CPU, operands []byte) { cpu.rrc_n(&cpu.Regs.D) }},
	&inst{0xB, "RRC E", 0, 2, func(cpu *CPU, operands []byte) { cpu.rrc_n(&cpu.Regs.E) }},
	&inst{0xC, "RRC H", 0, 2, func(cpu *CPU, operands []byte) { cpu.rrc_n(&cpu.Regs.H) }},
	&inst{0xD, "RRC L", 0, 2, func(cpu *CPU, operands []byte) { cpu.rrc_n(&cpu.Regs.L) }},
	&inst{0xE, "RRC (HL)", 0, 4, func(cpu *CPU, operands []byte) { cpu.rrc_hl() }},
	&inst{0xF, "RRC A", 0, 2, func(cpu *CPU, operands []byte) { cpu.rrc_n(&cpu.Regs.A) }},
	&inst{0x10, "RL B", 0, 2, func(cpu *CPU, operands []byte) { cpu.rl_n(&cpu.Regs.B) }},
	&inst{0x11, "RL C", 0, 2, func(cpu *CPU, operands []byte) { cpu.rl_n(&cpu.Regs.C) }},
	&inst{0x12, "RL D", 0, 2, func(cpu *CPU, operands []byte) { cpu.rl_n(&cpu.Regs.D) }},
	&inst{0x13, "RL E", 0, 2, func(cpu *CPU, operands []byte) { cpu.rl_n(&cpu.Regs.E) }},
	&inst{0x14, "RL H", 0, 2, func(cpu *CPU, operands []byte) { cpu.rl_n(&cpu.Regs.H) }},
	&inst{0x15, "RL L", 0, 2, func(cpu *CPU, operands []byte) { cpu.rl_n(&cpu.Regs.L) }},
	&inst{0x16, "RL (HL)", 0, 4, func(cpu *CPU, operands []byte) { cpu.rl_hl() }},
	&inst{0x17, "RL A", 0, 2, func(cpu *CPU, operands []byte) { cpu.rl_n(&cpu.Regs.A) }},
	&inst{0x18, "RR B", 0, 2, func(cpu *CPU, operands []byte) { cpu.rr_n(&cpu.Regs.B) }},
	&inst{0x19, "RR C", 0, 2, func(cpu *CPU, operands []byte) { cpu.rr_n(&cpu.Regs.C) }},
	&inst{0x1A, "RR D", 0, 2, func(cpu *CPU, operands []byte) { cpu.rr_n(&cpu.Regs.D) }},
	&inst{0x1B, "RR E", 0, 2, func(cpu *CPU, operands []byte) { cpu.rr_n(&cpu.Regs.E) }},
	&inst{0x1C, "RR H", 0, 2, func(cpu *CPU, operands []byte) { cpu.rr_n(&cpu.Regs.H) }},
	&inst{0x1D, "RR L", 0, 2, func(cpu *CPU, operands []byte) { cpu.rr_n(&cpu.Regs.L) }},
	&inst{0x1E, "RR (HL)", 0, 4, func(cpu *CPU, operands []byte) { cpu.rr_hl() }},
	&inst{0x1F, "RR A", 0, 2, func(cpu *CPU, operands []byte) { cpu.rr_n(&cpu.Regs.A) }},
	&inst{0x20, "SLA B", 0, 2, func(cpu *CPU, operands []byte) { cpu.sla_n(&cpu.Regs.B) }},
	&inst{0x21, "SLA C", 0, 2, func(cpu *CPU, operands []byte) { cpu.sla_n(&cpu.Regs.C) }},
	&inst{0x22, "SLA D", 0, 2, func(cpu *CPU, operands []byte) { cpu.sla_n(&cpu.Regs.D) }},
	&inst{0x23, "SLA E", 0, 2, func(cpu *CPU, operands []byte) { cpu.sla_n(&cpu.Regs.E) }},
	&inst{0x24, "SLA H", 0, 2, func(cpu *CPU, operands []byte) { cpu.sla_n(&cpu.Regs.H) }},
	&inst{0x25, "SLA L", 0, 2, func(cpu *CPU, operands []byte) { cpu.sla_n(&cpu.Regs.L) }},
	&inst{0x26, "SLA (HL)", 0, 4, func(cpu *CPU, operands []byte) { cpu.sla_hl() }},
	&inst{0x27, "SLA A", 0, 2, func(cpu *CPU, operands []byte) { cpu.sla_n(&cpu.Regs.A) }},
	&inst{0x28, "SRA B", 0, 2, func(cpu *CPU, operands []byte) { cpu.sra_n(&cpu.Regs.B) }},
	&inst{0x29, "SRA C", 0, 2, func(cpu *CPU, operands []byte) { cpu.sra_n(&cpu.Regs.C) }},
	&inst{0x2A, "SRA D", 0, 2, func(cpu *CPU, operands []byte) { cpu.sra_n(&cpu.Regs.D) }},
	&inst{0x2B, "SRA E", 0, 2, func(cpu *CPU, operands []byte) { cpu.sra_n(&cpu.Regs.E) }},
	&inst{0x2C, "SRA H", 0, 2, func(cpu *CPU, operands []byte) { cpu.sra_n(&cpu.Regs.H) }},
	&inst{0x2D, "SRA L", 0, 2, func(cpu *CPU, operands []byte) { cpu.sra_n(&cpu.Regs.L) }},
	&inst{0x2E, "SRA (HL)", 0, 4, func(cpu *CPU, operands []byte) { cpu.sra_hl() }},
	&inst{0x2F, "SRA A", 0, 2, func(cpu *CPU, operands []byte) { cpu.sra_n(&cpu.Regs.A) }},
	&inst{0x30, "SWAP B", 0, 2, func(cpu *CPU, operands []byte) { cpu.swap_n(&cpu.Regs.B) }},
	&inst{0x31, "SWAP C", 0, 2, func(cpu *CPU, operands []byte) { cpu.swap_n(&cpu.Regs.C) }},
	&inst{0x32, "SWAP D", 0, 2, func(cpu *CPU, operands []byte) { cpu.swap_n(&cpu.Regs.D) }},
	&inst{0x33, "SWAP E", 0, 2, func(cpu *CPU, operands []byte) { cpu.swap_n(&cpu.Regs.E) }},
	&inst{0x34, "SWAP H", 0, 2, func(cpu *CPU, operands []byte) { cpu.swap_n(&cpu.Regs.H) }},
	&inst{0x35, "SWAP L", 0, 2, func(cpu *CPU, operands []byte) { cpu.swap_n(&cpu.Regs.L) }},
	&inst{0x36, "SWAP (HL)", 0, 4, func(cpu *CPU, operands []byte) { cpu.swap_hl() }},
	&inst{0x37, "SWAP A", 0, 2, func(cpu *CPU, operands []byte) { cpu.swap_n(&cpu.Regs.A) }},
	&inst{0x38, "SRL B", 0, 2, func(cpu *CPU, operands []byte) { cpu.srl_n(&cpu.Regs.B) }},
	&inst{0x39, "SRL C", 0, 2, func(cpu *CPU, operands []byte) { cpu.srl_n(&cpu.Regs.C) }},
	&inst{0x3A, "SRL D", 0, 2, func(cpu *CPU, operands []byte) { cpu.srl_n(&cpu.Regs.D) }},
	&inst{0x3B, "SRL E", 0, 2, func(cpu *CPU, operands []byte) { cpu.srl_n(&cpu.Regs.E) }},
	&inst{0x3C, "SRL H", 0, 2, func(cpu *CPU, operands []byte) { cpu.srl_n(&cpu.Regs.H) }},
	&inst{0x3D, "SRL L", 0, 2, func(cpu *CPU, operands []byte) { cpu.srl_n(&cpu.Regs.L) }},
	&inst{0x3E, "SRL (HL)", 0, 4, func(cpu *CPU, operands []byte) { cpu.srl_hl() }},
	&inst{0x3F, "SRL A", 0, 2, func(cpu *CPU, operands []byte) { cpu.srl_n(&cpu.Regs.A) }},
	&inst{0x40, "BIT 0,B", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit0, &cpu.Regs.B) }},
	&inst{0x41, "BIT 0,C", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit0, &cpu.Regs.C) }},
	&inst{0x42, "BIT 0,D", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit0, &cpu.Regs.D) }},
	&inst{0x43, "BIT 0,E", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit0, &cpu.Regs.E) }},
	&inst{0x44, "BIT 0,H", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit0, &cpu.Regs.H) }},
	&inst{0x45, "BIT 0,L", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit0, &cpu.Regs.L) }},
	&inst{0x46, "BIT 0,(HL)", 0, 3, func(cpu *CPU, operands []byte) { cpu.bit_b_hl(types.Bit0) }},
	&inst{0x47, "BIT 0,A", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit0, &cpu.Regs.A) }},
	&inst{0x48, "BIT 1,B", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit1, &cpu.Regs.B) }},
	&inst{0x49, "BIT 1,C", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit1, &cpu.Regs.C) }},
	&inst{0x4A, "BIT 1,D", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit1, &cpu.Regs.D) }},
	&inst{0x4B, "BIT 1,E", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit1, &cpu.Regs.E) }},
	&inst{0x4C, "BIT 1,H", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit1, &cpu.Regs.H) }},
	&inst{0x4D, "BIT 1,L", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit1, &cpu.Regs.L) }},
	&inst{0x4E, "BIT 1,(HL)", 0, 3, func(cpu *CPU, operands []byte) { cpu.bit_b_hl(types.Bit1) }},
	&inst{0x4F, "BIT 1,A", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit1, &cpu.Regs.A) }},
	&inst{0x50, "BIT 2,B", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit2, &cpu.Regs.B) }},
	&inst{0x51, "BIT 2,C", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit2, &cpu.Regs.C) }},
	&inst{0x52, "BIT 2,D", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit2, &cpu.Regs.D) }},
	&inst{0x53, "BIT 2,E", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit2, &cpu.Regs.E) }},
	&inst{0x54, "BIT 2,H", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit2, &cpu.Regs.H) }},
	&inst{0x55, "BIT 2,L", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit2, &cpu.Regs.L) }},
	&inst{0x56, "BIT 2,(HL)", 0, 3, func(cpu *CPU, operands []byte) { cpu.bit_b_hl(types.Bit2) }},
	&inst{0x57, "BIT 2,A", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit2, &cpu.Regs.A) }},
	&inst{0x58, "BIT 3,B", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit3, &cpu.Regs.B) }},
	&inst{0x59, "BIT 3,C", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit3, &cpu.Regs.C) }},
	&inst{0x5A, "BIT 3,D", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit3, &cpu.Regs.D) }},
	&inst{0x5B, "BIT 3,E", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit3, &cpu.Regs.E) }},
	&inst{0x5C, "BIT 3,H", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit3, &cpu.Regs.H) }},
	&inst{0x5D, "BIT 3,L", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit3, &cpu.Regs.L) }},
	&inst{0x5E, "BIT 3,(HL)", 0, 3, func(cpu *CPU, operands []byte) { cpu.bit_b_hl(types.Bit3) }},
	&inst{0x5F, "BIT 3,A", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit3, &cpu.Regs.A) }},
	&inst{0x60, "BIT 4,B", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit4, &cpu.Regs.B) }},
	&inst{0x61, "BIT 4,C", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit4, &cpu.Regs.C) }},
	&inst{0x62, "BIT 4,D", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit4, &cpu.Regs.D) }},
	&inst{0x63, "BIT 4,E", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit4, &cpu.Regs.E) }},
	&inst{0x64, "BIT 4,H", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit4, &cpu.Regs.H) }},
	&inst{0x65, "BIT 4,L", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit4, &cpu.Regs.L) }},
	&inst{0x66, "BIT 4,(HL)", 0, 3, func(cpu *CPU, operands []byte) { cpu.bit_b_hl(types.Bit4) }},
	&inst{0x67, "BIT 4,A", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit4, &cpu.Regs.A) }},
	&inst{0x68, "BIT 5,B", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit5, &cpu.Regs.B) }},
	&inst{0x69, "BIT 5,C", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit5, &cpu.Regs.C) }},
	&inst{0x6A, "BIT 5,D", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit5, &cpu.Regs.D) }},
	&inst{0x6B, "BIT 5,E", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit5, &cpu.Regs.E) }},
	&inst{0x6C, "BIT 5,H", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit5, &cpu.Regs.H) }},
	&inst{0x6D, "BIT 5,L", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit5, &cpu.Regs.L) }},
	&inst{0x6E, "BIT 5,(HL)", 0, 3, func(cpu *CPU, operands []byte) { cpu.bit_b_hl(types.Bit5) }},
	&inst{0x6F, "BIT 5,A", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit5, &cpu.Regs.A) }},
	&inst{0x70, "BIT 6,B", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit6, &cpu.Regs.B) }},
	&inst{0x71, "BIT 6,C", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit6, &cpu.Regs.C) }},
	&inst{0x72, "BIT 6,D", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit6, &cpu.Regs.D) }},
	&inst{0x73, "BIT 6,E", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit6, &cpu.Regs.E) }},
	&inst{0x74, "BIT 6,H", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit6, &cpu.Regs.H) }},
	&inst{0x75, "BIT 6,L", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit6, &cpu.Regs.L) }},
	&inst{0x76, "BIT 6,(HL)", 0, 3, func(cpu *CPU, operands []byte) { cpu.bit_b_hl(types.Bit6) }},
	&inst{0x77, "BIT 6,A", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit6, &cpu.Regs.A) }},
	&inst{0x78, "BIT 7,B", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit7, &cpu.Regs.B) }},
	&inst{0x79, "BIT 7,C", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit7, &cpu.Regs.C) }},
	&inst{0x7A, "BIT 7,D", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit7, &cpu.Regs.D) }},
	&inst{0x7B, "BIT 7,E", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit7, &cpu.Regs.E) }},
	&inst{0x7C, "BIT 7,H", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit7, &cpu.Regs.H) }},
	&inst{0x7D, "BIT 7,L", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit7, &cpu.Regs.L) }},
	&inst{0x7E, "BIT 7,(HL)", 0, 3, func(cpu *CPU, operands []byte) { cpu.bit_b_hl(types.Bit7) }},
	&inst{0x7F, "BIT 7,A", 0, 2, func(cpu *CPU, operands []byte) { cpu.bit_b_r(types.Bit7, &cpu.Regs.A) }},
	&inst{0x80, "RES 0,B", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit0, &cpu.Regs.B) }},
	&inst{0x81, "RES 0,C", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit0, &cpu.Regs.C) }},
	&inst{0x82, "RES 0,D", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit0, &cpu.Regs.D) }},
	&inst{0x83, "RES 0,E", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit0, &cpu.Regs.E) }},
	&inst{0x84, "RES 0,H", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit0, &cpu.Regs.H) }},
	&inst{0x85, "RES 0,L", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit0, &cpu.Regs.L) }},
	&inst{0x86, "RES 0,(HL)", 0, 4, func(cpu *CPU, operands []byte) { cpu.res_b_hl(types.Bit0) }},
	&inst{0x87, "RES 0,A", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit0, &cpu.Regs.A) }},
	&inst{0x88, "RES 1,B", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit1, &cpu.Regs.B) }},
	&inst{0x89, "RES 1,C", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit1, &cpu.Regs.C) }},
	&inst{0x8A, "RES 1,D", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit1, &cpu.Regs.D) }},
	&inst{0x8B, "RES 1,E", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit1, &cpu.Regs.E) }},
	&inst{0x8C, "RES 1,H", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit1, &cpu.Regs.H) }},
	&inst{0x8D, "RES 1,L", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit1, &cpu.Regs.L) }},
	&inst{0x8E, "RES 1,(HL)", 0, 4, func(cpu *CPU, operands []byte) { cpu.res_b_hl(types.Bit1) }},
	&inst{0x8F, "RES 1,A", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit1, &cpu.Regs.A) }},
	&inst{0x90, "RES 2,B", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit2, &cpu.Regs.B) }},
	&inst{0x91, "RES 2,C", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit2, &cpu.Regs.C) }},
	&inst{0x92, "RES 2,D", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit2, &cpu.Regs.D) }},
	&inst{0x93, "RES 2,E", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit2, &cpu.Regs.E) }},
	&inst{0x94, "RES 2,H", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit2, &cpu.Regs.H) }},
	&inst{0x95, "RES 2,L", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit2, &cpu.Regs.L) }},
	&inst{0x96, "RES 2,(HL)", 0, 4, func(cpu *CPU, operands []byte) { cpu.res_b_hl(types.Bit2) }},
	&inst{0x97, "RES 2,A", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit2, &cpu.Regs.A) }},
	&inst{0x98, "RES 3,B", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit3, &cpu.Regs.B) }},
	&inst{0x99, "RES 3,C", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit3, &cpu.Regs.C) }},
	&inst{0x9A, "RES 3,D", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit3, &cpu.Regs.D) }},
	&inst{0x9B, "RES 3,E", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit3, &cpu.Regs.E) }},
	&inst{0x9C, "RES 3,H", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit3, &cpu.Regs.H) }},
	&inst{0x9D, "RES 3,L", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit3, &cpu.Regs.L) }},
	&inst{0x9E, "RES 3,(HL)", 0, 4, func(cpu *CPU, operands []byte) { cpu.res_b_hl(types.Bit3) }},
	&inst{0x9F, "RES 3,A", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit3, &cpu.Regs.A) }},
	&inst{0xA0, "RES 4,B", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit4, &cpu.Regs.B) }},
	&inst{0xA1, "RES 4,C", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit4, &cpu.Regs.C) }},
	&inst{0xA2, "RES 4,D", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit4, &cpu.Regs.D) }},
	&inst{0xA3, "RES 4,E", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit4, &cpu.Regs.E) }},
	&inst{0xA4, "RES 4,H", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit4, &cpu.Regs.H) }},
	&inst{0xA5, "RES 4,L", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit4, &cpu.Regs.L) }},
	&inst{0xA6, "RES 4,(HL)", 0, 4, func(cpu *CPU, operands []byte) { cpu.res_b_hl(types.Bit4) }},
	&inst{0xA7, "RES 4,A", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit4, &cpu.Regs.A) }},
	&inst{0xA8, "RES 5,B", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit5, &cpu.Regs.B) }},
	&inst{0xA9, "RES 5,C", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit5, &cpu.Regs.C) }},
	&inst{0xAA, "RES 5,D", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit5, &cpu.Regs.D) }},
	&inst{0xAB, "RES 5,E", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit5, &cpu.Regs.E) }},
	&inst{0xAC, "RES 5,H", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit5, &cpu.Regs.H) }},
	&inst{0xAD, "RES 5,L", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit5, &cpu.Regs.L) }},
	&inst{0xAE, "RES 5,(HL)", 0, 4, func(cpu *CPU, operands []byte) { cpu.res_b_hl(types.Bit5) }},
	&inst{0xAF, "RES 5,A", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit5, &cpu.Regs.A) }},
	&inst{0xB0, "RES 6,B", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit6, &cpu.Regs.B) }},
	&inst{0xB1, "RES 6,C", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit6, &cpu.Regs.C) }},
	&inst{0xB2, "RES 6,D", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit6, &cpu.Regs.D) }},
	&inst{0xB3, "RES 6,E", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit6, &cpu.Regs.E) }},
	&inst{0xB4, "RES 6,H", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit6, &cpu.Regs.H) }},
	&inst{0xB5, "RES 6,L", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit6, &cpu.Regs.L) }},
	&inst{0xB6, "RES 6,(HL)", 0, 4, func(cpu *CPU, operands []byte) { cpu.res_b_hl(types.Bit6) }},
	&inst{0xB7, "RES 6,A", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit6, &cpu.Regs.A) }},
	&inst{0xB8, "RES 7,B", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit7, &cpu.Regs.B) }},
	&inst{0xB9, "RES 7,C", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit7, &cpu.Regs.C) }},
	&inst{0xBA, "RES 7,D", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit7, &cpu.Regs.D) }},
	&inst{0xBB, "RES 7,E", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit7, &cpu.Regs.E) }},
	&inst{0xBC, "RES 7,H", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit7, &cpu.Regs.H) }},
	&inst{0xBD, "RES 7,L", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit7, &cpu.Regs.L) }},
	&inst{0xBE, "RES 7,(HL)", 0, 4, func(cpu *CPU, operands []byte) { cpu.res_b_hl(types.Bit7) }},
	&inst{0xBF, "RES 7,A", 0, 2, func(cpu *CPU, operands []byte) { cpu.res_b_r(types.Bit7, &cpu.Regs.A) }},
	&inst{0xC0, "SET 0,B", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit0, &cpu.Regs.B) }},
	&inst{0xC1, "SET 0,C", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit0, &cpu.Regs.C) }},
	&inst{0xC2, "SET 0,D", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit0, &cpu.Regs.D) }},
	&inst{0xC3, "SET 0,E", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit0, &cpu.Regs.E) }},
	&inst{0xC4, "SET 0,H", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit0, &cpu.Regs.H) }},
	&inst{0xC5, "SET 0,L", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit0, &cpu.Regs.L) }},
	&inst{0xC6, "SET 0,(HL)", 0, 4, func(cpu *CPU, operands []byte) { cpu.set_b_hl(types.Bit0) }},
	&inst{0xC7, "SET 0,A", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit0, &cpu.Regs.A) }},
	&inst{0xC8, "SET 1,B", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit1, &cpu.Regs.B) }},
	&inst{0xC9, "SET 1,C", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit1, &cpu.Regs.C) }},
	&inst{0xCA, "SET 1,D", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit1, &cpu.Regs.D) }},
	&inst{0xCB, "SET 1,E", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit1, &cpu.Regs.E) }},
	&inst{0xCC, "SET 1,H", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit1, &cpu.Regs.H) }},
	&inst{0xCD, "SET 1,L", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit1, &cpu.Regs.L) }},
	&inst{0xCE, "SET 1,(HL)", 0, 4, func(cpu *CPU, operands []byte) { cpu.set_b_hl(types.Bit1) }},
	&inst{0xCF, "SET 1,A", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit1, &cpu.Regs.A) }},
	&inst{0xD0, "SET 2,B", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit2, &cpu.Regs.B) }},
	&inst{0xD1, "SET 2,C", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit2, &cpu.Regs.C) }},
	&inst{0xD2, "SET 2,D", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit2, &cpu.Regs.D) }},
	&inst{0xD3, "SET 2,E", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit2, &cpu.Regs.E) }},
	&inst{0xD4, "SET 2,H", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit2, &cpu.Regs.H) }},
	&inst{0xD5, "SET 2,L", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit2, &cpu.Regs.L) }},
	&inst{0xD6, "SET 2,(HL)", 0, 4, func(cpu *CPU, operands []byte) { cpu.set_b_hl(types.Bit2) }},
	&inst{0xD7, "SET 2,A", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit2, &cpu.Regs.A) }},
	&inst{0xD8, "SET 3,B", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit3, &cpu.Regs.B) }},
	&inst{0xD9, "SET 3,C", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit3, &cpu.Regs.C) }},
	&inst{0xDA, "SET 3,D", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit3, &cpu.Regs.D) }},
	&inst{0xDB, "SET 3,E", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit3, &cpu.Regs.E) }},
	&inst{0xDC, "SET 3,H", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit3, &cpu.Regs.H) }},
	&inst{0xDD, "SET 3,L", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit3, &cpu.Regs.L) }},
	&inst{0xDE, "SET 3,(HL)", 0, 4, func(cpu *CPU, operands []byte) { cpu.set_b_hl(types.Bit3) }},
	&inst{0xDF, "SET 3,A", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit3, &cpu.Regs.A) }},
	&inst{0xE0, "SET 4,B", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit4, &cpu.Regs.B) }},
	&inst{0xE1, "SET 4,C", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit4, &cpu.Regs.C) }},
	&inst{0xE2, "SET 4,D", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit4, &cpu.Regs.D) }},
	&inst{0xE3, "SET 4,E", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit4, &cpu.Regs.E) }},
	&inst{0xE4, "SET 4,H", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit4, &cpu.Regs.H) }},
	&inst{0xE5, "SET 4,L", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit4, &cpu.Regs.L) }},
	&inst{0xE6, "SET 4,(HL)", 0, 4, func(cpu *CPU, operands []byte) { cpu.set_b_hl(types.Bit4) }},
	&inst{0xE7, "SET 4,A", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit4, &cpu.Regs.A) }},
	&inst{0xE8, "SET 5,B", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit5, &cpu.Regs.B) }},
	&inst{0xE9, "SET 5,C", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit5, &cpu.Regs.C) }},
	&inst{0xEA, "SET 5,D", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit5, &cpu.Regs.D) }},
	&inst{0xEB, "SET 5,E", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit5, &cpu.Regs.E) }},
	&inst{0xEC, "SET 5,H", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit5, &cpu.Regs.H) }},
	&inst{0xED, "SET 5,L", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit5, &cpu.Regs.L) }},
	&inst{0xEE, "SET 5,(HL)", 0, 4, func(cpu *CPU, operands []byte) { cpu.set_b_hl(types.Bit5) }},
	&inst{0xEF, "SET 5,A", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit5, &cpu.Regs.A) }},
	&inst{0xF0, "SET 6,B", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit6, &cpu.Regs.B) }},
	&inst{0xF1, "SET 6,C", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit6, &cpu.Regs.C) }},
	&inst{0xF2, "SET 6,D", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit6, &cpu.Regs.D) }},
	&inst{0xF3, "SET 6,E", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit6, &cpu.Regs.E) }},
	&inst{0xF4, "SET 6,H", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit6, &cpu.Regs.H) }},
	&inst{0xF5, "SET 6,L", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit6, &cpu.Regs.L) }},
	&inst{0xF6, "SET 6,(HL)", 0, 4, func(cpu *CPU, operands []byte) { cpu.set_b_hl(types.Bit6) }},
	&inst{0xF7, "SET 6,A", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit6, &cpu.Regs.A) }},
	&inst{0xF8, "SET 7,B", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit7, &cpu.Regs.B) }},
	&inst{0xF9, "SET 7,C", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit7, &cpu.Regs.C) }},
	&inst{0xFA, "SET 7,D", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit7, &cpu.Regs.D) }},
	&inst{0xFB, "SET 7,E", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit7, &cpu.Regs.E) }},
	&inst{0xFC, "SET 7,H", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit7, &cpu.Regs.H) }},
	&inst{0xFD, "SET 7,L", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit7, &cpu.Regs.L) }},
	&inst{0xFE, "SET 7,(HL)", 0, 4, func(cpu *CPU, operands []byte) { cpu.set_b_hl(types.Bit7) }},
	&inst{0xFF, "SET 7,A", 0, 2, func(cpu *CPU, operands []byte) { cpu.set_b_r(types.Bit7, &cpu.Regs.A) }},
}

var instructions = []*inst{
	&inst{0x0, "NOP", 0, 1, func(cpu *CPU, operands []byte) { cpu.nop() }},
	&inst{0x1, "LD BC,nn", 2, 3, func(cpu *CPU, operands []byte) { cpu.ldn_nn(&cpu.Regs.B, &cpu.Regs.C, operands) }},
	&inst{0x2, "LD (BC),A", 0, 2, func(cpu *CPU, operands []byte) { cpu.ldrr_r(cpu.Regs.B, cpu.Regs.C, cpu.Regs.A) }},
	&inst{0x3, "INC BC", 0, 2, func(cpu *CPU, operands []byte) { cpu.inc_nn(&cpu.Regs.B, &cpu.Regs.C) }},
	&inst{0x4, "INC B", 0, 1, func(cpu *CPU, operands []byte) { cpu.inc_n(&cpu.Regs.B) }},
	&inst{0x5, "DEC B", 0, 1, func(cpu *CPU, operands []byte) { cpu.dec_n(&cpu.Regs.B) }},
	&inst{0x6, "LD B,n", 1, 2, func(cpu *CPU, operands []byte) { cpu.ldnn_n(&cpu.Regs.B, operands) }},
	&inst{0x7, "RLCA", 0, 1, func(cpu *CPU, operands []byte) { cpu.rlca() }},
	&inst{0x8, "LD (nn),SP", 2, 5, func(cpu *CPU, operands []byte) { cpu.ldnn_sp(operands) }},
	&inst{0x9, "ADD HL,BC", 0, 2, func(cpu *CPU, operands []byte) { cpu.addhl_rr(&cpu.Regs.B, &cpu.Regs.C) }},
	&inst{0xA, "LD A,(BC)", 0, 2, func(cpu *CPU, operands []byte) { cpu.ldr_rr(cpu.Regs.B, cpu.Regs.C, &cpu.Regs.A) }},
	&inst{0xB, "DEC BC", 0, 2, func(cpu *CPU, operands []byte) { cpu.dec_nn(&cpu.Regs.B, &cpu.Regs.C) }},
	&inst{0xC, "INC C", 0, 1, func(cpu *CPU, operands []byte) { cpu.inc_n(&cpu.Regs.C) }},
	&inst{0xD, "DEC C", 0, 1, func(cpu *CPU, operands []byte) { cpu.dec_n(&cpu.Regs.C) }},
	&inst{0xE, "LD C,n", 1, 2, func(cpu *CPU, operands []byte) { cpu.ldnn_n(&cpu.Regs.C, operands) }},
	&inst{0xF, "RRCA", 0, 1, func(cpu *CPU, operands []byte) { cpu.rrca() }},
	&inst{0x10, "STOP", 1, 0, func(cpu *CPU, operands []byte) { cpu.stop() }},
	&inst{0x11, "LD DE,(nn)", 2, 3, func(cpu *CPU, operands []byte) { cpu.ldn_nn(&cpu.Regs.D, &cpu.Regs.E, operands) }},
	&inst{0x12, "LD (DE),A", 0, 2, func(cpu *CPU, operands []byte) { cpu.ldrr_r(cpu.Regs.D, cpu.Regs.E, cpu.Regs.A) }},
	&inst{0x13, "INC DE", 0, 2, func(cpu *CPU, operands []byte) { cpu.inc_nn(&cpu.Regs.D, &cpu.Regs.E) }},
	&inst{0x14, "INC D", 0, 1, func(cpu *CPU, operands []byte) { cpu.inc_n(&cpu.Regs.D) }},
	&inst{0x15, "DEC D", 0, 1, func(cpu *CPU, operands []byte) { cpu.dec_n(&cpu.Regs.D) }},
	&inst{0x16, "LD D,n", 1, 2, func(cpu *CPU, operands []byte) { cpu.ldnn_n(&cpu.Regs.D, operands) }},
	&inst{0x17, "RLA", 0, 1, func(cpu *CPU, operands []byte) { cpu.rla() }},
	&inst{0x18, "JR n", 1, 3, func(cpu *CPU, operands []byte) { cpu.jr_n(operands) }},
	&inst{0x19, "ADD HL,DE", 0, 2, func(cpu *CPU, operands []byte) { cpu.addhl_rr(&cpu.Regs.D, &cpu.Regs.E) }},
	&inst{0x1A, "LD A,(DE)", 0, 2, func(cpu *CPU, operands []byte) { cpu.ldr_rr(cpu.Regs.D, cpu.Regs.E, &cpu.Regs.A) }},
	&inst{0x1B, "DEC DE", 0, 2, func(cpu *CPU, operands []byte) { cpu.dec_nn(&cpu.Regs.D, &cpu.Regs.E) }},
	&inst{0x1C, "INC E", 0, 1, func(cpu *CPU, operands []byte) { cpu.inc_n(&cpu.Regs.E) }},
	&inst{0x1D, "DEC E", 0, 1, func(cpu *CPU, operands []byte) { cpu.dec_n(&cpu.Regs.E) }},
	&inst{0x1E, "LD E,n", 1, 2, func(cpu *CPU, operands []byte) { cpu.ldnn_n(&cpu.Regs.E, operands) }},
	&inst{0x1F, "RRA", 0, 1, func(cpu *CPU, operands []byte) { cpu.rra() }},
	&inst{0x20, "JR NZ,*", 1, 2, func(cpu *CPU, operands []byte) { cpu.jrcc_n(Z, false, operands) }},
	&inst{0x21, "LD HL,nn", 2, 3, func(cpu *CPU, operands []byte) { cpu.ldn_nn(&cpu.Regs.H, &cpu.Regs.L, operands) }},
	&inst{0x22, "LD (HL+),A", 0, 2, func(cpu *CPU, operands []byte) { cpu.ldihl_a() }},
	&inst{0x23, "INC HL", 0, 2, func(cpu *CPU, operands []byte) { cpu.inc_nn(&cpu.Regs.H, &cpu.Regs.L) }},
	&inst{0x24, "INC H", 0, 1, func(cpu *CPU, operands []byte) { cpu.inc_n(&cpu.Regs.H) }},
	&inst{0x25, "DEC H", 0, 1, func(cpu *CPU, operands []byte) { cpu.dec_n(&cpu.Regs.H) }},
	&inst{0x26, "LD H,n", 1, 2, func(cpu *CPU, operands []byte) { cpu.ldnn_n(&cpu.Regs.H, operands) }},
	&inst{0x27, "DAA", 0, 1, func(cpu *CPU, operands []byte) { cpu.daa() }},
	&inst{0x28, "JR Z,*", 1, 2, func(cpu *CPU, operands []byte) { cpu.jrcc_n(Z, true, operands) }},
	&inst{0x29, "ADD HL,HL", 0, 2, func(cpu *CPU, operands []byte) { cpu.addhl_rr(&cpu.Regs.H, &cpu.Regs.L) }},
	&inst{0x2A, "LDI A,(HL)", 0, 2, func(cpu *CPU, operands []byte) { cpu.ldia_hl() }},
	&inst{0x2B, "DEC HL", 0, 2, func(cpu *CPU, operands []byte) { cpu.dec_nn(&cpu.Regs.H, &cpu.Regs.L) }},
	&inst{0x2C, "INC L", 0, 1, func(cpu *CPU, operands []byte) { cpu.inc_n(&cpu.Regs.L) }},
	&inst{0x2D, "DEC L", 0, 1, func(cpu *CPU, operands []byte) { cpu.dec_n(&cpu.Regs.L) }},
	&inst{0x2E, "LD L,n", 1, 2, func(cpu *CPU, operands []byte) { cpu.ldnn_n(&cpu.Regs.L, operands) }},
	&inst{0x2F, "CPL", 0, 1, func(cpu *CPU, operands []byte) { cpu.cpl() }},
	&inst{0x30, "JR NC,*", 1, 2, func(cpu *CPU, operands []byte) { cpu.jrcc_n(C, false, operands) }},
	&inst{0x31, "LD SP,nn", 2, 3, func(cpu *CPU, operands []byte) { cpu.ldsp_nn(operands) }},
	&inst{0x32, "LD (HL-),A", 0, 2, func(cpu *CPU, operands []byte) { cpu.lddhl_a() }},
	&inst{0x33, "INC SP", 0, 2, func(cpu *CPU, operands []byte) { cpu.inc_sp() }},
	&inst{0x34, "INC (HL)", 0, 3, func(cpu *CPU, operands []byte) { cpu.inc_hl() }},
	&inst{0x35, "DEC (HL)", 0, 3, func(cpu *CPU, operands []byte) { cpu.dec_hl() }},
	&inst{0x36, "LD (HL),n", 1, 3, func(cpu *CPU, operands []byte) { cpu.ldhl_n(operands) }},
	&inst{0x37, "SCF", 0, 1, func(cpu *CPU, operands []byte) { cpu.scf() }},
	&inst{0x38, "JR C,*", 1, 2, func(cpu *CPU, operands []byte) { cpu.jrcc_n(C, true, operands) }},
	&inst{0x39, "ADD HL,SP", 0, 2, func(cpu *CPU, operands []byte) { cpu.addhl_sp() }},
	&inst{0x3A, "LD A,(HL-)", 0, 2, func(cpu *CPU, operands []byte) { cpu.ldda_hl() }},
	&inst{0x3B, "DEC SP", 0, 2, func(cpu *CPU, operands []byte) { cpu.dec_sp() }},
	&inst{0x3C, "INC A", 0, 1, func(cpu *CPU, operands []byte) { cpu.inc_n(&cpu.Regs.A) }},
	&inst{0x3D, "DEC A", 0, 1, func(cpu *CPU, operands []byte) { cpu.dec_r(&cpu.Regs.A) }},
	&inst{0x3E, "LD A,#", 1, 2, func(cpu *CPU, operands []byte) { cpu.lda_n(operands) }},
	&inst{0x3F, "CCF", 0, 1, func(cpu *CPU, operands []byte) { cpu.ccf() }},
	&inst{0x40, "LD B,B", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.B, &cpu.Regs.B) }},
	&inst{0x41, "LD B,C", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.B, &cpu.Regs.C) }},
	&inst{0x42, "LD B,D", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.B, &cpu.Regs.D) }},
	&inst{0x43, "LD B,E", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.B, &cpu.Regs.E) }},
	&inst{0x44, "LD B,H", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.B, &cpu.Regs.H) }},
	&inst{0x45, "LD B,L", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.B, &cpu.Regs.L) }},
	&inst{0x46, "LD B,(HL)", 0, 2, func(cpu *CPU, operands []byte) { cpu.ldr_rr(cpu.Regs.H, cpu.Regs.L, &cpu.Regs.B) }},
	&inst{0x47, "LD B,A", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.B, &cpu.Regs.A) }},
	&inst{0x48, "LD C,B", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.C, &cpu.Regs.B) }},
	&inst{0x49, "LD C,C", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.C, &cpu.Regs.C) }},
	&inst{0x4A, "LD C,D", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.C, &cpu.Regs.D) }},
	&inst{0x4B, "LD C,E", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.C, &cpu.Regs.E) }},
	&inst{0x4C, "LD C,H", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.C, &cpu.Regs.H) }},
	&inst{0x4D, "LD C,L", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.C, &cpu.Regs.L) }},
	&inst{0x4E, "LD C,(HL)", 0, 2, func(cpu *CPU, operands []byte) { cpu.ldr_rr(cpu.Regs.H, cpu.Regs.L, &cpu.Regs.C) }},
	&inst{0x4F, "LD C,A", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.C, &cpu.Regs.A) }},
	&inst{0x50, "LD D,B", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.D, &cpu.Regs.B) }},
	&inst{0x51, "LD D,C", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.D, &cpu.Regs.C) }},
	&inst{0x52, "LD D,D", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.D, &cpu.Regs.D) }},
	&inst{0x53, "LD D,E", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.D, &cpu.Regs.E) }},
	&inst{0x54, "LD D,H", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.D, &cpu.Regs.H) }},
	&inst{0x55, "LD D,L", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.D, &cpu.Regs.L) }},
	&inst{0x56, "LD D,(HL)", 0, 2, func(cpu *CPU, operands []byte) { cpu.ldr_rr(cpu.Regs.H, cpu.Regs.L, &cpu.Regs.D) }},
	&inst{0x57, "LD D,A", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.D, &cpu.Regs.A) }},
	&inst{0x58, "LD E,B", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.E, &cpu.Regs.B) }},
	&inst{0x59, "LD E,C", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.E, &cpu.Regs.C) }},
	&inst{0x5A, "LD E,D", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.E, &cpu.Regs.D) }},
	&inst{0x5B, "LD E,E", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.E, &cpu.Regs.E) }},
	&inst{0x5C, "LD E,H", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.E, &cpu.Regs.H) }},
	&inst{0x5D, "LD E,L", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.E, &cpu.Regs.L) }},
	&inst{0x5E, "LD E,(HL)", 0, 2, func(cpu *CPU, operands []byte) { cpu.ldr_rr(cpu.Regs.H, cpu.Regs.L, &cpu.Regs.E) }},
	&inst{0x5F, "LD E,A", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.E, &cpu.Regs.A) }},
	&inst{0x60, "LD H,B", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.H, &cpu.Regs.B) }},
	&inst{0x61, "LD H,C", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.H, &cpu.Regs.C) }},
	&inst{0x62, "LD H,D", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.H, &cpu.Regs.D) }},
	&inst{0x63, "LD H,E", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.H, &cpu.Regs.E) }},
	&inst{0x64, "LD H,H", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.H, &cpu.Regs.H) }},
	&inst{0x65, "LD H,L", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.H, &cpu.Regs.L) }},
	&inst{0x66, "LD H,(HL)", 0, 2, func(cpu *CPU, operands []byte) { cpu.ldr_rr(cpu.Regs.H, cpu.Regs.L, &cpu.Regs.H) }},
	&inst{0x67, "LD H,A", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.H, &cpu.Regs.A) }},
	&inst{0x68, "LD L,B", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.L, &cpu.Regs.B) }},
	&inst{0x69, "LD L,C", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.L, &cpu.Regs.C) }},
	&inst{0x6A, "LD L,D", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.L, &cpu.Regs.D) }},
	&inst{0x6B, "LD L,E", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.L, &cpu.Regs.E) }},
	&inst{0x6C, "LD L,H", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.L, &cpu.Regs.H) }},
	&inst{0x6D, "LD L,L", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.L, &cpu.Regs.L) }},
	&inst{0x6E, "LD L,(HL)", 0, 2, func(cpu *CPU, operands []byte) { cpu.ldr_rr(cpu.Regs.H, cpu.Regs.L, &cpu.Regs.L) }},
	&inst{0x6F, "LD L,A", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.L, &cpu.Regs.A) }},
	&inst{0x70, "LD (HL),B", 0, 2, func(cpu *CPU, operands []byte) { cpu.ldrr_r(cpu.Regs.H, cpu.Regs.L, cpu.Regs.B) }},
	&inst{0x71, "LD (HL),C", 0, 2, func(cpu *CPU, operands []byte) { cpu.ldrr_r(cpu.Regs.H, cpu.Regs.L, cpu.Regs.C) }},
	&inst{0x72, "LD (HL),D", 0, 2, func(cpu *CPU, operands []byte) { cpu.ldrr_r(cpu.Regs.H, cpu.Regs.L, cpu.Regs.D) }},
	&inst{0x73, "LD (HL),E", 0, 2, func(cpu *CPU, operands []byte) { cpu.ldrr_r(cpu.Regs.H, cpu.Regs.L, cpu.Regs.E) }},
	&inst{0x74, "LD (HL),H", 0, 2, func(cpu *CPU, operands []byte) { cpu.ldrr_r(cpu.Regs.H, cpu.Regs.L, cpu.Regs.H) }},
	&inst{0x75, "LD (HL),L", 0, 2, func(cpu *CPU, operands []byte) { cpu.ldrr_r(cpu.Regs.H, cpu.Regs.L, cpu.Regs.L) }},
	&inst{0x76, "HALT", 0, 1, func(cpu *CPU, operands []byte) { cpu.halt() }},
	&inst{0x77, "LD (HL),A", 0, 2, func(cpu *CPU, operands []byte) { cpu.ldrr_r(cpu.Regs.H, cpu.Regs.L, cpu.Regs.A) }},
	&inst{0x78, "LD A,B", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.A, &cpu.Regs.B) }},
	&inst{0x79, "LD A,C", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.A, &cpu.Regs.C) }},
	&inst{0x7A, "LD A,D", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.A, &cpu.Regs.D) }},
	&inst{0x7B, "LD A,E", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.A, &cpu.Regs.E) }},
	&inst{0x7C, "LD A,H", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.A, &cpu.Regs.H) }},
	&inst{0x7D, "LD A,L", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.A, &cpu.Regs.L) }},
	&inst{0x7E, "LD A,(HL)", 0, 2, func(cpu *CPU, operands []byte) { cpu.ldr_rr(cpu.Regs.H, cpu.Regs.L, &cpu.Regs.A) }},
	&inst{0x7F, "LD A,A", 0, 1, func(cpu *CPU, operands []byte) { cpu.ldrr(&cpu.Regs.A, &cpu.Regs.A) }},
	&inst{0x80, "ADD A,B", 0, 1, func(cpu *CPU, operands []byte) { cpu.adda_n(cpu.Regs.B) }},
	&inst{0x81, "ADD A,C", 0, 1, func(cpu *CPU, operands []byte) { cpu.adda_n(cpu.Regs.C) }},
	&inst{0x82, "ADD A,D", 0, 1, func(cpu *CPU, operands []byte) { cpu.adda_n(cpu.Regs.D) }},
	&inst{0x83, "ADD A,E", 0, 1, func(cpu *CPU, operands []byte) { cpu.adda_n(cpu.Regs.E) }},
	&inst{0x84, "ADD A,H", 0, 1, func(cpu *CPU, operands []byte) { cpu.adda_n(cpu.Regs.H) }},
	&inst{0x85, "ADD A,L", 0, 1, func(cpu *CPU, operands []byte) { cpu.adda_n(cpu.Regs.L) }},
	&inst{0x86, "ADD A,(HL)", 0, 2, func(cpu *CPU, operands []byte) { cpu.adda_n(cpu.bus.ReadByte(cpu.getHL())) }},
	&inst{0x87, "ADD A,A", 0, 1, func(cpu *CPU, operands []byte) { cpu.adda_n(cpu.Regs.A) }},
	&inst{0x88, "ADC A,B", 0, 1, func(cpu *CPU, operands []byte) { cpu.adca_n(cpu.Regs.B) }},
	&inst{0x89, "ADC A,C", 0, 1, func(cpu *CPU, operands []byte) { cpu.adca_n(cpu.Regs.C) }},
	&inst{0x8A, "ADC A,D", 0, 1, func(cpu *CPU, operands []byte) { cpu.adca_n(cpu.Regs.D) }},
	&inst{0x8B, "ADC A,E", 0, 1, func(cpu *CPU, operands []byte) { cpu.adca_n(cpu.Regs.E) }},
	&inst{0x8C, "ADC A,H", 0, 1, func(cpu *CPU, operands []byte) { cpu.adca_n(cpu.Regs.H) }},
	&inst{0x8D, "ADC A,L", 0, 1, func(cpu *CPU, operands []byte) { cpu.adca_n(cpu.Regs.L) }},
	&inst{0x8E, "ADC A,(HL)", 0, 2, func(cpu *CPU, operands []byte) { cpu.adca_n(cpu.bus.ReadByte(cpu.getHL())) }},
	&inst{0x8F, "ADC A,A", 0, 1, func(cpu *CPU, operands []byte) { cpu.adca_n(cpu.Regs.A) }},
	&inst{0x90, "SUB B", 0, 1, func(cpu *CPU, operands []byte) { cpu.sub_n(cpu.Regs.B) }},
	&inst{0x91, "SUB C", 0, 1, func(cpu *CPU, operands []byte) { cpu.sub_n(cpu.Regs.C) }},
	&inst{0x92, "SUB D", 0, 1, func(cpu *CPU, operands []byte) { cpu.sub_n(cpu.Regs.D) }},
	&inst{0x93, "SUB E", 0, 1, func(cpu *CPU, operands []byte) { cpu.sub_n(cpu.Regs.E) }},
	&inst{0x94, "SUB H", 0, 1, func(cpu *CPU, operands []byte) { cpu.sub_n(cpu.Regs.H) }},
	&inst{0x95, "SUB L", 0, 1, func(cpu *CPU, operands []byte) { cpu.sub_n(cpu.Regs.L) }},
	&inst{0x96, "SUB (HL)", 0, 2, func(cpu *CPU, operands []byte) { cpu.sub_n(cpu.bus.ReadByte(cpu.getHL())) }},
	&inst{0x97, "SUB A", 0, 1, func(cpu *CPU, operands []byte) { cpu.sub_n(cpu.Regs.A) }},
	&inst{0x98, "SBC A,B", 0, 1, func(cpu *CPU, operands []byte) { cpu.subca_n(cpu.Regs.B) }},
	&inst{0x99, "SBC A,C", 0, 1, func(cpu *CPU, operands []byte) { cpu.subca_n(cpu.Regs.C) }},
	&inst{0x9A, "SBC A,D", 0, 1, func(cpu *CPU, operands []byte) { cpu.subca_n(cpu.Regs.D) }},
	&inst{0x9B, "SBC A,E", 0, 1, func(cpu *CPU, operands []byte) { cpu.subca_n(cpu.Regs.E) }},
	&inst{0x9C, "SBC A,H", 0, 1, func(cpu *CPU, operands []byte) { cpu.subca_n(cpu.Regs.H) }},
	&inst{0x9D, "SBC A,L", 0, 1, func(cpu *CPU, operands []byte) { cpu.subca_n(cpu.Regs.L) }},
	&inst{0x9E, "SBC A,(HL)", 0, 2, func(cpu *CPU, operands []byte) { cpu.subca_n(cpu.bus.ReadByte(cpu.getHL())) }},
	&inst{0x9F, "SBC A,A", 0, 1, func(cpu *CPU, operands []byte) { cpu.subca_n(cpu.Regs.A) }},
	&inst{0xA0, "AND B", 0, 1, func(cpu *CPU, operands []byte) { cpu.and_n(cpu.Regs.B) }},
	&inst{0xA1, "AND C", 0, 1, func(cpu *CPU, operands []byte) { cpu.and_n(cpu.Regs.C) }},
	&inst{0xA2, "AND D", 0, 1, func(cpu *CPU, operands []byte) { cpu.and_n(cpu.Regs.D) }},
	&inst{0xA3, "AND E", 0, 1, func(cpu *CPU, operands []byte) { cpu.and_n(cpu.Regs.E) }},
	&inst{0xA4, "AND H", 0, 1, func(cpu *CPU, operands []byte) { cpu.and_n(cpu.Regs.H) }},
	&inst{0xA5, "AND L", 0, 1, func(cpu *CPU, operands []byte) { cpu.and_n(cpu.Regs.L) }},
	&inst{0xA6, "AND (HL)", 0, 2, func(cpu *CPU, operands []byte) { cpu.and_n(cpu.bus.ReadByte(cpu.getHL())) }},
	&inst{0xA7, "AND A", 0, 1, func(cpu *CPU, operands []byte) { cpu.and_n(cpu.Regs.A) }},
	&inst{0xA8, "XOR B", 0, 1, func(cpu *CPU, operands []byte) { cpu.xor_n(cpu.Regs.B) }},
	&inst{0xA9, "XOR C", 0, 1, func(cpu *CPU, operands []byte) { cpu.xor_n(cpu.Regs.C) }},
	&inst{0xAA, "XOR D", 0, 1, func(cpu *CPU, operands []byte) { cpu.xor_n(cpu.Regs.D) }},
	&inst{0xAB, "XOR E", 0, 1, func(cpu *CPU, operands []byte) { cpu.xor_n(cpu.Regs.E) }},
	&inst{0xAC, "XOR H", 0, 1, func(cpu *CPU, operands []byte) { cpu.xor_n(cpu.Regs.H) }},
	&inst{0xAD, "XOR L", 0, 1, func(cpu *CPU, operands []byte) { cpu.xor_n(cpu.Regs.L) }},
	&inst{0xAE, "XOR (HL)", 0, 2, func(cpu *CPU, operands []byte) { cpu.xor_n(cpu.bus.ReadByte(cpu.getHL())) }},
	&inst{0xAF, "XOR A", 0, 1, func(cpu *CPU, operands []byte) { cpu.xor_n(cpu.Regs.A) }},
	&inst{0xB0, "OR B", 0, 1, func(cpu *CPU, operands []byte) { cpu.or_n(cpu.Regs.B) }},
	&inst{0xB1, "OR C", 0, 1, func(cpu *CPU, operands []byte) { cpu.or_n(cpu.Regs.C) }},
	&inst{0xB2, "OR D", 0, 1, func(cpu *CPU, operands []byte) { cpu.or_n(cpu.Regs.D) }},
	&inst{0xB3, "OR E", 0, 1, func(cpu *CPU, operands []byte) { cpu.or_n(cpu.Regs.E) }},
	&inst{0xB4, "OR H", 0, 1, func(cpu *CPU, operands []byte) { cpu.or_n(cpu.Regs.H) }},
	&inst{0xB5, "OR L", 0, 1, func(cpu *CPU, operands []byte) { cpu.or_n(cpu.Regs.L) }},
	&inst{0xB6, "OR (HL)", 0, 2, func(cpu *CPU, operands []byte) { cpu.or_n(cpu.bus.ReadByte(cpu.getHL())) }},
	&inst{0xB7, "OR A", 0, 1, func(cpu *CPU, operands []byte) { cpu.or_n(cpu.Regs.A) }},
	&inst{0xB8, "CP B", 0, 1, func(cpu *CPU, operands []byte) { cpu.cp_n(cpu.Regs.B) }},
	&inst{0xB9, "CP C", 0, 1, func(cpu *CPU, operands []byte) { cpu.cp_n(cpu.Regs.C) }},
	&inst{0xBA, "CP D", 0, 1, func(cpu *CPU, operands []byte) { cpu.cp_n(cpu.Regs.D) }},
	&inst{0xBB, "CP E", 0, 1, func(cpu *CPU, operands []byte) { cpu.cp_n(cpu.Regs.E) }},
	&inst{0xBC, "CP H", 0, 1, func(cpu *CPU, operands []byte) { cpu.cp_n(cpu.Regs.H) }},
	&inst{0xBD, "CP L", 0, 1, func(cpu *CPU, operands []byte) { cpu.cp_n(cpu.Regs.L) }},
	&inst{0xBE, "CP (HL)", 0, 2, func(cpu *CPU, operands []byte) { cpu.cp_n(cpu.bus.ReadByte(cpu.getHL())) }},
	&inst{0xBF, "CP A", 0, 1, func(cpu *CPU, operands []byte) { cpu.cp_n(cpu.Regs.A) }},
	&inst{0xC0, "RET NZ", 0, 2, func(cpu *CPU, operands []byte) { cpu.retcc(Z, false) }},
	&inst{0xC1, "POP BC", 0, 3, func(cpu *CPU, operands []byte) { cpu.pop_nn(&cpu.Regs.B, &cpu.Regs.C) }},
	&inst{0xC2, "JP NZ,nn", 2, 3, func(cpu *CPU, operands []byte) { cpu.jpcc_nn(Z, false, operands) }},
	&inst{0xC3, "JP nn", 2, 4, func(cpu *CPU, operands []byte) { cpu.jp_nn(operands) }},
	&inst{0xC4, "CALL NZ,nn", 2, 3, func(cpu *CPU, operands []byte) { cpu.callcc_nn(Z, false, operands) }},
	&inst{0xC5, "PUSH BC", 0, 4, func(cpu *CPU, operands []byte) { cpu.push_nn(cpu.Regs.B, cpu.Regs.C) }},
	&inst{0xC6, "ADD A,#", 1, 2, func(cpu *CPU, operands []byte) { cpu.adda_n(operands[0]) }},
	&inst{0xC7, "RST n", 0, 4, func(cpu *CPU, operands []byte) { cpu.rst(0x00) }},
	&inst{0xC8, "RET Z", 0, 2, func(cpu *CPU, operands []byte) { cpu.retcc(Z, true) }},
	&inst{0xC9, "RET", 0, 4, func(cpu *CPU, operands []byte) { cpu.ret() }},
	&inst{0xCA, "JP Z,nn", 2, 3, func(cpu *CPU, operands []byte) { cpu.jpcc_nn(Z, true, operands) }},
	EMPTY,
	&inst{0xCC, "CALL Z,nn", 2, 3, func(cpu *CPU, operands []byte) { cpu.callcc_nn(Z, true, operands) }},
	&inst{0xCD, "CALL nn", 2, 3, func(cpu *CPU, operands []byte) { cpu.call_nn(operands) }},
	&inst{0xCE, "ADC A,#", 1, 2, func(cpu *CPU, operands []byte) { cpu.adca_n(operands[0]) }},
	&inst{0xCF, "RST n", 0, 4, func(cpu *CPU, operands []byte) { cpu.rst(0x08) }},
	&inst{0xD0, "RET NC", 0, 2, func(cpu *CPU, operands []byte) { cpu.retcc(C, false) }},
	&inst{0xD1, "POP DE", 0, 3, func(cpu *CPU, operands []byte) { cpu.pop_nn(&cpu.Regs.D, &cpu.Regs.E) }},
	&inst{0xD2, "JP NC,mm", 2, 3, func(cpu *CPU, operands []byte) { cpu.jpcc_nn(C, false, operands) }},
	EMPTY,
	&inst{0xD4, "CALL NC,nn", 2, 3, func(cpu *CPU, operands []byte) { cpu.callcc_nn(C, false, operands) }},
	&inst{0xD5, "PUSH DE", 0, 4, func(cpu *CPU, operands []byte) { cpu.push_nn(cpu.Regs.D, cpu.Regs.E) }},
	&inst{0xD6, "SUB n", 1, 2, func(cpu *CPU, operands []byte) { cpu.sub_n(operands[0]) }},
	&inst{0xD7, "RST n", 0, 4, func(cpu *CPU, operands []byte) { cpu.rst(0x10) }},
	&inst{0xD8, "RET C", 0, 2, func(cpu *CPU, operands []byte) { cpu.retcc(C, true) }},
	&inst{0xD9, "RETI", 0, 4, func(cpu *CPU, operands []byte) { cpu.ret_i() }},
	&inst{0xDA, "JP C,nn", 2, 3, func(cpu *CPU, operands []byte) { cpu.jpcc_nn(C, true, operands) }},
	EMPTY,
	&inst{0xDC, "CALL C,nn", 2, 3, func(cpu *CPU, operands []byte) { cpu.callcc_nn(C, true, operands) }},
	EMPTY,
	&inst{0xDE, "SBC A,#", 1, 2, func(cpu *CPU, operands []byte) { cpu.subca_n(operands[0]) }},
	&inst{0xDF, "RST n", 0, 4, func(cpu *CPU, operands []byte) { cpu.rst(0x18) }},
	&inst{0xE0, "LDH (n),A", 1, 3, func(cpu *CPU, operands []byte) { cpu.ldhn_a(operands) }},
	&inst{0xE1, "POP HL", 0, 3, func(cpu *CPU, operands []byte) { cpu.pop_nn(&cpu.Regs.H, &cpu.Regs.L) }},
	&inst{0xE2, "LD (C),A", 0, 2, func(cpu *CPU, operands []byte) { cpu.ldc_a() }},
	EMPTY,
	EMPTY,
	&inst{0xE5, "PUSH HL", 0, 4, func(cpu *CPU, operands []byte) { cpu.push_nn(cpu.Regs.H, cpu.Regs.L) }},
	&inst{0xE6, "AND n", 1, 2, func(cpu *CPU, operands []byte) { cpu.and_n(operands[0]) }},
	&inst{0xE7, "RST n", 0, 4, func(cpu *CPU, operands []byte) { cpu.rst(0x20) }},
	&inst{0xE8, "ADD SP,n", 1, 4, func(cpu *CPU, operands []byte) { cpu.addsp_n(operands) }},
	&inst{0xE9, "JP (HL)", 0, 1, func(cpu *CPU, operands []byte) { cpu.jp_hl() }},
	&inst{0xEA, "LD (nn),A", 2, 4, func(cpu *CPU, operands []byte) { cpu.ldnn_r(operands) }},
	EMPTY,
	EMPTY,
	EMPTY,
	&inst{0xEE, "XOR n", 1, 2, func(cpu *CPU, operands []byte) { cpu.xor_n(operands[0]) }},
	&inst{0xEF, "RST n", 0, 4, func(cpu *CPU, operands []byte) { cpu.rst(0x28) }},
	&inst{0xF0, "LDH A,(n)", 1, 3, func(cpu *CPU, operands []byte) { cpu.ldha_n(operands) }},
	&inst{0xF1, "POP AF", 0, 3, func(cpu *CPU, operands []byte) { cpu.pop_af() }},
	&inst{0xF2, "LD A,(C)", 0, 2, func(cpu *CPU, operands []byte) { cpu.lda_c() }},
	&inst{0xF3, "DI", 0, 1, func(cpu *CPU, operands []byte) { cpu.di() }},
	EMPTY,
	&inst{0xF5, "PUSH AF", 0, 4, func(cpu *CPU, operands []byte) { cpu.push_nn(cpu.Regs.A, cpu.Regs.F) }},
	&inst{0xF6, "OR #", 1, 2, func(cpu *CPU, operands []byte) { cpu.or_n(operands[0]) }},
	&inst{0xF7, "RST n", 0, 4, func(cpu *CPU, operands []byte) { cpu.rst(0x30) }},
	&inst{0xF8, "LD HL,SP+n", 1, 3, func(cpu *CPU, operands []byte) { cpu.ldhlsp_n(operands[0]) }},
	&inst{0xF9, "LD SP,HL", 0, 2, func(cpu *CPU, operands []byte) { cpu.ldsp_hl() }},
	&inst{0xFA, "LD A,(nn)", 2, 4, func(cpu *CPU, operands []byte) { cpu.lda_nn(operands) }},
	&inst{0xFB, "EI", 0, 1, func(cpu *CPU, operands []byte) { cpu.ei() }},
	EMPTY,
	EMPTY,
	&inst{0xFE, "CP n", 1, 2, func(cpu *CPU, operands []byte) { cpu.cp_n(operands[0]) }},
	&inst{0xFF, "RST n", 0, 4, func(cpu *CPU, operands []byte) { cpu.rst(0x38) }},
}

func (cpu *CPU) nop() {
	// NOP
}

// LD nn,n
// Description:
//  Put value nn into n.
// Use with:
//  nn = B,C,D,E,H,L,BC,DE,HL,SP
func (cpu *CPU) ldnn_n(reg *types.Register, operands []byte) {
	*reg = operands[0]
}

// LD n,nn
// Description:
//  Put value nn into n.
// Use with:
//  n = BC,DE,HL,SP
//  nn = 16 bit immediate value
func (cpu *CPU) ldn_nn(r1, r2 *types.Register, operands []byte) {
	*r1 = operands[1]
	*r2 = operands[0]
}

func (cpu *CPU) ldrr_r(upper, lower, r byte) {
	addr := utils.Bytes2Word(upper, lower)
	cpu.bus.WriteByte(addr, r)
}

// INC nn
// Description:
//  Increment register nn.
// Use with:
//  nn = BC,DE,HL,SP
// Flags affected:
//  None.
func (cpu *CPU) inc_nn(r1, r2 *types.Register) {
	data := types.Word(utils.Bytes2Word(*r1, *r2))
	data++
	*r1, *r2 = utils.Word2Bytes(data)
}

// INC n
// Description:
//  Increment register n.
// Use with:
//  n = A,B,C,D,E,H,L,(HL)
// Flags affected:
//  Z - Set if result is zero.
//  N - Reset.
//  H - Set if carry from bit 3.
//  C - Not affected
func (cpu *CPU) inc_n(r *types.Register) {
	*r = cpu.inc(*r)
}

func (cpu *CPU) dec_n(r *types.Register) {
	// cpu.logger.Info("dec_n", cpu.dec(*r))
	*r = cpu.dec(*r)
}

func (cpu *CPU) rlca() {
	computed := cpu.Regs.A << 1
	if cpu.Regs.A&0x80 == 0x80 {
		cpu.setFlag(C)
		computed ^= 0x01
	} else {
		cpu.clearFlag(C)
	}
	cpu.clearFlag(Z)
	cpu.clearFlag(N)
	cpu.clearFlag(H)
	cpu.Regs.A = computed
}

func (cpu *CPU) ldnn_sp(operands []byte) {
	addr := utils.Bytes2Word(operands[1], operands[0])
	cpu.bus.WriteWord(addr, cpu.SP)
}

func (cpu *CPU) addhl_rr(r1, r2 *byte) {
	hl := cpu.getHL()
	rr := utils.Bytes2Word(*r1, *r2)
	result := cpu.addWords(hl, rr)
	cpu.Regs.H, cpu.Regs.L = utils.Word2Bytes(result)
}

// LD dest,(r1, r2)
// Description:
//  Put value of (word(r1, r2)) to dest
// Use with:
//  r1,r2 = A,B,C,D,E,H,L
//  dest = A,B,C,D,E,H,L
func (cpu *CPU) ldr_rr(r1 byte, r2 byte, dest *byte) {
	addr := utils.Bytes2Word(r1, r2)
	*dest = cpu.bus.ReadByte(addr)
}

// LD A,n
// Description:
//  Put value n into A.
// Use with:
//  n = A
//  nn = two byte immediate value. (LS byte first.)
func (cpu *CPU) lda_nn(operands []byte) {
	addr := utils.Bytes2Word(operands[1], operands[0])
	cpu.Regs.A = cpu.bus.ReadByte(addr)
}

// DEC nn
// Description:
//  Decrement register nn.
//  Use with:
//   nn = BC,DE,HL,SP
//  Flags affected:
//   None.
func (cpu *CPU) dec_nn(r1, r2 *types.Register) {
	*r1, *r2 = utils.Word2Bytes(utils.Bytes2Word(*r1, *r2) - 1)
}

// RRCA
// Description:
//  Rotate A right. Old bit 0 to Carry flag.
// Flags affected:
//  Z - Set if result is zero.
//  N - Reset.
//  H - Reset.
//  C - Contains old bit 0 data.
func (cpu *CPU) rrca() {
	computed := cpu.Regs.A >> 1
	if cpu.Regs.A&0x01 == 0x01 {
		cpu.setFlag(C)
		computed ^= 0x80
	} else {
		cpu.clearFlag(C)
	}
	cpu.clearFlag(Z)
	cpu.clearFlag(N)
	cpu.clearFlag(H)
	cpu.Regs.A = computed
}

// Stop
// The STOP command halts the GameBoy processor
// and screen until any button is pressed. The GB
// and GBP screen goes white with a single dark
// horizontal line. The GBC screen goes black.
func (cpu *CPU) stop() {
	cpu.stopped = true
}

// RLA
// Description:
//  Rotate A left through Carry flag.
// Flags affected:
//  Z - Set if result is zero.
//  N - Reset.
//  H - Reset.
//  C - Contains old bit 7 data.
func (cpu *CPU) rla() {
	bit7 := false
	computed := cpu.Regs.A
	if computed&0x80 == 0x80 {
		bit7 = true
	}
	computed = computed << 1
	if cpu.isSet(C) {
		computed ^= 0x01
	}
	if bit7 {
		cpu.setFlag(C)
	} else {
		cpu.clearFlag(C)
	}
	cpu.clearFlag(Z)
	cpu.clearFlag(N)
	cpu.clearFlag(H)
	cpu.Regs.A = computed
}

//JR n
// Description:
//  Add n to current address and jump to it.
// Use with:
//  n = one byte signed immediate value
func (cpu *CPU) jr_n(operands []byte) {
	v := int8(operands[0])
	if v == 0x00 {
		return
	}
	if v < 0 {
		cpu.PC -= types.Word(-v)
	} else {
		cpu.PC += types.Word(v)
	}

}

//RRA
// Description:
//  Rotate A right through Carry flag.
// Flags affected:
//  Z - Set if result is zero.
//  N - Reset.
//  H - Reset.
//  C - Contains old bit 0 data.
func (cpu *CPU) rra() {
	bit0 := false
	computed := cpu.Regs.A

	if computed&0x01 == 0x01 {
		bit0 = true
	}
	computed = computed >> 1

	if cpu.isSet(C) {
		computed ^= 0x80
	}
	if bit0 {
		cpu.setFlag(C)
	} else {
		cpu.clearFlag(C)
	}
	cpu.clearFlag(Z)
	cpu.clearFlag(N)
	cpu.clearFlag(H)
	cpu.Regs.A = computed
}

// JR cc, n
// If following condition is true then add n to current
// address and jump to it:
// n = one byte signed immediate value
// cc = NZ, Jump if Z flag is reset.
// cc = Z, Jump if Z flag is set.
// cc = NC, Jump if C flag is reset.
// cc = C, Jump if C flag is set.
func (cpu *CPU) jrcc_n(flag flags, isSet bool, operands []byte) {
	n := int8(operands[0])
	if cpu.isSet(flag) == isSet {
		if n != 0x00 {
			if n < 0 {
				cpu.PC -= types.Word(-n)
			} else {
				cpu.PC += types.Word(n)
			}
		}
	}
}

// LD (HL+),A
// Description:
//  Put A into memory address HL. Increment HL.
//  Same as: LD (HL),A - INC HL
func (cpu *CPU) ldihl_a() {
	hl := types.Word(utils.Bytes2Word(cpu.Regs.H, cpu.Regs.L))
	cpu.bus.WriteByte(hl, cpu.Regs.A)
	hl++
	cpu.toHLRegs(hl)
}

// DAA
// Description:
//  Decimal adjust register A.
//  This instruction adjusts register A so that the
//  correct representation of Binary Coded Decimal (BCD)
//  is obtained.
// Flags affected:
//  Z - Set if register A is zero.
//  N - Not affected.
//  H - Reset.
//  C - Set or reset according to operation.
func (cpu *CPU) daa() {
	a := types.Word(cpu.Regs.A)
	if cpu.isSet(N) == false {
		if cpu.isSet(H) || a&0x0F > 9 {
			a += 0x06
		}
		if cpu.isSet(C) || a > 0x9F {
			a += 0x60
		}
	} else {
		if cpu.isSet(H) {
			a = (a - 0x06) & 0xFF
		}
		if cpu.isSet(C) {
			a -= 0x60
		}
	}
	cpu.clearFlag(H)
	if a&0x100 == 0x100 {
		cpu.setFlag(C)
	}
	a &= 0xFF
	cpu.applyZeroBy(byte(a))
	cpu.Regs.A = byte(a)
}

// LDI A,(HL)
// Description:
// Put value at address HL into A. Increment HL.
//  Same as: LD A,(HL) - INC HL
func (cpu *CPU) ldia_hl() {
	hl := cpu.getHL()
	cpu.Regs.A = cpu.bus.ReadByte(hl)
	hl++
	cpu.toHLRegs(hl)
}

//CPL
// Description:
//  Complement A register. (Flip all bits.)
// Flags affected:
//  Z - Not affected.
//  N - Set.
//  H - Set.
//  C - Not affected.
func (cpu *CPU) cpl() {
	cpu.Regs.A = ^cpu.Regs.A
	cpu.setFlag(N)
	cpu.setFlag(H)
}

// LD SP,nn
// Description:
//  Put value nn into n.
// Use with:
//  nn = 16 bit immediate value
func (cpu *CPU) ldsp_nn(operands []byte) {
	cpu.SP = utils.Bytes2Word(operands[1], operands[0])
}

// LDD (HL),A
// Description:
//  Put A into memory address HL. Decrement HL.
func (cpu *CPU) lddhl_a() {
	hl := cpu.getHL()
	cpu.bus.WriteByte(hl, cpu.Regs.A)
	hl--
	cpu.toHLRegs(hl)
}

// Description:
//  Increment register SP.
// Flags affected:
//  None.
func (cpu *CPU) inc_sp() {
	cpu.SP = (cpu.SP + 1) & 0xFFFF
}

//  INC (HL)
// Description:
//  Increment value pointed HL.
// Flags affected:
//  Z - Set if result is zero.
//  N - Reset.
//  H - Set if carry from bit 3.
//  C - Not affected.
func (cpu *CPU) inc_hl() {
	hl := cpu.getHL()
	v := cpu.bus.ReadByte(hl)
	result := cpu.inc(v)
	cpu.bus.WriteByte(hl, result)
}

//DEC (HL)
// Description:
//  Decrement value pointed HL.
// Flags affected:
//  Z - Set if result is zero.
//  N - Reset.
//  H - Set if carry from bit 3.
//  C - Not affected.
func (cpu *CPU) dec_hl() {
	hl := cpu.getHL()
	v := cpu.bus.ReadByte(hl)
	result := cpu.dec(v)
	cpu.bus.WriteByte(hl, result)
}

//LD (HL),n
// Description:
// Put value operands[0] into (HL)
func (cpu *CPU) ldhl_n(operands []byte) {
	hl := cpu.getHL()
	cpu.bus.WriteByte(hl, operands[0])
}

//SCF
// Description:
//  Set Carry flag.
// Flags affected:
//  Z - Not affected.
//  N - Reset.
//  H - Reset.
//  C - Set.
func (cpu *CPU) scf() {
	cpu.setFlag(C)
	cpu.clearFlag(N)
	cpu.clearFlag(H)
}

//ADD HL,SP
// Description:
//  Add SP to HL.
// Use with:
// Flags affected:
//  Z - Not affected.
//  N - Reset.
//  H - Set if carry from bit 11.
//  C - Set if carry from bit 15.
func (cpu *CPU) addhl_sp() {
	hl := cpu.getHL()
	result := cpu.addWords(hl, cpu.SP)
	cpu.toHLRegs(result)
}

// LDD A,(HL)
// Description:
//  Put value at address HL into A. Decrement HL.
//  Same as: LD A,(HL) - DEC HL
func (cpu *CPU) ldda_hl() {
	hl := cpu.getHL()
	cpu.Regs.A = cpu.bus.ReadByte(hl)
	hl--
	cpu.toHLRegs(hl)
}

// Description:
//  Decrement register SP.
// Flags affected:
//  None.
func (cpu *CPU) dec_sp() {
	cpu.SP = (cpu.SP - 1) & 0xFFFF
}

// INC n
// Description:
//  Increment register n.
// Use with:
//  n = A,B,C,D,E,H,L,(HL)
// Flags affected:
//  Z - Set if result is zero.
//  N - Reset.
//  H - Set if carry from bit 3.
//  C - Not affected.
func (cpu *CPU) inc_r(r *types.Register) {
	*r = cpu.inc(*r)
}

// DEC n
// Description:
//  Decrement register n.
// Use with:
//  n = A,B,C,D,E,H,L,(HL)
// Flags affected:
//  Z - Set if reselt is zero.
//  N - Set.
//  H - Set if no borrow from bit 4.
//  C - Not affected.
func (cpu *CPU) dec_r(r *types.Register) {
	*r = cpu.dec(*r)
}

// LD A,n
// Description:
//  Put value n into A.
// Use with:
//  n = A,B,C,D,E,H,L,(BC),(DE),(HL),(nn),#
//  nn = two byte immediate value. (LS byte first.)
func (cpu *CPU) lda_n(operands []byte) {
	cpu.Regs.A = operands[0]
}

// CCF
// Description:
//  Complement carry flag.
//  If C flag is set, then reset it.
//  If C flag is reset, then set it.
// Flags affected:
//  Z - Not affected.
//  N - Reset.
//  H - Reset.
//  C - Complemented.
func (cpu *CPU) ccf() {
	if cpu.isSet(C) {
		cpu.clearFlag(C)
	} else {
		cpu.setFlag(C)
	}
	cpu.clearFlag(N)
	cpu.clearFlag(H)
}

// LD r1,r2
// Description:
//  Put value r2 into r1.
// Use with:
//  r1,r2 = A,B,C,D,E,H,L
func (cpu *CPU) ldrr(r1 *types.Register, r2 *types.Register) {
	*r1 = *r2
}

// HALT
// Description:
//  Power down CPU until an interrupt occurs. Use this
//  when ever possible to reduce energy consumption.
func (cpu *CPU) halt() {
	cpu.halted = true
}

// ADD A,n
// Description:
//  Add n to A.
// Use with:
//  n = A,B,C,D,E,H,L,(HL),#
// Flags affected:
//  Z - Set if result is zero.
//  N - Reset.
//  H - Set if carry from bit 3.
//  C - Set if carry from bit 7.
func (cpu *CPU) adda_n(r byte) {
	cpu.Regs.A = cpu.addBytes(cpu.Regs.A, r)
}

// ADC A,n
// Description:
//  Add n + Carry flag to A.
// Use with:
//  n = A,B,C,D,E,H,L,(HL),#
// Flags affected:
//  Z - Set if result is zero.
//  N - Reset.
//  H - Set if carry from bit 3.
//  C - Set if carry from bit 7.
func (cpu *CPU) adca_n(n byte) {
	carry := uint(0)
	if cpu.isSet(C) {
		carry = 1
	}
	if ((cpu.Regs.A & 0xF) + (n & 0xF) + byte(carry)) > 0xF {
		cpu.setFlag(H)
	} else {
		cpu.clearFlag(H)
	}
	if ((uint(cpu.Regs.A & 0xFF)) + (uint(n) & 0xFF) + carry) > 0xFF {
		cpu.setFlag(C)
	} else {
		cpu.clearFlag(C)
	}
	cpu.clearFlag(N)
	cpu.Regs.A += n + byte(carry)
	cpu.applyZeroBy(cpu.Regs.A)
}

// SUB n
// Description:
//  Subtract n from A.
// Use with:
//  n = A,B,C,D,E,H,L,(HL),#
// Flags affected:
//  Z - Set if result is zero.
//  N - Set.
//  H - Set if no borrow from bit 4.
//  C - Set if no borrow.
func (cpu *CPU) sub_n(v byte) {
	cpu.Regs.A = cpu.subBytes(cpu.Regs.A, v)
}

// SBC A,n
// Description:
//  Subtract n + Carry flag from A.
// Use with:
//  n = A,B,C,D,E,H,L,(HL),#
// Flags affected:
//  Z - Set if result is zero.
//  N - Set.
//  H - Set if no borrow from bit 4.
//  C - Set if no borrow.
func (cpu *CPU) subca_n(v byte) {
	a := cpu.Regs.A
	computed := int(cpu.Regs.A)
	computed -= int(v)
	if cpu.isSet(C) {
		computed--
	}
	if computed < 0 {
		// cpu.setFlag(C)
		cpu.Regs.F = 0x50
	} else {
		cpu.Regs.F = 0x40
	}
	// cpu.setFlag(N)
	cpu.applyZeroBy(uint8(computed))

	if ((byte(computed) ^ v ^ a) & 0x10) == 0x10 {
		cpu.setFlag(H)
	} else {
		cpu.clearFlag(H)
	}

	cpu.Regs.A = byte(computed)
}

// AND n
// Description:
//  Logically AND n with A, result in A.
// Use with:
//  n = A,B,C,D,E,H,L,(HL),#
// Flags affected:
//  Z - Set if result is zero.`m`
//  N - Reset.
//  H - Set.
//  C - Reset.
func (cpu *CPU) and_n(v byte) {
	cpu.Regs.A = cpu.and(cpu.Regs.A, v)
}

// XOR n
// Description:
//  Logical exclusive OR n with register A, result in A.
// Use with:
//  n = A,B,C,D,E,H,L,(HL),#
// Flags affected:
//  Z - Set if result is zero.
//  N - Reset.
//  H - Reset.
//  C - Reset.
func (cpu *CPU) xor_n(v byte) {
	cpu.Regs.A = cpu.xor(cpu.Regs.A, v)
}

// OR n
// Description:
//  Logical OR n with register A, result in A.
// Use with:
//  n = A,B,C,D,E,H,L,(HL),#
// Flags affected:
//  Z - Set if result is zero.
//  N - Reset.
//  H - Reset.
//  C - Reset.
func (cpu *CPU) or_n(v byte) {
	cpu.Regs.A = cpu.or(cpu.Regs.A, v)
}

// CP n
// Description:
//  Compare A with n. This is basically an A - n
//  subtraction instruction but the results are thrown
//  away.
// Use with:
//  n = A,B,C,D,E,H,L,(HL),#
// Flags affected:
//  Z - Set if result is zero. (Set if A = n.)
//  N - Set.
//  H - Set if no borrow from bit 4.
//  C - Set for no borrow. (Set if A < n.)
func (cpu *CPU) cp_n(v byte) {
	cpu.setFlag(N)
	if cpu.Regs.A&0xF < v&0xF {
		cpu.setFlag(H)
	} else {
		cpu.clearFlag(H)
	}
	if cpu.Regs.A < v {
		cpu.setFlag(C)
	} else {
		cpu.clearFlag(C)
	}
	cpu.applyZeroBy(cpu.Regs.A - v)
}

// RET cc
// Description:
//  Return if following condition is true:
// Use with:
//  cc = NZ, Return if Z flag is reset.
//  cc = Z, Return if Z flag is set.
//  cc = NC, Return if C flag is reset.
//  cc = C, Return if C flag is set.
func (cpu *CPU) retcc(flag flags, isSet bool) {
	if cpu.isSet(flag) == isSet {
		cpu.pop2PC()
	}
}

// POP nn
// Description:
//  Pop two bytes off stack into register pair nn.
//  Increment Stack Pointer (SP) twice.
// Use with:
//  nn = AF,BC,DE,HL
func (cpu *CPU) pop_nn(r1, r2 *types.Register) {
	*r2 = cpu.pop()
	*r1 = cpu.pop()
}

func (cpu *CPU) pop_af() {
	cpu.Regs.F = cpu.pop() & 0xF0
	cpu.Regs.A = cpu.pop()
}

// JP cc,nn
// Description:
//  Jump to address n if following condition is true:
//  cc = NZ, Jump if Z flag is reset.
//  cc = Z, Jump if Z flag is set.
//  cc = NC, Jump if C flag is reset.
//  cc = C, Jump if C flag is set.
// Use with:
//  nn = two byte immediate value. (LS byte first.)
func (cpu *CPU) jpcc_nn(flag flags, isSet bool, operands []byte) {
	if cpu.isSet(flag) == isSet {
		cpu.PC = utils.Bytes2Word(operands[1], operands[0])
	}
}

// JP nn
// Description:
//  Jump to address nn.
// Use with:
//  nn = two byte immediate value. (LS byte first.)
func (cpu *CPU) jp_nn(operands []byte) {
	cpu.PC = utils.Bytes2Word(operands[1], operands[0])
}

// CALL cc,nn
// Description:
//  Call address n if following condition is true:
//  cc = NZ, Call if Z flag is reset.
//  cc = Z, Call if Z flag is set.
//  cc = NC, Call if C flag is reset.
//  cc = C, Call if C flag is set.
// Use with:
//  nn = two byte immediate value. (LS byte first.)
func (cpu *CPU) callcc_nn(flag flags, isSet bool, operands []byte) {
	if cpu.isSet(flag) == isSet {
		cpu.push(byte(cpu.PC >> 8))
		cpu.push(byte(cpu.PC & 0xFF))
		cpu.PC = utils.Bytes2Word(operands[1], operands[0])
	}
}

// PUSH nn
// Description:
//  Push register pair nn onto stack.
//  Decrement Stack Pointer (SP) twice.
// Use with:
//  nn = AF,BC,DE,HL
func (cpu *CPU) push_nn(h, l types.Register) {
	cpu.push(h)
	cpu.push(l)
}

// RST n
// Description:
//  Push present address onto stack.
//  Jump to address $0000 + n.
// Use with:
//  n = $00,$08,$10,$18,$20,$28,$30,$38
func (cpu *CPU) rst(n byte) {
	cpu.push(byte(cpu.PC >> 8))
	cpu.push(byte(cpu.PC & 0xFF))
	cpu.PC = types.Word(n)
}

// RET
// Description:
//  Pop two bytes from stack & jump to that address.
func (cpu *CPU) ret() {
	l := cpu.pop()
	h := cpu.pop()
	cpu.PC = utils.Bytes2Word(h, l)
}

// CALL nn
// Description:
//  Push address of next instruction onto stack and then
//  jump to address nn.
// Use with:
//  nn = two byte immediate value. (LS byte first.)
func (cpu *CPU) call_nn(operands []byte) {
	cpu.push(byte(cpu.PC >> 8))
	cpu.push(byte(cpu.PC & 0xFF))
	cpu.PC = utils.Bytes2Word(operands[1], operands[0])
}

// RETI
// Description:
//  Pop two bytes from stack & jump to that address then
//  enable interrupts.
func (cpu *CPU) ret_i() {
	l := cpu.pop()
	h := cpu.pop()
	cpu.PC = utils.Bytes2Word(h, l)
	cpu.irq.Enable()
}

// LDH (n),A
// Description:
//  Put A into memory address $FF00+n.
// Use with:
//  n = one byte immediate value.
func (cpu *CPU) ldhn_a(operands []byte) {
	cpu.bus.WriteByte(0xFF00+types.Word(operands[0]), cpu.Regs.A)
}

// LD (C),A
// Description:
//  Put A into address $FF00 + register C.
func (cpu *CPU) ldc_a() {
	addr := 0xFF00 + types.Word(cpu.Regs.C)
	cpu.bus.WriteByte(addr, cpu.Regs.A)
}

// ADD SP,n
// Description:
//  Add n to Stack Pointer (SP).
// Use with:
//  n = one byte signed immediate value (#).
// Flags affected:
//  Z - Reset.
//  N - Reset.
//  H - Set or reset according to operation.
//  C - Set or reset according to operation.
func (cpu *CPU) addsp_n(operands []byte) {
	n := operands[0]
	var computed types.Word
	if n > 127 {
		computed = cpu.SP - types.Word(-n)
	} else {
		computed = cpu.SP + types.Word(n)
	}
	c := types.Word(cpu.SP ^ types.Word(n) ^ ((cpu.SP + types.Word(n)) & 0xffff))
	cpu.SP = computed
	if (c & 0x100) == 0x100 {
		cpu.setFlag(C)
	} else {
		cpu.clearFlag(C)
	}

	if (c & 0x10) == 0x10 {
		cpu.setFlag(H)
	} else {
		cpu.clearFlag(H)
	}
	cpu.clearFlag(Z)
	cpu.clearFlag(N)
}

// JP (HL)
// Description:
//  Jump to address contained in HL.
func (cpu *CPU) jp_hl() {
	cpu.PC = utils.Bytes2Word(cpu.Regs.H, cpu.Regs.L)
}

// LD n,A
// Description:
//  Put value A into n.
// Use with:
//  n = A,B,C,D,E,H,L,(BC),(DE),(HL),(nn)
//  nn = two byte immediate value. (LS byte first.)
func (cpu *CPU) ldnn_r(operands []byte) {
	addr := utils.Bytes2Word(operands[1], operands[0])
	cpu.bus.WriteByte(addr, cpu.Regs.A)
}

// LDH A,(n)
// Description:
//  Put memory address $FF00+n into A.
// Use with:
//  n = one byte immediate value.
func (cpu *CPU) ldha_n(operands []byte) {
	cpu.Regs.A = cpu.bus.ReadByte(types.Word(0xFF00) + types.Word(operands[0]))
}

// LD A,(C)
// Description:
//  Put value at address $FF00 + register C into A.
//  Same as: LD A,($FF00+C)
func (cpu *CPU) lda_c() {
	cpu.Regs.A = cpu.bus.ReadByte(types.Word(0xFF00) + types.Word(cpu.Regs.C))
}

// DI
// Description:
//  This instruction disables interrupts but not
//  immediately. Interrupts are disabled after
//  instruction after DI is executed.
// Flags affected:
//  None.
func (cpu *CPU) di() {
	cpu.irq.Disable()
}

// EI
// Description:
//  Enable interrupts. This intruction enables interrupts
//  but not immediately. Interrupts are enabled after
//  instruction after EI is executed.
// Flags affected:
//  None.
func (cpu *CPU) ei() {
	cpu.irq.Enable()
}

// LD HL,SP+n / LDHL SP,n
// Description:
//  Put SP + n effective address into HL.
// Use with:
//  n = one byte signed immediate value.
// Flags affected:
//  Z - Reset.
//  N - Reset.
//  H - Set or reset according to operation.
//  C - Set or reset according to operation.
func (cpu *CPU) ldhlsp_n(n byte) {
	var hl types.Word
	if n > 127 {
		hl = cpu.SP - types.Word(-n)
	} else {
		hl = cpu.SP + types.Word(n)
	}
	c := types.Word(cpu.SP ^ types.Word(n) ^ ((cpu.SP + types.Word(n)) & 0xffff))
	if (c & 0x100) == 0x100 {
		cpu.setFlag(C)
	} else {
		cpu.clearFlag(C)
	}
	if (c & 0x10) == 0x10 {
		cpu.setFlag(H)
	} else {
		cpu.clearFlag(H)
	}
	cpu.clearFlag(Z)
	cpu.clearFlag(N)
	cpu.Regs.H, cpu.Regs.L = utils.Word2Bytes(hl)
}

// LD SP,HL
// Description:
//  Put HL into Stack Pointer (SP).
func (cpu *CPU) ldsp_hl() {
	cpu.SP = utils.Bytes2Word(cpu.Regs.H, cpu.Regs.L)
}

// RLC n
// Description:
//  Rotate n left. Old bit 7 to Carry flag.
// Use with:
//  n = A,B,C,D,E,H,L,(HL)
// Flags affected:
// Z - Set if result is zero.
//  N - Reset.
//  H - Reset.
//  C - Contains old bit 7 data.
func (cpu *CPU) rlc(v byte) byte {
	rotated := v << 1
	if v&0x80 == 0x80 {
		cpu.setFlag(C)
		rotated ^= 0x01
	} else {
		cpu.clearFlag(C)
	}
	cpu.clearFlag(N)
	cpu.clearFlag(H)
	cpu.applyZeroBy(rotated)
	return rotated
}

func (cpu *CPU) rlc_n(r *types.Register) {
	*r = cpu.rlc(*r)
}

func (cpu *CPU) rlc_hl() {
	addr := cpu.getHL()
	v := cpu.bus.ReadByte(addr)
	cpu.bus.WriteByte(addr, cpu.rlc(v))
}

//  RRC n
//  Description:
//   Rotate n right. Old bit 0 to Carry flag.
//  Use with:
//  n = A,B,C,D,E,H,L,(HL)
//  Flags affected:
//   Z - Set if result is zero.
//   N - Reset.
//   H - Reset.
//   C - Contains old bit 0 data.
func (cpu *CPU) rrc(v byte) byte {
	rotated := v >> 1
	if v&0x01 == 0x01 {
		cpu.setFlag(C)
		rotated ^= 0x80
	} else {
		cpu.clearFlag(C)
	}
	cpu.applyZeroBy(rotated)
	cpu.clearFlag(N)
	cpu.clearFlag(H)
	return rotated
}

func (cpu *CPU) rrc_n(r *types.Register) {
	*r = cpu.rrc(*r)
}

func (cpu *CPU) rrc_hl() {
	addr := cpu.getHL()
	v := cpu.bus.ReadByte(addr)
	cpu.bus.WriteByte(addr, cpu.rrc(v))
}

//  RL n
// Description:
//  Rotate n left through Carry flag.
// Use with:
//  n = A,B,C,D,E,H,L,(HL)
// Flags affected:
//  Z - Set if result is zero.
//  N - Reset.
//  H - Reset.
//  C - Contains old bit 7 data.

func (cpu *CPU) rl(v byte) byte {
	rotated := v << 1
	if cpu.isSet(C) {
		rotated ^= 0x01
	}
	if v&0x80 == 0x80 {
		cpu.setFlag(C)
	} else {
		cpu.clearFlag(C)
	}
	cpu.applyZeroBy(rotated)
	cpu.clearFlag(N)
	cpu.clearFlag(H)
	return rotated
}

func (cpu *CPU) rl_n(r *types.Register) {
	*r = cpu.rl(*r)
}

func (cpu *CPU) rl_hl() {
	addr := cpu.getHL()
	v := cpu.bus.ReadByte(addr)
	cpu.bus.WriteByte(addr, cpu.rl(v))
}

// RR n
// Description:
//  Rotate n right through Carry flag.
// Use with:
//  n = A,B,C,D,E,H,L,(HL)
// Flags affected:
//  Z - Set if result is zero.
//  N - Reset.
//  H - Reset.
//  C - Contains old bit 0 data.
func (cpu *CPU) rr(v byte) byte {
	rotated := v >> 1
	if cpu.isSet(C) {
		rotated ^= 0x80
	}
	if v&0x01 == 0x01 {
		cpu.setFlag(C)
	} else {
		cpu.clearFlag(C)
	}
	cpu.applyZeroBy(rotated)
	cpu.clearFlag(N)
	cpu.clearFlag(H)
	return rotated
}

func (cpu *CPU) rr_n(r *types.Register) {
	*r = cpu.rr(*r)
}

func (cpu *CPU) rr_hl() {
	addr := cpu.getHL()
	v := cpu.bus.ReadByte(addr)
	cpu.bus.WriteByte(addr, cpu.rr(v))
}

//  SLA n
// Description:
//  Shift n left into Carry. LSB of n set to 0.
// Use with:
//  n = A,B,C,D,E,H,L,(HL)
// Flags affected:
//  Z - Set if result is zero.
//  N - Reset.
//  H - Reset.
//  C - Contains old bit 7 data.
func (cpu *CPU) sla(v byte) byte {
	shifted := v << 1
	cpu.clearFlag(N)
	cpu.clearFlag(H)
	cpu.applyZeroBy(shifted)
	if v&0x80 == 0x80 {
		cpu.setFlag(C)
	} else {
		cpu.clearFlag(C)
	}
	return shifted
}

func (cpu *CPU) sla_n(r *types.Register) {
	*r = cpu.sla(*r)
}

func (cpu *CPU) sla_hl() {
	addr := cpu.getHL()
	v := cpu.bus.ReadByte(addr)
	cpu.bus.WriteByte(addr, cpu.sla(v))
}

// SRA n
// Description:
//  Shift n right into Carry. MSB doesn't change.
// Use with:
//  n = A,B,C,D,E,H,L,(HL)
// Flags affected:
//  Z - Set if result is zero.
//  N - Reset.
//  H - Reset.
//  C - Contains old bit 0 data.
func (cpu *CPU) sra(v byte) byte {
	shifted := (v >> 1) | (v & 0x80)
	cpu.clearFlag(N)
	cpu.clearFlag(H)
	cpu.applyZeroBy(shifted)
	if v&0x01 == 0x01 {
		cpu.setFlag(C)
	} else {
		cpu.clearFlag(C)
	}
	return shifted
}

func (cpu *CPU) sra_n(r *types.Register) {
	*r = cpu.sra(*r)
}

func (cpu *CPU) sra_hl() {
	addr := cpu.getHL()
	v := cpu.bus.ReadByte(addr)
	cpu.bus.WriteByte(addr, cpu.sra(v))
}

// SWAP n
// Description:
//  Swap upper & lower nibles of n.
// Use with:
//  n = A,B,C,D,E,H,L,(HL)
// Flags affected:
//  Z - Set if result is zero.
//  N - Reset.
//  H - Reset.
//  C - Reset.
func (cpu *CPU) swap(v byte) byte {
	upper := (v & 0xF0) >> 4
	lower := (v & 0x0F) << 4
	computed := upper ^ lower
	cpu.clearFlag(N)
	cpu.clearFlag(H)
	cpu.clearFlag(C)
	cpu.applyZeroBy(computed)
	return computed
}

func (cpu *CPU) swap_n(r *types.Register) {
	*r = cpu.swap(*r)
}

func (cpu *CPU) swap_hl() {
	addr := cpu.getHL()
	v := cpu.bus.ReadByte(addr)
	cpu.bus.WriteByte(addr, cpu.swap(v))
}

// SRL n
// Description:
//  Shift n right into Carry. MSB set to 0.
// Use with:
//  n = A,B,C,D,E,H,L,(HL)
// Flags affected:
//  Z - Set if result is zero.
//  N - Reset.
//  H - Reset.
//  C - Contains old bit 0 data.
func (cpu *CPU) srl(v byte) byte {
	shifted := v >> 1
	cpu.clearFlag(N)
	cpu.clearFlag(H)
	cpu.applyZeroBy(shifted)
	if v&0x01 == 0x01 {
		cpu.setFlag(C)
	} else {
		cpu.clearFlag(C)
	}
	return shifted
}

func (cpu *CPU) srl_n(r *types.Register) {
	*r = cpu.srl(*r)
}

func (cpu *CPU) srl_hl() {
	addr := cpu.getHL()
	v := cpu.bus.ReadByte(addr)
	cpu.bus.WriteByte(addr, cpu.srl(v))
}

// BIT b,r
// Description:
//  Test bit b in register r.
// Use with:
//  b = 0 - 7, r = A,B,C,D,E,H,L,(HL)
// Flags affected:
//  Z - Set if bit b of register r is 0.
//  N - Reset.
//  H - Set.
//  C - Not affected.
func (cpu *CPU) testBit(bit types.Bit, a byte) {
	cpu.applyZeroBy(a & byte(bit))
	cpu.clearFlag(N)
	cpu.setFlag(H)
}

func (cpu *CPU) bit_b_r(b types.Bit, r *byte) {
	cpu.testBit(b, *r)
}

func (cpu *CPU) bit_b_hl(b types.Bit) {
	addr := cpu.getHL()
	v := cpu.bus.ReadByte(addr)
	cpu.testBit(b, v)
}

// RES b,r
// Description:
// Reset bit b in register r.
// Use with:
// b = 0 - 7, r = A,B,C,D,E,H,L,(HL)
// Flags affected: None.
func (cpu *CPU) res_b(b types.Bit, a byte) byte {
	return a & ^byte(b)
}

func (cpu *CPU) res_b_r(b types.Bit, r *types.Register) {
	*r = cpu.res_b(b, *r)
}

func (cpu *CPU) res_b_hl(b types.Bit) {
	addr := cpu.getHL()
	v := cpu.bus.ReadByte(addr)
	cpu.bus.WriteByte(addr, cpu.res_b(b, v))
}

// SET b,r
// Description:
// Set bit b in register r.
// Use with:
// b = 0 - 7, r = A,B,C,D,E,H,L,(HL)
// Flags affected: None.
func (cpu *CPU) set_b(b types.Bit, a byte) byte {
	return a | byte(b)
}

func (cpu *CPU) set_b_r(b types.Bit, r *types.Register) {
	*r = cpu.set_b(b, *r)
}

func (cpu *CPU) set_b_hl(b types.Bit) {
	addr := cpu.getHL()
	v := cpu.bus.ReadByte(addr)
	cpu.bus.WriteByte(addr, cpu.set_b(b, v))
}

func (cpu *CPU) resolveIRQ() bool {
	if !cpu.irq.Enabled() || !cpu.irq.HasIRQ() {
		return false
	}
	cpu.pushPC()
	addr := cpu.irq.ResolveISRAddr()
	if addr == nil {
		return false
	}
	cpu.PC = *addr
	cpu.irq.Disable()
	return true
}

// For Debugging
func (cpu *CPU) GetRegisters() Registers {
	return cpu.Regs
}
