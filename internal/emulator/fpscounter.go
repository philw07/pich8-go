package emulator

import "time"

// FpsCounter calculates the frames per second
type FpsCounter struct {
	start       time.Time
	frames      int
	fps         float64
	previousFps float64
}

// NewFpsCounter creates and initializes a new instance
func NewFpsCounter() *FpsCounter {
	return &FpsCounter{
		start: time.Now(),
	}
}

// Tick needs to be called every time a frame is drawn and returns the current fps value
func (fps *FpsCounter) Tick() float64 {
	fps.frames++

	// Update fps value every second
	if time.Since(fps.start).Seconds() >= 1 {
		newFps := float64(fps.frames) / float64(time.Since(fps.start).Nanoseconds()) * 1_000_000_000
		fps.previousFps = fps.fps
		if fps.previousFps > 0 {
			fps.fps = 0.33*fps.previousFps + 0.33*fps.fps + 0.34*newFps
		} else {
			fps.fps = newFps
		}

		fps.start = time.Now()
		fps.frames = 0
	}

	return fps.fps
}
