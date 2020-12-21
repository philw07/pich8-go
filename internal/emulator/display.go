package emulator

import (
	"image"
	"image/color"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

const (
	windowTitle = "pich8-go"

	c8Width  = 64
	c8Height = 32
)

type Display struct {
	Window pixelgl.Window
	image  image.RGBA
}

func NewDisplay() (*Display, error) {
	montitorWidth, monitorHeight := pixelgl.PrimaryMonitor().Size()

	cfg := pixelgl.WindowConfig{
		Title: windowTitle,
		// Icon: TODO:
		Bounds:   pixel.R(0, 0, 10*c8Width, 10*c8Height),
		Position: pixel.V(montitorWidth-5*c8Width, monitorHeight-5*c8Height),
		VSync:    true,
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		return nil, err
	}

	return &Display{
		Window: *win,
	}, nil
}

func (disp *Display) Draw(vmem VideoMemory) {
	disp.copyFrame(vmem)

	pic := pixel.PictureDataFromImage(&disp.image)
	sprite := pixel.NewSprite(pic, pic.Bounds())

	mat := pixel.IM
	mat = mat.Moved(disp.Window.Bounds().Center())
	mat = mat.ScaledXY(disp.Window.Bounds().Center(), pixel.V(disp.Window.Bounds().W()/64, disp.Window.Bounds().H()/32))

	disp.Window.Clear(color.Black)
	sprite.Draw(&disp.Window, mat)
	disp.Window.Update()
}

func (disp *Display) copyFrame(vmem VideoMemory) {
	for x := 0; x < vmem.RenderWidth(); x++ {
		for y := 0; y < vmem.RenderHeight(); y++ {
			if vmem.Get(FirstPlane, x, y) && vmem.Get(SecondPlane, x, y) {
				disp.image.Set(x, y, color.RGBA{R: 85, G: 85, B: 85, A: 255})
			} else if vmem.Get(FirstPlane, x, y) {
				disp.image.Set(x, y, color.White)
			} else if vmem.Get(SecondPlane, x, y) {
				disp.image.Set(x, y, color.RGBA{R: 170, G: 170, B: 170, A: 255})
			} else {
				disp.image.Set(x, y, color.Black)
			}
		}
	}
}
