package main

import (
	"io"
	"log"
	"time"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/smf"
)

type TrackOptions struct {
	MeterNumerator   int
	MeterDenominator int
	Tempo            float64
	Instrument       string
}

func (o TrackOptions) NewTrack() smf.Track {
	track := smf.Track{}
	track.Add(0, smf.MetaMeter(uint8(o.MeterNumerator), uint8(o.MeterDenominator)))
	track.Add(0, smf.MetaTempo(o.Tempo))
	track.Add(0, smf.MetaInstrument(o.Instrument))
	return track
}

type RecorderOptions struct {
	Cancel          chan struct{}
	PortManager     *PortManager
	TrackOptions    TrackOptions
	GetRecordOutput func() io.WriteCloser
	Timeout         int
}

type midiUnit struct {
	msg     midi.Message
	deltams int32
}

type Recorder struct {
	opts     RecorderOptions
	ticks    smf.MetricTicks
	track    smf.Track
	unitChan chan midiUnit
	lastms   int32
}

func StartRecording(options RecorderOptions) *Recorder {
	r := &Recorder{
		ticks:    smf.MetricTicks(96),
		opts:     options,
		track:    nil,
		unitChan: make(chan midiUnit, 1024),
	}

	go r.worker()

	r.opts.PortManager.AddListener(r)
	if r.opts.Cancel != nil {
		go func() {
			<-r.opts.Cancel
			log.Println("stopping...")
			r.opts.PortManager.Stop()
		}()
	}

	return r
}

func (r *Recorder) newRecording() {
	if r.track != nil {
		mf := smf.New()
		mf.TimeFormat = r.ticks
		mf.Add(r.track)

		output := r.opts.GetRecordOutput()
		_, err := mf.WriteTo(output)
		if err != nil {
			log.Println("got error while attempting to save recording:", err.Error())
		}

		err = output.Close()
		if err != nil {
			log.Println("got error while attempting to close output file:", err.Error())
		}
	}
	r.track = r.opts.TrackOptions.NewTrack()
}

func (r *Recorder) worker() {
	timer := time.NewTimer(9999 * 24 * time.Hour)
	for {
		select {
		case <-timer.C:
			if r.track == nil {
				break
			}
			r.newRecording()
		case unit := <-r.unitChan:
			delta := r.ticks.Ticks(
				r.opts.TrackOptions.Tempo,
				time.Duration(unit.deltams)*time.Millisecond,
			)
			r.track.Add(delta, unit.msg)
			timer.Stop()
			timer.Reset(time.Second * time.Duration(r.opts.Timeout))
		}
	}
}

func (r *Recorder) OnMessage(msg midi.Message, currentms int32) {
	deltams := currentms - r.lastms
	r.lastms = currentms
	r.unitChan <- midiUnit{
		msg:     msg,
		deltams: deltams,
	}
}

