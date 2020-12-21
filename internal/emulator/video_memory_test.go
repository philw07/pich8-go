package emulator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitialization(t *testing.T) {
	assert := assert.New(t)

	vmem := NewVideoMemory()
	assert.EqualValues(DefaultVideoMode, vmem.VideoMode)
	assert.EqualValues(FirstPlane, vmem.Plane)
	for i := range vmem.vmemPlane1 {
		assert.False(vmem.vmemPlane1[i])
		assert.False(vmem.vmemPlane2[i])
	}
}

func TestWidthHeight(t *testing.T) {
	assert := assert.New(t)

	vmem := NewVideoMemory()
	assert.EqualValues(DefaultVideoMode, vmem.VideoMode)
	assert.EqualValues(widthDefault, vmem.Width())
	assert.EqualValues(heightDefault, vmem.Height())
	assert.EqualValues(widthExtended, vmem.RenderWidth())
	assert.EqualValues(heightExtended, vmem.RenderHeight())
	vmem.VideoMode = HiResVideoMode
	assert.EqualValues(HiResVideoMode, vmem.VideoMode)
	assert.EqualValues(widthHiRes, vmem.Width())
	assert.EqualValues(heightHiRes, vmem.Height())
	assert.EqualValues(widthHiRes, vmem.RenderWidth())
	assert.EqualValues(heightHiRes, vmem.RenderHeight())
	vmem.VideoMode = ExtendedVideoMode
	assert.EqualValues(ExtendedVideoMode, vmem.VideoMode)
	assert.EqualValues(widthExtended, vmem.Width())
	assert.EqualValues(heightExtended, vmem.Height())
	assert.EqualValues(widthExtended, vmem.RenderWidth())
	assert.EqualValues(heightExtended, vmem.RenderHeight())
}

func TestOperations(t *testing.T) {
	assert := assert.New(t)

	// Set - Default
	vmem := NewVideoMemory()
	vmem.VideoMode = DefaultVideoMode
	vmem.Set(vmem.Plane, 32, 20, true)
	assert.True(vmem.Get(vmem.Plane, 32, 20))
	// Set - HiRes
	vmem = NewVideoMemory()
	vmem.VideoMode = HiResVideoMode
	vmem.Set(vmem.Plane, 32, 50, true)
	assert.True(vmem.Get(vmem.Plane, 32, 50))
	// Set - Extended
	vmem = NewVideoMemory()
	vmem.VideoMode = ExtendedVideoMode
	vmem.Set(vmem.Plane, 100, 50, true)
	assert.True(vmem.Get(vmem.Plane, 100, 50))

	// Set all - Default
	vmem = NewVideoMemory()
	vmem.VideoMode = DefaultVideoMode
	vmem.SetAll(true)
	for i := 0; i < widthDefault*heightDefault; i++ {
		assert.True(vmem.GetIndex(vmem.Plane, i))
	}
	// Set all - HiRes
	vmem = NewVideoMemory()
	vmem.VideoMode = HiResVideoMode
	vmem.SetAll(true)
	for i := 0; i < widthHiRes*heightHiRes; i++ {
		assert.True(vmem.GetIndex(vmem.Plane, i))
	}
	// Set all - Extended
	vmem = NewVideoMemory()
	vmem.VideoMode = ExtendedVideoMode
	vmem.SetAll(true)
	for i := 0; i < widthExtended*heightExtended; i++ {
		assert.True(vmem.GetIndex(vmem.Plane, i))
	}

	// Clear - Default
	vmem = NewVideoMemory()
	vmem.VideoMode = DefaultVideoMode
	vmem.SetAll(true)
	vmem.Clear()
	for i := 0; i < widthDefault*heightDefault; i++ {
		assert.False(vmem.GetIndex(vmem.Plane, i))
	}
	// Clear - HiRes
	vmem = NewVideoMemory()
	vmem.VideoMode = HiResVideoMode
	vmem.SetAll(true)
	vmem.Clear()
	for i := 0; i < widthHiRes*heightHiRes; i++ {
		assert.False(vmem.GetIndex(vmem.Plane, i))
	}
	// Clear - Extended
	vmem = NewVideoMemory()
	vmem.VideoMode = ExtendedVideoMode
	vmem.SetAll(true)
	vmem.Clear()
	for i := 0; i < widthExtended*heightExtended; i++ {
		assert.False(vmem.GetIndex(vmem.Plane, i))
	}

	// Scroll down
	vmem = NewVideoMemory()
	vmem.VideoMode = ExtendedVideoMode
	vmem.SetAll(true)
	for x := 0; x < widthExtended; x++ {
		vmem.Set(vmem.Plane, x, 35, false)
	}
	vmem.ScrollDown(3)
	for x := 0; x < widthExtended; x++ {
		for y := 0; y < 4; y++ {
			assert.EqualValues(y == 3, vmem.Get(vmem.Plane, x, y))
		}
		assert.False(vmem.Get(vmem.Plane, x, 38))
	}
	// Scroll down - default video mode
	vmem = NewVideoMemory()
	vmem.VideoMode = DefaultVideoMode
	vmem.SetAll(true)
	for x := 0; x < widthDefault; x++ {
		vmem.Set(vmem.Plane, x, 3, false)
	}
	vmem.ScrollDown(3)
	for x := 0; x < widthDefault; x++ {
		for y := 0; y < 4; y++ {
			assert.EqualValues(y == 3, vmem.Get(vmem.Plane, x, y))
		}
		assert.False(vmem.Get(vmem.Plane, x, 6))
	}

	// Scroll up
	vmem = NewVideoMemory()
	vmem.VideoMode = ExtendedVideoMode
	vmem.SetAll(true)
	for x := 0; x < widthExtended; x++ {
		vmem.Set(vmem.Plane, x, 35, false)
	}
	vmem.ScrollUp(7)
	for x := 0; x < widthExtended; x++ {
		for y := 56; y < heightExtended; y++ {
			assert.EqualValues(y == 56, vmem.Get(vmem.Plane, x, y))
		}
		assert.False(vmem.Get(vmem.Plane, x, 28))
	}
	// Scroll up - default video mode
	vmem = NewVideoMemory()
	vmem.VideoMode = DefaultVideoMode
	vmem.SetAll(true)
	for x := 0; x < widthDefault; x++ {
		vmem.Set(vmem.Plane, x, 10, false)
	}
	vmem.ScrollUp(7)
	for x := 0; x < widthDefault; x++ {
		for y := 24; y < heightDefault; y++ {
			assert.EqualValues(y == 24, vmem.Get(vmem.Plane, x, y))
		}
		assert.False(vmem.Get(vmem.Plane, x, 3))
	}

	// Scroll left
	vmem = NewVideoMemory()
	vmem.VideoMode = ExtendedVideoMode
	vmem.SetAll(true)
	for y := 0; y < heightExtended; y++ {
		vmem.Set(vmem.Plane, 108, y, false)
	}
	vmem.ScrollLeft()
	for y := 0; y < heightExtended; y++ {
		for x := 123; x < widthExtended; x++ {
			assert.EqualValues(x == 123, vmem.Get(vmem.Plane, x, y))
		}
		assert.False(vmem.Get(vmem.Plane, 104, y))
	}
	// Scroll left - default video mode
	vmem = NewVideoMemory()
	vmem.VideoMode = DefaultVideoMode
	vmem.SetAll(true)
	for y := 0; y < heightDefault; y++ {
		vmem.Set(vmem.Plane, 44, y, false)
	}
	vmem.ScrollLeft()
	for y := 0; y < heightDefault; y++ {
		for x := 59; x < widthDefault; x++ {
			assert.EqualValues(x == 59, vmem.Get(vmem.Plane, x, y))
		}
		assert.False(vmem.Get(vmem.Plane, 40, y))
	}

	// Scroll right
	vmem = NewVideoMemory()
	vmem.VideoMode = ExtendedVideoMode
	vmem.SetAll(true)
	for y := 0; y < heightExtended; y++ {
		vmem.Set(vmem.Plane, 99, y, false)
	}
	vmem.ScrollRight()
	for y := 0; y < heightExtended; y++ {
		for x := 0; x < 5; x++ {
			assert.EqualValues(x == 4, vmem.Get(vmem.Plane, x, y))
		}
		assert.False(vmem.Get(vmem.Plane, 103, y))
	}
	// Scroll right - default video mode
	vmem = NewVideoMemory()
	vmem.VideoMode = DefaultVideoMode
	vmem.SetAll(true)
	for y := 0; y < heightDefault; y++ {
		vmem.Set(vmem.Plane, 35, y, false)
	}
	vmem.ScrollRight()
	for y := 0; y < heightDefault; y++ {
		for x := 0; x < 5; x++ {
			assert.EqualValues(x == 4, vmem.Get(vmem.Plane, x, y))
		}
		assert.False(vmem.Get(vmem.Plane, 39, y))
	}
}

func TestPlanes(t *testing.T) {
	assert := assert.New(t)

	for _, plane := range [...]Plane{NoPlane, FirstPlane, SecondPlane, BothPlanes} {
		first := plane == FirstPlane || plane == BothPlanes
		second := plane == SecondPlane || plane == BothPlanes

		vmem := NewVideoMemory()
		vmem.Plane = plane

		// Set all
		vmem.SetAll(true)
		for i := 0; i < 128*64; i++ {
			assert.EqualValues(first, vmem.GetIndex(FirstPlane, i))
			assert.EqualValues(second, vmem.GetIndex(SecondPlane, i))
		}

		// Clear
		vmem.Clear()
		for i := 0; i < 128*64; i++ {
			assert.False(vmem.GetIndex(FirstPlane, i))
			assert.False(vmem.GetIndex(SecondPlane, i))
		}

		// Set x y + Get x y
		vmem.Set(vmem.Plane, 10, 10, true)
		assert.EqualValues(first, vmem.Get(FirstPlane, 10, 10))
		assert.EqualValues(second, vmem.Get(SecondPlane, 10, 10))

		// Scroll down
		vmem = NewVideoMemory()
		vmem.VideoMode = ExtendedVideoMode
		vmem.Plane = plane
		vmem.SetAll(true)
		for x := 0; x < widthExtended; x++ {
			vmem.Set(vmem.Plane, x, 35, false)
		}
		vmem.ScrollDown(3)
		for x := 0; x < widthExtended; x++ {
			for y := 0; y < 4; y++ {
				assert.EqualValues(first && y == 3, vmem.Get(FirstPlane, x, y))
				assert.EqualValues(second && y == 3, vmem.Get(SecondPlane, x, y))
			}
			assert.False(vmem.Get(FirstPlane, x, 38))
			assert.False(vmem.Get(SecondPlane, x, 38))
		}

		// Scroll up
		vmem = NewVideoMemory()
		vmem.VideoMode = ExtendedVideoMode
		vmem.Plane = plane
		vmem.SetAll(true)
		for x := 0; x < widthExtended; x++ {
			vmem.Set(vmem.Plane, x, 35, false)
		}
		vmem.ScrollUp(7)
		for x := 0; x < widthExtended; x++ {
			for y := 56; y < heightExtended; y++ {
				assert.EqualValues(first && y == 56, vmem.Get(FirstPlane, x, y))
				assert.EqualValues(second && y == 56, vmem.Get(SecondPlane, x, y))
			}
			assert.False(vmem.Get(FirstPlane, x, 28))
			assert.False(vmem.Get(SecondPlane, x, 28))
		}

		// Scroll left
		vmem = NewVideoMemory()
		vmem.VideoMode = ExtendedVideoMode
		vmem.Plane = plane
		vmem.SetAll(true)
		for y := 0; y < heightExtended; y++ {
			vmem.Set(vmem.Plane, 108, y, false)
		}
		vmem.ScrollLeft()
		for y := 0; y < heightExtended; y++ {
			for x := 123; x < widthExtended; x++ {
				assert.EqualValues(first && x == 123, vmem.Get(FirstPlane, x, y))
				assert.EqualValues(second && x == 123, vmem.Get(SecondPlane, x, y))
			}
			assert.False(vmem.Get(FirstPlane, 104, y))
			assert.False(vmem.Get(SecondPlane, 104, y))
		}

		// Scroll right
		vmem = NewVideoMemory()
		vmem.VideoMode = ExtendedVideoMode
		vmem.Plane = plane
		vmem.SetAll(true)
		for y := 0; y < heightExtended; y++ {
			vmem.Set(vmem.Plane, 99, y, false)
		}
		vmem.ScrollRight()
		for y := 0; y < heightExtended; y++ {
			for x := 0; x < 5; x++ {
				assert.EqualValues(first && x == 4, vmem.Get(FirstPlane, x, y))
				assert.EqualValues(second && x == 4, vmem.Get(SecondPlane, x, y))
			}
			assert.False(vmem.Get(FirstPlane, 103, y))
			assert.False(vmem.Get(SecondPlane, 103, y))
		}
	}
}
