package cpu

import (
	"errors"
	"math/rand"
)

func (cpu *CPU) opcodeInvalid() {
	cpu.PC += 2
}

// 0x00E0 - Clear display
func (cpu *CPU) opcode0x00E0() {
	cpu.vmem.Clear()
	cpu.draw = true
	cpu.PC += 2
}

// 0x00EE - Return from subroutine
func (cpu *CPU) opcode0x00EE() {
	cpu.sp--
	cpu.PC = cpu.stack[cpu.sp] + 2
}

// 0x0NNN - Legacy SYS call, ignored
func (cpu *CPU) opcode0x0NNN() {
	cpu.PC += 2
}

// 0x1NNN - Goto nnn
func (cpu *CPU) opcode0x1NNN(nnn uint16) {
	cpu.PC = nnn
}

// 0x2NNN - Call subroutine at nnn
func (cpu *CPU) opcode0x2NNN(nnn uint16) error {
	if int(cpu.sp) >= len(cpu.stack) {
		return errors.New("stack overflow")
	}

	cpu.stack[cpu.sp] = cpu.PC
	cpu.sp++
	cpu.PC = nnn
	return nil
}

// 0x3XNN - Skip next instruction if Vx == nn
func (cpu *CPU) opcode0x3XNN(x, nn byte) {
	if cpu.V[x] == nn {
		cpu.PC += 2
	}
	cpu.PC += 2
}

// 0x4XNN - Skip next instruction if Vx != nn
func (cpu *CPU) opcode0x4XNN(x, nn byte) {
	if cpu.V[x] != nn {
		cpu.PC += 2
	}
	cpu.PC += 2
}

// 0x5XY0 - Skip next instruction if Vx == Vy
func (cpu *CPU) opcode0x5XNN(x, y byte) {
	if cpu.V[x] == cpu.V[y] {
		cpu.PC += 2
	}
	cpu.PC += 2
}

// 0x6XNN - Vx = nn
func (cpu *CPU) opcode0x6XNN(x, nn byte) {
	cpu.V[x] = nn
	cpu.PC += 2
}

// 0x7XNN - Vx += nn
func (cpu *CPU) opcode0x7XNN(x, nn byte) {
	res := uint16(cpu.V[x]) + uint16(nn)
	cpu.V[x] = byte(res)
	cpu.PC += 2
}

// 0x8XY0 - Vx = Vy
func (cpu *CPU) opcode0x8XY0(x, y byte) {
	cpu.V[x] = cpu.V[y]
	cpu.PC += 2
}

// 0x8XY1 - Vx |= Vy
func (cpu *CPU) opcode0x8XY1(x, y byte) {
	cpu.V[x] |= cpu.V[y]
	cpu.PC += 2
}

// 0x8XY2 - Vx &= Vy
func (cpu *CPU) opcode0x8XY2(x, y byte) {
	cpu.V[x] &= cpu.V[y]
	cpu.PC += 2
}

// 0x8XY3 - Vx ^= Vy
func (cpu *CPU) opcode0x8XY3(x, y byte) {
	cpu.V[x] ^= cpu.V[y]
	cpu.PC += 2
}

// 0x8XY4 - Vx += Vy
func (cpu *CPU) opcode0x8XY4(x, y byte) {
	res := uint16(cpu.V[x]) + uint16(cpu.V[y])
	vf := byte(0)
	if res > 0xFF {
		vf = 1
	}
	cpu.writeVf(x, byte(res), vf)
	cpu.PC += 2
}

// 0x8XY5 - Vx -= Vy
func (cpu *CPU) opcode0x8XY5(x, y byte) {
	res := int16(cpu.V[x]) - int16(cpu.V[y])
	vf := byte(1)
	if res < 0 {
		vf = 0
	}
	cpu.writeVf(x, byte(res), vf)
	cpu.PC += 2
}

// 0x8XY6 - Bitshift right
// Original: Vx = Vy >> 1
// Quirk:    Vx >>= 1
func (cpu *CPU) opcode0x8XY6(x, y byte) {
	if cpu.quirkShift {
		cpu.writeVf(x, cpu.V[x]>>1, cpu.V[x]&1)
	} else {
		cpu.writeVf(x, cpu.V[y]>>1, cpu.V[y]&1)
	}
	cpu.PC += 2
}

// 0x8XY7 - Vx = Vy - Vx
func (cpu *CPU) opcode0x8XY7(x, y byte) {
	res := int16(cpu.V[y]) - int16(cpu.V[x])
	vf := byte(1)
	if res < 0 {
		vf = 0
	}
	cpu.writeVf(x, byte(res), vf)
	cpu.PC += 2
}

// 0x8XYE - Bitshift left
// Original: Vx = Vy << 1
// Quirk:    Vx <<= 1
func (cpu *CPU) opcode0x8XYE(x, y byte) {
	if cpu.quirkShift {
		cpu.writeVf(x, cpu.V[x]<<1, (cpu.V[x]&0x80)>>7)
	} else {
		cpu.writeVf(x, cpu.V[y]<<1, (cpu.V[x]&0x80)>>7)
	}
	cpu.PC += 2
}

// 0x9XY0 - Skip next instruction if Vx != Vy
func (cpu *CPU) opcode0x9XY0(x, y byte) {
	if cpu.V[x] != cpu.V[y] {
		cpu.PC += 2
	}
	cpu.PC += 2
}

// 0xANNN - I = nnn
func (cpu *CPU) opcode0xANNN(nnn uint16) {
	cpu.I = nnn
	cpu.PC += 2
}

// 0xBNNN - PC = nnn + V0
// Original: PC = nnn + V0
// Quirk:    PC = xnn + Vx
func (cpu *CPU) opcode0xBNNN(nnn uint16) {
	cpu.PC = nnn
	if cpu.quirkJump {
		cpu.PC += uint16(cpu.V[(nnn >> 8 & 0xF)])
	} else {
		cpu.PC += uint16(cpu.V[0])
	}
}

// 0xCXNN - Vx = rand() & nn
func (cpu *CPU) opcode0xCXNN(x, nn byte) {
	cpu.V[x] = byte(rand.Uint32()) & nn
	cpu.PC += 2
}

// 0xDXYN - draw(Vx, Vy, n)
func (cpu *CPU) opcode0xDXYN(x, y, n byte) {
	cpu.drawSprite(cpu.V[x], cpu.V[y], n)
	cpu.draw = true
	cpu.PC += 2
}

// 0xEX9E - Skip next instruction if key(Vx) is pressed
func (cpu *CPU) opcode0xEX9E(x byte) {
	if cpu.keys[cpu.V[x]] {
		cpu.PC += 2
	}
	cpu.PC += 2
}

// 0xEXA1 - Skip next instruction if key(Vx) is not pressed
func (cpu *CPU) opcode0xEXA1(x byte) {
	if !cpu.keys[cpu.V[x]] {
		cpu.PC += 2
	}
	cpu.PC += 2
}

// 0xFX07 - Vx = DT
func (cpu *CPU) opcode0xFX07(x byte) {
	cpu.V[x] = cpu.DT
	cpu.PC += 2
}

// 0xFX0A - Vx = get_key();
func (cpu *CPU) opcode0xFX0A(x byte) {
	cpu.keyWait = true
	cpu.keyReg = x
	cpu.PC += 2
}

// 0xFX15 - DT = Vx
func (cpu *CPU) opcode0xFX15(x byte) {
	cpu.DT = cpu.V[x]
	cpu.PC += 2
}

// 0xFX18 - ST = Vx
func (cpu *CPU) opcode0xFX18(x byte) {
	cpu.ST = cpu.V[x]
	cpu.PC += 2
}

// 0xFX1E - I += Vx
func (cpu *CPU) opcode0xFX1E(x byte) {
	cpu.I += uint16(cpu.V[x])
	cpu.PC += 2
}

// 0xFX29 - I = sprite_add(Vx)
func (cpu *CPU) opcode0xFX29(x byte) {
	cpu.I = uint16(cpu.V[x]) * 5
	cpu.PC += 2
}

// 0xFX33 - set_BCD(Vx)
func (cpu *CPU) opcode0xFX33(x byte) {
	hundreds := cpu.V[x] / 100
	tens := (cpu.V[x] % 100) / 10
	ones := cpu.V[x] % 10
	cpu.mem[cpu.I] = hundreds
	cpu.mem[cpu.I+1] = tens
	cpu.mem[cpu.I+2] = ones
	cpu.PC += 2
}

// 0xFX55 - reg_dump(Vx, &I)
// Original: I is incremented
// Quirk:    I is not incremented
func (cpu *CPU) opcode0xFX55(x byte) {
	start := cpu.I
	end := cpu.I + uint16(x)
	copy(cpu.mem[start:end+1], cpu.V[:x+1])
	if !cpu.quirkLoadStore {
		cpu.I += uint16(x) + 1
	}
	cpu.PC += 2
}

// 0xFX65 - reg_load(Vx, &I)
// Original: I is incremented
// Quirk:    I is not incremented
func (cpu *CPU) opcode0xFX65(x byte) {
	start := cpu.I
	end := cpu.I + uint16(x)
	copy(cpu.V[:x+1], cpu.mem[start:end+1])
	if !cpu.quirkLoadStore {
		cpu.I += uint16(x) + 1
	}
	cpu.PC += 2
}

func (cpu *CPU) writeVf(reg, value, vf byte) {
	if cpu.quirkVfOrder {
		cpu.V[reg] = value
		cpu.V[0xF] = vf
	} else {
		cpu.V[0xF] = vf
		cpu.V[reg] = value
	}
}
