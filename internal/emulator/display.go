package emulator

import (
	"image"
	"image/color"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/philw07/pich8-go/internal/videomemory"
)

const (
	windowTitle = "pich8-go"

	c8Width  = 64
	c8Height = 32
)

// Display represents a display for the CHIP-8 emulator
type Display struct {
	Window *pixelgl.Window
}

// NewDisplay creates and initializes a new Display instance
func NewDisplay() (*Display, error) {
	montitorWidth, monitorHeight := pixelgl.PrimaryMonitor().Size()

	cfg := pixelgl.WindowConfig{
		Title: windowTitle,
		// Icon: TODO:
		Bounds:    pixel.R(0, 0, 10*c8Width, 10*c8Height),
		Position:  pixel.V(montitorWidth/2-5*c8Width, monitorHeight/2-5*c8Height),
		VSync:     true,
		Resizable: true,
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		return nil, err
	}

	return &Display{
		Window: win,
	}, nil
}

// Draw draws the content of the given VideoMemory to the window
func (disp *Display) Draw(vmem videomemory.VideoMemory) {
	image := disp.copyFrameToImage(vmem)

	pic := pixel.PictureDataFromImage(image)
	sprite := pixel.NewSprite(pic, pic.Bounds())

	mat := pixel.IM
	mat = mat.Moved(disp.Window.Bounds().Center())
	mat = mat.ScaledXY(disp.Window.Bounds().Center(), pixel.V(disp.Window.Bounds().W()/float64(vmem.RenderWidth()), disp.Window.Bounds().H()/float64(vmem.RenderHeight())))

	disp.Window.Clear(color.Black)
	sprite.Draw(disp.Window, mat)
	disp.Window.Update()
}

func (disp *Display) copyFrameToImage(vmem videomemory.VideoMemory) image.Image {
	image := image.NewRGBA(image.Rect(0, 0, vmem.RenderWidth(), vmem.RenderHeight()))
	for x := 0; x < vmem.RenderWidth(); x++ {
		for y := 0; y < vmem.RenderHeight(); y++ {
			if vmem.GetIndex(videomemory.FirstPlane, vmem.ToIndex(x, y)) && vmem.GetIndex(videomemory.SecondPlane, vmem.ToIndex(x, y)) {
				image.Set(x, y, color.RGBA{R: 85, G: 85, B: 85, A: 255})
			} else if vmem.GetIndex(videomemory.FirstPlane, vmem.ToIndex(x, y)) {
				image.Set(x, y, color.White)
			} else if vmem.GetIndex(videomemory.SecondPlane, vmem.ToIndex(x, y)) {
				image.Set(x, y, color.RGBA{R: 170, G: 170, B: 170, A: 255})
			} else {
				image.Set(x, y, color.Black)
			}
		}
	}
	return image
}
