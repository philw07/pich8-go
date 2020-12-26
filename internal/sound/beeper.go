package sound

import (
	"time"

	"github.com/faiface/beep"
)

const (
	beepFrequency = 440
	beepVolume    = 0.05
)

type beeper struct {
	beepInterval int
	beepPos      int

	samplesPerPlay int
	samplesToPlay  int
}

func newBeeper(sampleRate beep.SampleRate) *beeper {
	beepDuration := time.Second / 30
	return &beeper{
		beepInterval: int(sampleRate) / beepFrequency,

		samplesPerPlay: sampleRate.N(beepDuration),
	}
}

func (b *beeper) Play() {
	b.samplesToPlay += b.samplesPerPlay
}

func (b *beeper) Stream(samples [][2]float64) (int, bool) {
	for i := range samples {
		sample := 0.0
		if b.samplesToPlay > 0 && b.beepPos >= b.beepInterval {
			sample = 0.05
		}
		samples[i][0] = sample
		samples[i][1] = sample

		if b.samplesToPlay > 0 {
			b.samplesToPlay--
		}
		b.beepPos = (b.beepPos + 1) % (b.beepInterval * 2)
	}

	return len(samples), true
}

func (b *beeper) Err() error {
	return nil
}
