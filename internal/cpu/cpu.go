package cpu

import (
	"errors"

	"github.com/philw07/pich8-go/internal/emulator"
)

const initialPC = 0x200

// CPU implements the CHIP-8 CPU
type CPU struct {
	mem         [4096]byte
	vmem        *emulator.VideoMemory
	stack       [16]uint16
	keys        [16]bool
	audioBuffer *[16]byte

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
	cpu.vmem = emulator.NewVideoMemory()
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

	bigSprite := (cpu.vmem.VideoMode == emulator.ExtendedVideoMode || cpu.quirkDraw) && height == 0
	step := 1
	width := 8
	if bigSprite {
		step = 2
		width = 16
	}
	if height == 0 {
		height = 16
	}

	collision := false
	i := int(cpu.I)
	length := width / 8 * int(height)

	for _, plane := range [...]emulator.Plane{emulator.FirstPlane, emulator.SecondPlane} {
		if cpu.vmem.Plane == plane || cpu.vmem.Plane == emulator.BothPlanes {
			sprite := cpu.mem[i : i+length]
			i += length

			for k := 0; k < len(sprite); k += step {
				curY := int(y) + k
		// Clip or wrap
				if curY >= cpu.vmem.Height() {
			if cpu.quirkPartialWrapV {
						curY %= cpu.vmem.Height()
			} else {
				continue
			}
		}

				for j := 0; j < width; j++ {
					curX := int(x) + j
			// Clip or wrap
					if curX >= cpu.vmem.Width() {
				if cpu.quirkPartialWrapH {
							curX %= cpu.vmem.Width()
				} else {
					continue
				}
			}

					// Get bit
					revJ := width - 1 - j
					bit := sprite[k]>>revJ&0b1 > 0
					if width == 16 {
						bit = (uint16(sprite[k])<<8|uint16(sprite[k+1]))>>revJ&0b1 > 0
					}

			// Detect collision and draw pixel
					if bit && cpu.vmem.Get(plane, curX, curY) {
				collision = true
			}
					res := cpu.vmem.Get(plane, curX, curY) != bit
					cpu.vmem.Set(plane, curX, curY, res)
				}
			}
		}
	}

	cpu.V[0xF] = 0
	if collision {
		cpu.V[0xF] = 1
	}
}
