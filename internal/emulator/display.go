package emulator

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/philw07/pich8-go/internal/videomemory"
	"golang.org/x/image/font/basicfont"
)

const (
	windowTitle = "pich8-go"

	c8Width  = 64
	c8Height = 32
)

// Display represents a display for the CHIP-8 emulator
type Display struct {
	Window              *pixelgl.Window
	fpsCounter          FpsCounter
	fpsText             *text.Text
	lastCPUSpeedTime    time.Time
	cpuSpeedText        *text.Text
	DisplayFps          bool
	DisplayInstructions bool
	instructionsText    *text.Text
	imd                 *imdraw.IMDraw
}

// NewDisplay creates and initializes a new Display instance
func NewDisplay() (*Display, error) {
	montitorWidth, monitorHeight := pixelgl.PrimaryMonitor().Size()

	cfg := pixelgl.WindowConfig{
		Title: windowTitle,
		// Icon: TODO:
		Bounds:    pixel.R(0, 0, 10*c8Width, 10*c8Height),
		Position:  pixel.V(montitorWidth/2-5*c8Width, monitorHeight/2-5*c8Height),
		VSync:     false,
		Resizable: true,
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		return nil, err
	}

	textAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	instuctionsText := text.New(pixel.ZV, textAtlas)
	fmt.Fprintln(instuctionsText, "Ctrl + O    Open ROM")
	fmt.Fprintln(instuctionsText, "Page Up     Increase CPU Speed")
	fmt.Fprintln(instuctionsText, "Page Down   Decrease CPU Speed")
	fmt.Fprintln(instuctionsText, "P           Pause on/off")
	fmt.Fprintln(instuctionsText, "M           Mute on/off")
	fmt.Fprintln(instuctionsText, "F1          Display these instructions")
	fmt.Fprintln(instuctionsText, "F2          Display FPS")
	fmt.Fprintln(instuctionsText, "F3          VSync on/off")
	fmt.Fprintln(instuctionsText, "F5          Reset")
	fmt.Fprintln(instuctionsText, "F11         Fullscreen")

	return &Display{
		Window:              win,
		fpsText:             text.New(pixel.ZV, textAtlas),
		cpuSpeedText:        text.New(pixel.ZV, textAtlas),
		DisplayInstructions: true,
		instructionsText:    instuctionsText,
		imd:                 imdraw.New(nil),
	}, nil
}

// ToggleFullscreen toggles between fullscreen and windowed
func (disp *Display) ToggleFullscreen() {
	if disp.Window.Monitor() == nil {
		disp.Window.SetMonitor(pixelgl.PrimaryMonitor())
	} else {
		disp.Window.SetMonitor(nil)
	}
}

// ToggleVSync toggles between vsync on and off
func (disp *Display) ToggleVSync() {
	disp.Window.SetVSync(!disp.Window.VSync())
}

// DisplayCPUSpeed displays the given speed for a short time
func (disp *Display) DisplayCPUSpeed(speed int) {
	disp.lastCPUSpeedTime = time.Now()
	disp.cpuSpeedText.Clear()
	fmt.Fprintf(disp.cpuSpeedText, "%vHz", speed)
}

// Draw draws the content of the given VideoMemory to the window
func (disp *Display) Draw(vmem videomemory.VideoMemory) {
	w := disp.Window.Bounds().W()
	h := disp.Window.Bounds().H()

	disp.Window.Clear(color.Black)

	// Draw
	image := disp.copyFrameToImage(vmem)
	pic := pixel.PictureDataFromImage(image)
	sprite := pixel.NewSprite(pic, pic.Bounds())
	mat := pixel.IM
	mat = mat.Moved(disp.Window.Bounds().Center())
	mat = mat.ScaledXY(disp.Window.Bounds().Center(), pixel.V(w/float64(vmem.RenderWidth()), h/float64(vmem.RenderHeight())))
	sprite.Draw(disp.Window, mat)

	// Update and draw fps
	fps := disp.fpsCounter.Tick()
	if disp.DisplayFps {
		disp.fpsText.Clear()
		fmt.Fprintf(disp.fpsText, "%v", int(fps))
		disp.drawText(disp.fpsText, pixel.ZV)
	}

	// Display CPU speed
	if time.Since(disp.lastCPUSpeedTime).Seconds() <= 2 {
		xPos := disp.Window.Bounds().W() - disp.cpuSpeedText.Bounds().W()
		disp.drawText(disp.cpuSpeedText, pixel.V(xPos, 0))
	}

	// Display instructions
	if disp.DisplayInstructions {
		x := math.Floor(w/2 - disp.instructionsText.Bounds().W()/2)
		y := math.Floor(-disp.instructionsText.Dot.Y)
		disp.drawText(disp.instructionsText, pixel.V(x, y))
	}

	disp.Window.Update()
}

func (disp *Display) copyFrameToImage(vmem videomemory.VideoMemory) image.Image {
	image := image.NewRGBA(image.Rect(0, 0, vmem.RenderWidth(), vmem.RenderHeight()))
	for x := 0; x < vmem.RenderWidth(); x++ {
		for y := 0; y < vmem.RenderHeight(); y++ {
			if vmem.GetIndex(videomemory.FirstPlane, vmem.ToIndex(x, y)) && vmem.GetIndex(videomemory.SecondPlane, vmem.ToIndex(x, y)) {
				image.Set(x, y, pixel.RGB(0.33, 0.33, 0.33))
			} else if vmem.GetIndex(videomemory.FirstPlane, vmem.ToIndex(x, y)) {
				image.Set(x, y, pixel.RGB(1.0, 1.0, 1.0))
			} else if vmem.GetIndex(videomemory.SecondPlane, vmem.ToIndex(x, y)) {
				image.Set(x, y, pixel.RGB(0.66, 0.66, 0.66))
			} else {
				image.Set(x, y, color.Black)
			}
		}
	}
	return image
}

func (disp *Display) drawText(text *text.Text, pos pixel.Vec) {
	const margin = 5

	disp.imd.Clear()

	disp.imd.Color = pixel.RGB(0.1, 0.1, 0.1).Mul(pixel.Alpha(0.85))
	disp.imd.Push(text.Bounds().Min.Add(pos).Add(pixel.V(-margin, -margin)))
	disp.imd.Push(text.Bounds().Max.Add(pos).Add(pixel.V(margin, margin)))
	disp.imd.Rectangle(0)
	disp.imd.Draw(disp.Window)

	text.Draw(disp.Window, pixel.IM.Moved(pos))
}
