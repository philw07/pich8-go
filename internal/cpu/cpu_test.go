package cpu

import (
	"testing"

	"github.com/philw07/pich8-go/internal/videomemory"
	"github.com/stretchr/testify/assert"
)

func TestInitialization(t *testing.T) {
	assert := assert.New(t)

	cpu := NewCPU()
	assert.EqualValues(0x200, cpu.PC)
	assert.EqualValues(fontset[:], cpu.mem[:len(fontset)])
	assert.EqualValues(fontsetBig[:], cpu.mem[0x50:0x50+len(fontsetBig)])
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

	// 0x0NNN
	cpu = NewCPU()
	cpu.LoadRom([]byte{0x03, 0x33})
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

func TestOpcodesSChip(t *testing.T) {
	assert := assert.New(t)

	// 0x00FD
	cpu := NewCPU()
	cpu.LoadRom([]byte{0x00, 0xFD})
	cpu.emulateCycle()
	assert.EqualValues(0x200, cpu.PC)
	assert.EqualValues([]byte{0x12, 0x00}, cpu.mem[0x200:0x202])

	// 0x00FF & 0x00FE
	cpu = NewCPU()
	cpu.LoadRom([]byte{0x00, 0xFF, 0x00, 0xFE})
	cpu.emulateCycle()
	assert.EqualValues(videomemory.ExtendedVideoMode, cpu.vmem.VideoMode)
	assert.EqualValues(0x202, cpu.PC)
	cpu.emulateCycle()
	assert.EqualValues(videomemory.DefaultVideoMode, cpu.vmem.VideoMode)
	assert.EqualValues(0x204, cpu.PC)

	// 0xDXYN
	cpu = NewCPU()
	cpu.LoadRom([]byte{0x00, 0xFF, 0xD0, 0x10})
	cpu.emulateCycle()
	assert.EqualValues(videomemory.ExtendedVideoMode, cpu.vmem.VideoMode)
	assert.EqualValues(0x202, cpu.PC)
	cpu.V[0] = 65
	cpu.V[1] = 2
	cpu.I = 0x300
	for i := 0x300; i < 0x320; i++ {
		cpu.mem[i] = 0xFF
	}
	cpu.emulateCycle()
	for x := 65; x < 81; x++ {
		for y := 2; y < 18; y++ {
			assert.True(cpu.vmem.Get(cpu.vmem.Plane, x, y))
		}
	}
	assert.EqualValues(0, cpu.V[0xF])
	assert.True(cpu.draw)
	assert.EqualValues(0x204, cpu.PC)

	// 0xFX30
	cpu = NewCPU()
	cpu.LoadRom([]byte{0xF0, 0x30})
	for i := 0; i < 10; i++ {
		cpu.V[0] = byte(i)
		cpu.emulateCycle()
		assert.EqualValues(0x50+i*10, cpu.I)
		cpu.PC -= 2
	}

	// 0xFX75
	reg := []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xFF, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	cpu = NewCPU()
	cpu.LoadRom([]byte{0xF4, 0x75})
	copy(cpu.V[:], reg)
	cpu.emulateCycle()
	assert.EqualValues(cpu.V[:5], cpu.RPL[:5])
	assert.Zero(cpu.RPL[5])
	assert.EqualValues(0x202, cpu.PC)

	// 0xFX85
	cpu = NewCPU()
	cpu.LoadRom([]byte{0xF4, 0x85})
	copy(cpu.RPL[:], reg)
	cpu.emulateCycle()
	assert.EqualValues(cpu.RPL[:5], cpu.V[:5])
	assert.Zero(cpu.V[5])
	assert.EqualValues(0x202, cpu.PC)
}

func TestOpcodesXOChip(t *testing.T) {
	assert := assert.New(t)

	regs := []byte{1, 2, 3, 4, 5}

	// 0x5XY2
	cpu := NewCPU()
	cpu.LoadRom([]byte{0x51, 0x52})
	copy(cpu.V[1:6], regs)
	cpu.I = 0x300
	cpu.emulateCycle()
	assert.EqualValues(regs, cpu.mem[0x300:0x305])
	assert.EqualValues(0x202, cpu.PC)

	cpu = NewCPU()
	cpu.LoadRom([]byte{0x58, 0x42})
	copy(cpu.V[4:9], regs)
	cpu.I = 0x300
	cpu.emulateCycle()
	assert.EqualValues(regs, cpu.mem[0x300:0x305])
	assert.EqualValues(0x202, cpu.PC)

	// 0x5XY3
	cpu = NewCPU()
	cpu.LoadRom([]byte{0x51, 0x53})
	copy(cpu.mem[0x300:0x306], regs)
	cpu.I = 0x300
	cpu.emulateCycle()
	assert.EqualValues(regs, cpu.V[1:6])
	assert.EqualValues(0x202, cpu.PC)

	cpu = NewCPU()
	cpu.LoadRom([]byte{0x5F, 0xB3})
	copy(cpu.mem[0x300:0x306], regs)
	cpu.I = 0x300
	cpu.emulateCycle()
	assert.EqualValues(regs, cpu.V[0xB:0x10])
	assert.EqualValues(0x202, cpu.PC)

	// 0xF000 NNNN
	cpu = NewCPU()
	cpu.LoadRom([]byte{0xF0, 0x00, 0xFE, 0xDC})
	cpu.emulateCycle()
	assert.EqualValues(0xFEDC, cpu.I)
	assert.EqualValues(0x204, cpu.PC)

	// 0xFN01
	cpu = NewCPU()
	cpu.LoadRom([]byte{0xF0, 0x01, 0xF1, 0x01, 0xF2, 0x01, 0xF3, 0x01})
	cpu.emulateCycle()
	assert.EqualValues(0x202, cpu.PC)
	assert.EqualValues(videomemory.NoPlane, cpu.vmem.Plane)
	cpu.emulateCycle()
	assert.EqualValues(0x204, cpu.PC)
	assert.EqualValues(videomemory.FirstPlane, cpu.vmem.Plane)
	cpu.emulateCycle()
	assert.EqualValues(0x206, cpu.PC)
	assert.EqualValues(videomemory.SecondPlane, cpu.vmem.Plane)
	cpu.emulateCycle()
	assert.EqualValues(0x208, cpu.PC)
	assert.EqualValues(videomemory.BothPlanes, cpu.vmem.Plane)

	// 0xF002
	buf := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0xA, 0xB, 0xC, 0xD, 0xE, 0xF}
	cpu = NewCPU()
	cpu.LoadRom([]byte{0xF0, 0x02})
	copy(cpu.mem[0x300:0x310], buf)
	cpu.I = 0x300
	cpu.emulateCycle()
	for i := range buf {
		assert.EqualValues(buf[i], cpu.audioBuffer[i])
	}
	assert.EqualValues(0x202, cpu.PC)

	// Skip with 4 byte opcode
	cpu = NewCPU()
	cpu.LoadRom([]byte{0x30, 0x00, 0xF0, 0x00, 0x12, 0x34, 0x12, 0x00})
	cpu.emulateCycle()
	assert.EqualValues(0x206, cpu.PC)
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
