package emulator

type VideoMode byte
type Plane byte

const (
	DefaultVideoMode  VideoMode = 1
	HiResVideoMode    VideoMode = 2
	ExtendedVideoMode VideoMode = 3

	NoPlane     Plane = 0
	FirstPlane  Plane = 1
	SecondPlane Plane = 2
	BothPlanes  Plane = 3

	widthDefault   = 64
	heightDefault  = 32
	widthExtended  = 128
	heightExtended = 64
	widthHiRes     = 64
	heightHiRes    = 64
)

type VideoMemory struct {
	vmemPlane1 [128 * 64]bool
	vmemPlane2 [128 * 64]bool
	VideoMode  VideoMode
	Plane      Plane
}

// NewVideoMemory creates and initializes a new instance
func NewVideoMemory() *VideoMemory {
	return &VideoMemory{
		VideoMode: DefaultVideoMode,
		Plane:     FirstPlane,
	}
}

func (vmem *VideoMemory) Set(plane Plane, x, y int, value bool) {

	// In default video mode, we're translating the 64x32 screen to 128x64,
	// only this way the scroll commands work correctly in S-CHIP low res mode.
	if vmem.VideoMode == DefaultVideoMode {
		x *= 2
		y *= 2
		vmem.setIndex(plane, vmem.ToIndex(x, y), value)
		vmem.setIndex(plane, vmem.ToIndex(x+1, y), value)
		vmem.setIndex(plane, vmem.ToIndex(x, y+1), value)
		vmem.setIndex(plane, vmem.ToIndex(x+1, y+1), value)
	} else {
		vmem.setIndex(plane, vmem.ToIndex(x, y), value)
	}
}

func (vmem *VideoMemory) setIndex(plane Plane, index int, value bool) {
	switch plane {
	case FirstPlane:
		vmem.vmemPlane1[index] = value
	case SecondPlane:
		vmem.vmemPlane2[index] = value
	case BothPlanes:
		vmem.vmemPlane1[index] = value
		vmem.vmemPlane2[index] = value
	}
}

func (vmem *VideoMemory) Clear() {
	vmem.SetAll(false)
}

func (vmem *VideoMemory) SetAll(value bool) {
	for i := 0; i < len(vmem.vmemPlane1); i++ {
		if vmem.Plane == FirstPlane || vmem.Plane == BothPlanes {
			vmem.vmemPlane1[i] = value
		}
		if vmem.Plane == SecondPlane || vmem.Plane == BothPlanes {
			vmem.vmemPlane2[i] = value
		}
	}
}

func (vmem *VideoMemory) ToIndex(x, y int) int {
	return y*vmem.RenderWidth() + x
}

func (vmem *VideoMemory) Width() int {
	switch vmem.VideoMode {
	case ExtendedVideoMode:
		return widthExtended
	case HiResVideoMode:
		return widthHiRes
	default:
		return widthDefault
	}
}

func (vmem *VideoMemory) Height() int {
	switch vmem.VideoMode {
	case ExtendedVideoMode:
		return heightExtended
	case HiResVideoMode:
		return heightHiRes
	default:
		return heightDefault
	}
}

func (vmem *VideoMemory) RenderWidth() int {
	switch vmem.VideoMode {
	case HiResVideoMode:
		return widthHiRes
	default:
		return widthExtended
	}
}

func (vmem *VideoMemory) RenderHeight() int {
	switch vmem.VideoMode {
	case HiResVideoMode:
		return heightHiRes
	default:
		return heightExtended
	}
}

func (vmem *VideoMemory) GetIndex(plane Plane, index int) bool {
	switch plane {
	case FirstPlane:
		return vmem.vmemPlane1[index]
	case SecondPlane:
		return vmem.vmemPlane2[index]
	case BothPlanes:
		panic("shouldn't call get with both planes selected")
	default:
		return false
	}
}

func (vmem *VideoMemory) Get(plane Plane, x, y int) bool {
	curX := x
	curY := y
	if vmem.VideoMode == DefaultVideoMode {
		curX *= 2
		curY *= 2
	}
	return vmem.GetIndex(plane, vmem.ToIndex(curX, curY))
}

func (vmem *VideoMemory) ScrollDown(lines int) {
	num := lines
	if vmem.VideoMode == DefaultVideoMode {
		num *= 2
	}

	for y := vmem.RenderHeight() - 1; y >= 0; y-- {
		for x := 0; x < vmem.RenderWidth(); x++ {
			for _, plane := range [...]Plane{FirstPlane, SecondPlane} {
				if vmem.Plane == plane || vmem.Plane == BothPlanes {
					// Need to use get_index and set_index instead of get/set, because set expects 64x32 coordinates in default video mode
					val := false
					if y >= num {
						val = vmem.GetIndex(plane, vmem.ToIndex(x, y-num))
					}
					vmem.setIndex(plane, vmem.ToIndex(x, y), val)
				}
			}
		}
	}
}

func (vmem *VideoMemory) ScrollUp(lines int) {
	num := lines
	if vmem.VideoMode == DefaultVideoMode {
		num *= 2
	}

	for y := 0; y < vmem.RenderHeight(); y++ {
		for x := 0; x < vmem.RenderWidth(); x++ {
			for _, plane := range [...]Plane{FirstPlane, SecondPlane} {
				if vmem.Plane == plane || vmem.Plane == BothPlanes {
					// Need to use get_index and set_index instead of get/set, because set expects 64x32 coordinates in default video mode
					val := false
					if y < vmem.RenderHeight()-num {
						val = vmem.GetIndex(plane, vmem.ToIndex(x, y+num))
					}
					vmem.setIndex(plane, vmem.ToIndex(x, y), val)
				}
			}
		}
	}
}

func (vmem *VideoMemory) ScrollLeft() {
	num := 4
	if vmem.VideoMode == DefaultVideoMode {
		num *= 2
	}

	for x := 0; x < vmem.RenderWidth(); x++ {
		for y := 0; y < vmem.RenderHeight(); y++ {
			for _, plane := range [...]Plane{FirstPlane, SecondPlane} {
				if vmem.Plane == plane || vmem.Plane == BothPlanes {
					// Need to use get_index and set_index instead of get/set, because set expects 64x32 coordinates in default video mode
					val := false
					if x < vmem.RenderWidth()-num {
						val = vmem.GetIndex(plane, vmem.ToIndex(x+num, y))
					}
					vmem.setIndex(plane, vmem.ToIndex(x, y), val)
				}
			}
		}
	}
}

func (vmem *VideoMemory) ScrollRight() {
	num := 4
	if vmem.VideoMode == DefaultVideoMode {
		num *= 2
	}

	for x := vmem.RenderWidth() - 1; x >= 0; x-- {
		for y := 0; y < vmem.RenderHeight(); y++ {
			for _, plane := range [...]Plane{FirstPlane, SecondPlane} {
				if vmem.Plane == plane || vmem.Plane == BothPlanes {
					// Need to use get_index and set_index instead of get/set, because set expects 64x32 coordinates in default video mode
					val := false
					if x >= num {
						val = vmem.GetIndex(plane, vmem.ToIndex(x-num, y))
					}
					vmem.setIndex(plane, vmem.ToIndex(x, y), val)
				}
			}
		}
	}
}
