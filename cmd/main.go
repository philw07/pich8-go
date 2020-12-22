package main

import (
	"io/ioutil"

	"github.com/faiface/pixel/pixelgl"
	"github.com/philw07/pich8-go/internal/emulator"
)

func main() {
	pixelgl.Run(func() {
		data, err := ioutil.ReadFile("roms/c8games/BRIX")
		if err != nil {
			panic(err)
		}

		emu, err := emulator.NewEmulator()
		if err != nil {
			panic(err)
		}

		emu.LoadRom(data)
		emu.Run()
	})
}
