package sound

import (
	"time"

	"github.com/faiface/beep"
)

const (
	bufferFrequency = 4000
)

type xoChipBuffer struct {
	samples []float64
	index   int
}

// newXOChipBuffer creates a new streamer for the given XO-CHIP audio buffer
func newXOChipBuffer(buffer [16]byte, sampleRate beep.SampleRate) *xoChipBuffer {
	reps := sampleRate / bufferFrequency
	samplesLen := sampleRate.N(time.Second / 60)
	samples := make([]float64, samplesLen)
	i := 0
outer:
	for _, byt := range buffer {
		for bitIdx := 0; bitIdx < 8; bitIdx++ {
			bit := byt>>(7-bitIdx)&1 == 1

			for x := 0; x < int(reps); x++ {
				if bit {
					samples[i] = volume
				} else {
					samples[i] = 0
				}

				i++
				if i >= samplesLen {
					break outer
				}
			}
		}
	}

	return &xoChipBuffer{
		samples: samples,
		index:   0,
	}
}

func (sb *xoChipBuffer) Stream(samples [][2]float64) (n int, ok bool) {
	if sb.index >= len(sb.samples) {
		return 0, false
	}

	filled := 0
	for i := range samples {
		val := 0.0
		if sb.index < len(sb.samples) {
			val = sb.samples[sb.index]
			sb.index++
			filled++
		}

		samples[i][0] = val
		samples[i][1] = val
	}

	return filled, true
}

func (sb *xoChipBuffer) Err() error {
	return nil
}
