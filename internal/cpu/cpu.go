package cpu

import (
	"errors"
)

const initialPC = 0x200

// CPU implements the CHIP-8 CPU
type CPU struct {
	mem   [4096]byte
	vmem  [64 * 32]bool
	stack [16]uint16
	keys  [16]bool

	PC uint16
	V  [16]byte
	I  uint16
	DT byte
	ST byte

	opcode uint16
	sp     byte

	draw    bool
	keyWait bool
	keyReg  byte

	quirkLoadStore    bool
	quirkShift        bool
	quirkJump         bool
	quirkVfOrder      bool
	quirkDraw         bool
	quirkPartialWrapH bool
	quirkPartialWrapV bool

	fontset []byte
}

var fontset = [...]byte{
	0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
	0x20, 0x60, 0x20, 0x20, 0x70, // 1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
	0x90, 0x90, 0xF0, 0x10, 0x10, // 4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
	0xF0, 0x10, 0x20, 0x40, 0x40, // 7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
	0xF0, 0x90, 0xF0, 0x90, 0x90, // A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
	0xF0, 0x80, 0x80, 0x80, 0xF0, // C
	0xE0, 0x90, 0x90, 0x90, 0xE0, // D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
	0xF0, 0x80, 0xF0, 0x80, 0x80, // F
}

// NewCPU creates a new CPU instance
func NewCPU() *CPU {
	cpu := new(CPU)
	cpu.PC = initialPC
	cpu.draw = true
	cpu.quirkLoadStore = true
	cpu.quirkShift = true
	cpu.quirkJump = true
	cpu.quirkVfOrder = true
	cpu.quirkDraw = true

	// Load font set
	copy(cpu.mem[0:len(fontset)], fontset[:])

	return cpu
}

// LoadRom loads the given ROM into the memory
func (cpu *CPU) LoadRom(prog []byte) error {
	if len(prog) <= len(cpu.mem)-0x200 {
		copy(cpu.mem[0x200:0x200+len(prog)], prog[:])
		cpu.PC = initialPC
		cpu.sp = 0
		return nil
	}

	return errors.New("invalid ROM")
}

func (cpu *CPU) emulateCycle() error {
	// Fetch opcode
	cpu.opcode = uint16(cpu.mem[cpu.PC])<<8 | uint16(cpu.mem[cpu.PC+1])

	// Decode opcode
	h := byte((cpu.opcode & 0xF000) >> 12)
	x := byte((cpu.opcode & 0x0F00) >> 8)
	y := byte((cpu.opcode & 0x00F0) >> 4)
	n := byte(cpu.opcode & 0x000F)
	nn := byte(cpu.opcode & 0x00FF)
	nnn := cpu.opcode & 0x0FFF

	// Execute opcode
	switch h {
	case 0:
		switch nn {
		case 0xE0:
			cpu.opcode0x00E0()
		case 0xEE:
			cpu.opcode0x00EE()
		default:
			cpu.opcode0x0NNN()
		}
	case 1:
		cpu.opcode0x1NNN(nnn)
	case 2:
		cpu.opcode0x2NNN(nnn)
	case 3:
		cpu.opcode0x3XNN(x, nn)
	case 4:
		cpu.opcode0x4XNN(x, nn)
	case 5:
		switch n {
		case 0:
			cpu.opcode0x5XNN(x, y)
		default:
			cpu.opcodeInvalid()
		}
	case 6:
		cpu.opcode0x6XNN(x, nn)
	case 7:
		cpu.opcode0x7XNN(x, nn)
	case 8:
		switch n {
		case 0:
			cpu.opcode0x8XY0(x, y)
		case 1:
			cpu.opcode0x8XY1(x, y)
		case 2:
			cpu.opcode0x8XY2(x, y)
		case 3:
			cpu.opcode0x8XY3(x, y)
		case 4:
			cpu.opcode0x8XY4(x, y)
		case 5:
			cpu.opcode0x8XY5(x, y)
		case 6:
			cpu.opcode0x8XY6(x, y)
		case 7:
			cpu.opcode0x8XY7(x, y)
		case 0xE:
			cpu.opcode0x8XYE(x, y)
		default:
			cpu.opcodeInvalid()
		}
	case 9:
		cpu.opcode0x9XY0(x, y)
	case 0xA:
		cpu.opcode0xANNN(nnn)
	case 0xB:
		cpu.opcode0xBNNN(nnn)
	case 0xC:
		cpu.opcode0xCXNN(x, nn)
	case 0xD:
		cpu.opcode0xDXYN(x, y, n)
	case 0xE:
		switch nn {
		case 0x9E:
			cpu.opcode0xEX9E(x)
		case 0xA1:
			cpu.opcode0xEXA1(x)
		default:
			cpu.opcodeInvalid()
		}
	case 0xF:
		switch nn {
		case 0x07:
			cpu.opcode0xFX07(x)
		case 0x0A:
			cpu.opcode0xFX0A(x)
		case 0x15:
			cpu.opcode0xFX15(x)
		case 0x18:
			cpu.opcode0xFX18(x)
		case 0x1E:
			cpu.opcode0xFX1E(x)
		case 0x29:
			cpu.opcode0xFX29(x)
		case 0x33:
			cpu.opcode0xFX33(x)
		case 0x55:
			cpu.opcode0xFX55(x)
		case 0x65:
			cpu.opcode0xFX65(x)
		default:
			cpu.opcodeInvalid()
		}
	default:
		cpu.opcodeInvalid()
	}

	return nil
}

func (cpu *CPU) drawSprite(x, y byte, height byte) {
	// Wrap around
	x %= 64
	y %= 32

	sprite := cpu.mem[cpu.I : cpu.I+uint16(height)]
	collision := false

	for i := byte(0); i < byte(len(sprite)); i++ {
		curY := y + i
		// Clip or wrap
		if curY >= 32 {
			if cpu.quirkPartialWrapV {
				curY %= 32
			} else {
				continue
			}
		}

		for j := byte(0); j < 8; j++ {
			curX := x + j
			// Clip or wrap
			if curX >= 64 {
				if cpu.quirkPartialWrapH {
					curX %= 64
				} else {
					continue
				}
			}

			idx := cpu.getVmemIndex(curX, curY)
			bit := (sprite[i]>>j)&0b1 > 0

			// Detect collision and draw pixel
			if bit && cpu.vmem[idx] {
				collision = true
			}
			res := cpu.vmem[idx] != bit
			cpu.vmem[idx] = res
		}
	}

	cpu.V[0xF] = 0
	if collision {
		cpu.V[0xF] = 1
	}
}

func (cpu *CPU) getVmemIndex(x, y byte) uint16 {
	return (uint16(y) * 64) + uint16(x)
}
