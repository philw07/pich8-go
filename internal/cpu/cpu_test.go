package cpu

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitialization(t *testing.T) {
	assert := assert.New(t)

	cpu := NewCPU()
	assert.EqualValues(0x200, cpu.PC)
	assert.EqualValues(fontset[:], cpu.mem[:len(fontset)])
}

func TestLoadRom(t *testing.T) {
	assert := assert.New(t)

	prog := []byte{1, 2, 3, 4, 5, 6, 7, 8}

	cpu := NewCPU()
	cpu.LoadRom(prog)
	assert.EqualValues(prog, cpu.mem[0x200:0x208])
	assert.EqualValues(0x200, cpu.PC)
	assert.EqualValues(0, cpu.sp)
}

func TestOpcodes(t *testing.T) {
	assert := assert.New(t)

	// Invalid
	cpu := NewCPU()
	cpu.LoadRom([]byte{0xF0, 0xFF})
	cpu.emulateCycle()
	assert.EqualValues(0x202, cpu.PC)

	// 0x00E0
	cpu = NewCPU()
	cpu.LoadRom([]byte{0x00, 0xE0})
	cpu.vmem.SetAll(true)
	cpu.emulateCycle()
	for i := 0; i < cpu.vmem.Width()*cpu.vmem.Height(); i++ {
		assert.False(cpu.vmem.GetIndex(cpu.vmem.Plane, i))
	}
	assert.True(cpu.draw)
	assert.EqualValues(0x202, cpu.PC)

	// 0x1NNN
	cpu = NewCPU()
	cpu.LoadRom([]byte{0x12, 0xE0})
	cpu.emulateCycle()
	assert.EqualValues(0x2E0, cpu.PC)
	assert.EqualValues(0, cpu.sp)

	// 0x2NNN
	cpu = NewCPU()
	cpu.LoadRom([]byte{0x22, 0xE0})
	cpu.emulateCycle()
	assert.EqualValues(0x2E0, cpu.PC)
	assert.EqualValues(1, cpu.sp)
	assert.EqualValues(0x200, cpu.stack[0])

	// 0x3XNN - Equal
	cpu = NewCPU()
	cpu.LoadRom([]byte{0x30, 0x12})
	cpu.V[0] = 0x12
	cpu.emulateCycle()
	assert.EqualValues(0x204, cpu.PC)
	// 0x3XNN - Not equal
	cpu = NewCPU()
	cpu.LoadRom([]byte{0x30, 0x12})
	cpu.V[0] = 0x23
	cpu.emulateCycle()
	assert.EqualValues(0x202, cpu.PC)

	// 0x4XNN - Equal
	cpu = NewCPU()
	cpu.LoadRom([]byte{0x40, 0x12})
	cpu.V[0] = 0x12
	cpu.emulateCycle()
	assert.EqualValues(0x202, cpu.PC)
	// 0x4XNN - Not equal
	cpu = NewCPU()
	cpu.LoadRom([]byte{0x40, 0x12})
	cpu.V[0] = 0x23
	cpu.emulateCycle()
	assert.EqualValues(0x204, cpu.PC)

	// 0x5XY0 - Equal
	cpu = NewCPU()
	cpu.LoadRom([]byte{0x50, 0x10})
	cpu.V[0] = 0x45
	cpu.V[1] = 0x45
	cpu.emulateCycle()
	assert.EqualValues(0x204, cpu.PC)
	// 0x5XY0 - Not equal
	cpu = NewCPU()
	cpu.LoadRom([]byte{0x50, 0x10})
	cpu.V[0] = 0x45
	cpu.V[1] = 0x52
	cpu.emulateCycle()
	assert.EqualValues(0x202, cpu.PC)

	// 0x6XNN
	testArithmetic(assert, 0x60AB, 0x11, 0x0, 0xAB)

	// 0x7XNN
	testArithmetic(assert, 0x70AB, 0x01, 0x00, 0xAC)

	// 0x9XY0 - Equal
	cpu = NewCPU()
	cpu.LoadRom([]byte{0x90, 0x10})
	cpu.V[0] = 0x45
	cpu.V[1] = 0x45
	cpu.emulateCycle()
	assert.EqualValues(0x202, cpu.PC)
	// 0x9XY0 - Not equal
	cpu = NewCPU()
	cpu.LoadRom([]byte{0x90, 0x10})
	cpu.V[0] = 0x45
	cpu.V[1] = 0x52
	cpu.emulateCycle()
	assert.EqualValues(0x204, cpu.PC)

	// 0xANNN
	cpu = NewCPU()
	cpu.LoadRom([]byte{0xA1, 0x23})
	cpu.emulateCycle()
	assert.EqualValues(0x123, cpu.I)
	assert.EqualValues(0x202, cpu.PC)

	// 0xBNNN - Quirk
	cpu = NewCPU()
	cpu.LoadRom([]byte{0xB1, 0x23})
	cpu.quirkJump = true
	cpu.V[1] = 0x11
	cpu.emulateCycle()
	assert.EqualValues(0x134, cpu.PC)
	// 0xBNNN - No quirk
	cpu = NewCPU()
	cpu.LoadRom([]byte{0xB1, 0x23})
	cpu.quirkJump = false
	cpu.V[0] = 0x11
	cpu.emulateCycle()
	assert.EqualValues(0x134, cpu.PC)

	// 0xCNNN
	cpu = NewCPU()
	cpu.LoadRom([]byte{0xC0, 0x00, 0xC0, 0x0F, 0xC0, 0xF0})
	cpu.emulateCycle()
	assert.EqualValues(0, cpu.V[0])
	cpu.emulateCycle()
	assert.EqualValues(0, cpu.V[0]&0xF0)
	cpu.emulateCycle()
	assert.EqualValues(0, cpu.V[0]&0x0F)

	// 0xDXYN - Completely on screen, no clipping/wrapping
	cpu = NewCPU()
	cpu.LoadRom([]byte{0xD0, 0x15})
	cpu.V[0] = 7
	cpu.V[1] = 2
	cpu.I = 0x300
	copy(cpu.mem[0x300:0x305], []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF})
	cpu.emulateCycle()
	for y := 2; y < 7; y++ {
		for x := 7; x < 15; x++ {
			assert.EqualValues(true, cpu.vmem.Get(cpu.vmem.Plane, x, y))
		}
	}
	assert.EqualValues(0, cpu.V[0xF])
	assert.True(cpu.draw)
	assert.EqualValues(0x202, cpu.PC)
	// 0xDXYN - Wrapping when off screen
	cpu = NewCPU()
	cpu.LoadRom([]byte{0xD0, 0x15})
	cpu.V[0] = 71
	cpu.V[1] = 34
	cpu.I = 0x300
	copy(cpu.mem[0x300:0x305], []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF})
	cpu.emulateCycle()
	for y := 2; y < 7; y++ {
		for x := 7; x < 15; x++ {
			assert.EqualValues(true, cpu.vmem.Get(cpu.vmem.Plane, x, y))
		}
	}
	assert.EqualValues(0, cpu.V[0xF])
	assert.True(cpu.draw)
	assert.EqualValues(0x202, cpu.PC)
	// 0xDXYN - Clipping x and y
	cpu = NewCPU()
	cpu.LoadRom([]byte{0xD0, 0x15})
	cpu.V[0] = 60
	cpu.V[1] = 30
	cpu.I = 0x300
	copy(cpu.mem[0x300:0x305], []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF})
	cpu.emulateCycle()
	for y := 30; y < 35; y++ {
		curY := y % 32
		for x := 60; x < 68; x++ {
			curX := x % 64
			assert.EqualValues(curX >= 60 && curY >= 30, cpu.vmem.Get(cpu.vmem.Plane, curX, curY))
		}
	}
	assert.EqualValues(0, cpu.V[0xF])
	assert.True(cpu.draw)
	assert.EqualValues(0x202, cpu.PC)
	// 0xDXYN - Wrapping x, but not y
	cpu = NewCPU()
	cpu.LoadRom([]byte{0xD0, 0x15})
	cpu.quirkPartialWrapH = true
	cpu.V[0] = 60
	cpu.V[1] = 30
	cpu.I = 0x300
	copy(cpu.mem[0x300:0x305], []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF})
	cpu.emulateCycle()
	for y := 30; y < 35; y++ {
		curY := y % 32
		for x := 60; x < 68; x++ {
			curX := x % 64
			assert.EqualValues(curY >= 30, cpu.vmem.Get(cpu.vmem.Plane, curX, curY))
		}
	}
	assert.EqualValues(0, cpu.V[0xF])
	assert.True(cpu.draw)
	assert.EqualValues(0x202, cpu.PC)
	// 0xDXYN - Wrapping y, but not x
	cpu = NewCPU()
	cpu.LoadRom([]byte{0xD0, 0x15})
	cpu.quirkPartialWrapV = true
	cpu.V[0] = 60
	cpu.V[1] = 30
	cpu.I = 0x300
	copy(cpu.mem[0x300:0x305], []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF})
	cpu.emulateCycle()
	for y := 30; y < 35; y++ {
		curY := y % 32
		for x := 60; x < 68; x++ {
			curX := x % 64
			assert.EqualValues(curX >= 60, cpu.vmem.Get(cpu.vmem.Plane, curX, curY))
		}
	}
	assert.EqualValues(0, cpu.V[0xF])
	assert.True(cpu.draw)
	assert.EqualValues(0x202, cpu.PC)
	// 0xDXYN - Wrapping x and y
	cpu = NewCPU()
	cpu.LoadRom([]byte{0xD0, 0x15})
	cpu.quirkPartialWrapH = true
	cpu.quirkPartialWrapV = true
	cpu.V[0] = 60
	cpu.V[1] = 30
	cpu.I = 0x300
	copy(cpu.mem[0x300:0x305], []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF})
	cpu.emulateCycle()
	for y := 30; y < 35; y++ {
		curY := y % 32
		for x := 60; x < 68; x++ {
			curX := x % 64
			assert.EqualValues(true, cpu.vmem.Get(cpu.vmem.Plane, curX, curY))
		}
	}
	assert.EqualValues(0, cpu.V[0xF])
	assert.True(cpu.draw)
	assert.EqualValues(0x202, cpu.PC)
	// 0xDXYN - Collision
	cpu = NewCPU()
	cpu.LoadRom([]byte{0xD0, 0x15})
	cpu.quirkPartialWrapH = true
	cpu.V[0] = 7
	cpu.V[1] = 2
	cpu.I = 0x300
	copy(cpu.mem[0x300:0x305], []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF})
	for x := 7; x < 15; x++ {
		cpu.vmem.Set(cpu.vmem.Plane, x, 3, true)
	}
	cpu.emulateCycle()
	for y := 2; y < 7; y++ {
		for x := 7; x < 15; x++ {
			assert.EqualValues(y != 3, cpu.vmem.Get(cpu.vmem.Plane, x, y))
		}
	}
	assert.EqualValues(1, cpu.V[0xF])
	assert.True(cpu.draw)
	assert.EqualValues(0x202, cpu.PC)

	// 0xEX9E - Pressed
	cpu = NewCPU()
	cpu.LoadRom([]byte{0xE0, 0x9E})
	cpu.keys[3] = true
	cpu.V[0] = 3
	cpu.emulateCycle()
	assert.EqualValues(0x204, cpu.PC)
	// 0xEX9E - Not pressed
	cpu = NewCPU()
	cpu.LoadRom([]byte{0xE0, 0x9E})
	cpu.keys[3] = false
	cpu.V[0] = 3
	cpu.emulateCycle()
	assert.EqualValues(0x202, cpu.PC)

	// 0xEXA1 - Pressed
	cpu = NewCPU()
	cpu.LoadRom([]byte{0xE0, 0xA1})
	cpu.keys[3] = true
	cpu.V[0] = 3
	cpu.emulateCycle()
	assert.EqualValues(0x202, cpu.PC)
	// 0xEXA1 - Not pressed
	cpu = NewCPU()
	cpu.LoadRom([]byte{0xE0, 0xA1})
	cpu.keys[3] = false
	cpu.V[0] = 3
	cpu.emulateCycle()
	assert.EqualValues(0x204, cpu.PC)

	// 0xFX07
	cpu = NewCPU()
	cpu.LoadRom([]byte{0xF0, 0x07})
	cpu.DT = 0xAB
	cpu.emulateCycle()
	assert.EqualValues(cpu.DT, cpu.V[0])
	assert.EqualValues(0x202, cpu.PC)

	// 0xFX0A
	cpu = NewCPU()
	cpu.LoadRom([]byte{0xF5, 0x0A})
	cpu.emulateCycle()
	assert.True(cpu.keyWait)
	assert.EqualValues(5, cpu.keyReg)

	// 0xFX15
	cpu = NewCPU()
	cpu.LoadRom([]byte{0xF0, 0x15})
	cpu.V[0] = 0x15
	cpu.emulateCycle()
	assert.EqualValues(0x15, cpu.DT)

	// 0xFX18
	cpu = NewCPU()
	cpu.LoadRom([]byte{0xF0, 0x18})
	cpu.V[0] = 0x15
	cpu.emulateCycle()
	assert.EqualValues(0x15, cpu.ST)

	// 0xFX1E
	cpu = NewCPU()
	cpu.LoadRom([]byte{0xF0, 0x1E})
	cpu.V[0] = 2
	cpu.I = 0xAB
	cpu.emulateCycle()
	assert.EqualValues(0xAD, cpu.I)
	assert.EqualValues(0x202, cpu.PC)

	// 0xFX29
	cpu = NewCPU()
	cpu.LoadRom([]byte{0xF0, 0x29})
	for i := 0; i <= 0xF; i++ {
		cpu.V[0] = byte(i)
		cpu.emulateCycle()
		assert.EqualValues(i*5, cpu.I)
		cpu.PC -= 2
	}

	// 0xFX33
	cpu = NewCPU()
	cpu.LoadRom([]byte{0xF0, 0x33})
	cpu.I = 0x300
	cpu.V[0] = 194
	cpu.emulateCycle()
	assert.EqualValues(1, cpu.mem[cpu.I])
	assert.EqualValues(9, cpu.mem[cpu.I+1])
	assert.EqualValues(4, cpu.mem[cpu.I+2])

	// 0xFX55 - Quirk
	reg := []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xFF, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	cpu = NewCPU()
	cpu.LoadRom([]byte{0xF5, 0x55})
	cpu.quirkLoadStore = true
	cpu.I = 0x300
	copy(cpu.V[:], reg[:])
	cpu.emulateCycle()
	assert.EqualValues(reg[:5], cpu.mem[0x300:0x305])
	assert.EqualValues(0, cpu.mem[0x306])
	assert.EqualValues(0x202, cpu.PC)
	assert.EqualValues(0x300, cpu.I)
	// 0xFX55 - No quirk
	cpu = NewCPU()
	cpu.LoadRom([]byte{0xF5, 0x55})
	cpu.quirkLoadStore = false
	cpu.I = 0x300
	copy(cpu.V[:], reg[:])
	cpu.emulateCycle()
	assert.EqualValues(reg[:5], cpu.mem[0x300:0x305])
	assert.EqualValues(0, cpu.mem[0x306])
	assert.EqualValues(0x202, cpu.PC)
	assert.EqualValues(0x306, cpu.I)

	// 0xFX65 - Quirk
	prog := []byte{0xF5, 0x65, 0xA9, 0x87, 0x65, 0x43, 0x21, 0xFF}
	cpu = NewCPU()
	cpu.LoadRom(prog)
	cpu.quirkLoadStore = true
	cpu.I = 0x202
	cpu.emulateCycle()
	assert.EqualValues(prog[2:8], cpu.V[:6])
	assert.EqualValues(0, cpu.V[6])
	assert.EqualValues(0x202, cpu.PC)
	assert.EqualValues(0x202, cpu.I)
	// 0xFX55 - No quirk
	cpu = NewCPU()
	cpu.LoadRom(prog)
	cpu.quirkLoadStore = false
	cpu.I = 0x202
	cpu.emulateCycle()
	assert.EqualValues(prog[2:8], cpu.V[:6])
	assert.EqualValues(0, cpu.V[6])
	assert.EqualValues(0x202, cpu.PC)
	assert.EqualValues(0x208, cpu.I)
}

func TestOpcodesArithmetic(t *testing.T) {
	assert := assert.New(t)

	b1 := byte(0b11110010)
	b2 := byte(0b00001011)

	// 0x8XY0
	testArithmetic(assert, 0x8010, b1, b2, b2)

	// 0x8XY1
	testArithmetic(assert, 0x8011, b1, b2, b1|b2)

	// 0x8XY2
	testArithmetic(assert, 0x8012, b1, b2, b1&b2)

	// 0x8XY3
	testArithmetic(assert, 0x8013, b1, b2, b1^b2)

	// 0x8XY4 w/ carry
	testArithmeticV(assert, 0x8014, 0xFF, 1, 0, 1)
	// 0x8XY4 w/o carry
	testArithmeticV(assert, 0x8014, 0xFE, 1, 0xFF, 0)

	// 0x8XY5 w/ borrow
	testArithmeticV(assert, 0x8015, 2, 3, 0xFF, 0)
	// 0x8XY5 w/o borrow
	testArithmeticV(assert, 0x8015, 2, 1, 1, 1)

	// 0x8XY6 w/ 0 - Quirk
	testArithmeticV(assert, 0x8016, b1, b2, b1>>1, 0)
	// 0x8XY6 w/ 0 - No Quirk
	testArithmeticVNoQuirk(assert, 0x8016, b2, b1, b1>>1, 0)
	// 0x8XY6 w/ 1 - Quirk
	testArithmeticV(assert, 0x8016, b2, b1, b2>>1, 1)
	// 0x8XY6 w/ 1 - No Quirk
	testArithmeticVNoQuirk(assert, 0x8016, b1, b2, b2>>1, 1)

	// 0x8XY7 w/ borrow
	testArithmeticV(assert, 0x8017, 2, 3, 1, 1)
	// 0x8XY7 w/o borrow
	testArithmeticV(assert, 0x8017, 3, 2, 0xFF, 0)

	// 0x8XYE w/ 0 - Quirk
	testArithmeticV(assert, 0x801E, b1, b2, b1<<1, 1)
	// 0x8XYE w/ 0 - No quirk
	testArithmeticVNoQuirk(assert, 0x801E, b1, b2, b2<<1, 1)
	// 0x8XYE w/ 1 - Quirk
	testArithmeticV(assert, 0x801E, b2, b1, b2<<1, 0)
	// 0x8XYE w/ 1 - No quirk
	testArithmeticVNoQuirk(assert, 0x801E, b2, b1, b1<<1, 0)
}

func testArithmeticV(assert *assert.Assertions, opcode uint16, v1, v2, res, resv byte) {
	cpu := NewCPU()
	cpu.LoadRom([]byte{byte(opcode >> 8), byte(opcode)})
	cpu.V[0] = v1
	cpu.V[1] = v2
	cpu.emulateCycle()
	assert.EqualValues(res, cpu.V[0])
	assert.EqualValues(resv, cpu.V[0xF])
	assert.EqualValues(0x202, cpu.PC)
}

func testArithmeticVNoQuirk(assert *assert.Assertions, opcode uint16, v1, v2, res, resv byte) {
	cpu := NewCPU()
	cpu.LoadRom([]byte{byte(opcode >> 8), byte(opcode)})
	cpu.quirkLoadStore = false
	cpu.quirkShift = false
	cpu.V[0] = v1
	cpu.V[1] = v2
	cpu.emulateCycle()
	assert.EqualValues(res, cpu.V[0])
	assert.EqualValues(resv, cpu.V[0xF])
	assert.EqualValues(0x202, cpu.PC)
}

func testArithmetic(assert *assert.Assertions, opcode uint16, v1, v2, res byte) {
	cpu := NewCPU()
	cpu.LoadRom([]byte{byte(opcode >> 8), byte(opcode)})
	cpu.V[0] = v1
	cpu.V[1] = v2
	cpu.emulateCycle()
	assert.EqualValues(res, cpu.V[0])
	assert.EqualValues(0x202, cpu.PC)
}
