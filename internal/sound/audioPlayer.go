package sound

import (
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
)

const (
	sampleRate = 48000
	volume     = 0.05
)

type AudioPlayer struct {
	sampleRate beep.SampleRate
	queue      *queue
	beeper     *beeper
}

func NewAudioPlayer() *AudioPlayer {
	sr := beep.SampleRate(sampleRate)
	speaker.Init(sr, sr.N(time.Second/15))

	queue := queue{}
	beeper := newBeeper(sr)
	speaker.Play(beep.Mix(&queue, beeper))

	return &AudioPlayer{
		sampleRate: sr,
		queue:      &queue,
		beeper:     beeper,
	}
}

func (ap *AudioPlayer) Beep() {
	ap.beeper.Play()
}

func (ap *AudioPlayer) PlayBuffer(buffer [16]byte) {
	ap.queue.Add(newXOChipBuffer(buffer, ap.sampleRate))
}
