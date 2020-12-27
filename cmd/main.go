package main

import (
	"github.com/faiface/pixel/pixelgl"
	"github.com/philw07/pich8-go/internal/emulator"
)

func main() {

	pixelgl.Run(func() {
		emu, err := emulator.NewEmulator()
		if err != nil {
			panic(err)
		}
		emu.Run()
	})
}
