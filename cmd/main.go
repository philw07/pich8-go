package main

import (
	"fmt"

	"github.com/philw07/pich8-go/internal/cpu"
)

func main() {
	cpu := cpu.NewCPU()
	fmt.Printf("%d", cpu.PC)
}
