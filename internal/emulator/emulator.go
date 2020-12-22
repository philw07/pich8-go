package emulator

import (
	"time"

	"github.com/faiface/pixel/pixelgl"
	"github.com/philw07/pich8-go/internal/cpu"
)

const (
	timerFrequency = 60
	nanosPerTimer  = 1_000_000_000 / timerFrequency
)

// Emulator implements the CHIP-8 emulator
type Emulator struct {
	cpu      cpu.CPU
	cpuSpeed int
	display  Display
	input    [16]bool

	rom  []byte
	mute bool

	lastCycle           time.Time
	lastCorrectionCPU   time.Time
	counterCPU          int
	lastTimer           time.Time
	lastCorrectionTimer time.Time
	counterTimer        int
	pause               bool
	pauseTime           time.Time
}

// NewEmulator creates a new instance
func NewEmulator() (*Emulator, error) {
	disp, err := NewDisplay()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &Emulator{
		cpu:      *cpu.NewCPU(),
		cpuSpeed: 720,
		display:  *disp,

		lastCycle:           now,
		lastCorrectionCPU:   now,
		lastTimer:           now,
		lastCorrectionTimer: now,
	}, nil
}

func (emu *Emulator) reset() error {
	emu.cpu = *cpu.NewCPU()
	if err := emu.cpu.LoadRom(emu.rom); err != nil {
		return err
	}
	return nil
}

// LoadRom loads the given ROM into the emulator
func (emu *Emulator) LoadRom(rom []byte) error {
	emu.rom = rom
	return emu.reset()
}

func (emu *Emulator) setPause(pause bool) {
	emu.pause = pause
	if pause {
		// Store timestamp
		emu.pauseTime = time.Now()
	} else {
		// "Subtract" paused time so the emulation doesn't jump
		diff := time.Since(emu.pauseTime)
		emu.lastCycle = emu.lastCycle.Add(diff)
	}
}

// Run runs the main loop of the emulator
func (emu *Emulator) Run() {
	for !emu.display.Window.Closed() {
		// Handle input
		emu.handleInput()

		// Perform emulation
		emu.performEmulation()

		// Draw the frame
		emu.display.Draw(emu.cpu.Vmem())
	}
}

func (emu *Emulator) performEmulation() {
	if !emu.pause {
		// Emulate CPU cycles
		nanosPerCycle := 1_000_000_000 / emu.cpuSpeed
		if time.Since(emu.lastCycle).Nanoseconds() >= 10*int64(nanosPerCycle) {
			cycles := int(time.Since(emu.lastCycle).Nanoseconds()) / nanosPerCycle
			emu.lastCycle = time.Now()

			// Check if additional cycles are needed
			if time.Since(emu.lastCorrectionCPU).Seconds() >= 0.25 {
				target := emu.cpuSpeed / 4
				if emu.counterCPU < target {
					cycles += target - emu.counterCPU
				}
				emu.lastCorrectionCPU = time.Now()
				emu.counterCPU = 0
			} else {
				emu.counterCPU += cycles
			}

			for i := 0; i < int(cycles); i++ {
				emu.cpu.Tick(emu.input)
			}
		}

		// Update timers
		if time.Since(emu.lastTimer).Nanoseconds() >= nanosPerTimer {
			emu.lastTimer = time.Now()
			reps := 1

			// Check and correct frequency
			if time.Since(emu.lastCorrectionTimer).Seconds() >= 0.25 {
				target := timerFrequency / 4
				if emu.counterTimer+1 < target {
					reps += target - emu.counterTimer - 1
				}
				emu.lastCorrectionTimer = time.Now()
				emu.counterTimer = 0
			} else {
				emu.counterTimer += reps
			}

			for i := 0; i < reps; i++ {
				if emu.cpu.ST > 0 && !emu.mute {
					//TODO: Play sound
				}
				emu.cpu.UpdateTimers()
			}
		}
	}
}

func (emu *Emulator) handleInput() {
	// CHIP-8 keys
	emu.input[0] = emu.display.Window.Pressed(pixelgl.Key1)
	emu.input[1] = emu.display.Window.Pressed(pixelgl.Key2)
	emu.input[2] = emu.display.Window.Pressed(pixelgl.Key3)
	emu.input[3] = emu.display.Window.Pressed(pixelgl.Key4)
	emu.input[4] = emu.display.Window.Pressed(pixelgl.KeyQ)
	emu.input[5] = emu.display.Window.Pressed(pixelgl.KeyW)
	emu.input[6] = emu.display.Window.Pressed(pixelgl.KeyE)
	emu.input[7] = emu.display.Window.Pressed(pixelgl.KeyR)
	emu.input[8] = emu.display.Window.Pressed(pixelgl.KeyA)
	emu.input[9] = emu.display.Window.Pressed(pixelgl.KeyS)
	emu.input[10] = emu.display.Window.Pressed(pixelgl.KeyD)
	emu.input[11] = emu.display.Window.Pressed(pixelgl.KeyF)
	emu.input[12] = emu.display.Window.Pressed(pixelgl.KeyZ)
	emu.input[13] = emu.display.Window.Pressed(pixelgl.KeyX)
	emu.input[14] = emu.display.Window.Pressed(pixelgl.KeyC)
	emu.input[15] = emu.display.Window.Pressed(pixelgl.KeyV)

	// Commands
	if emu.display.Window.JustPressed(pixelgl.KeyEscape) {
		emu.display.Window.SetClosed(true)
	}
	if emu.display.Window.JustPressed(pixelgl.KeyF1) {
		emu.display.DisplayFps = !emu.display.DisplayFps
	}
	if emu.display.Window.JustPressed(pixelgl.KeyF2) {
		emu.display.ToggleVSync()
	}
	if emu.display.Window.JustPressed(pixelgl.KeyF5) {
		emu.reset()
	}
	if emu.display.Window.JustPressed(pixelgl.KeyF11) {
		emu.display.ToggleFullscreen()
	}
	if emu.display.Window.JustPressed(pixelgl.KeyP) {
		emu.setPause(!emu.pause)
	}
	if emu.display.Window.JustPressed(pixelgl.KeyM) {
		emu.mute = !emu.mute
	}
}
