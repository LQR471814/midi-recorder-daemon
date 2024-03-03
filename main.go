package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"time"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	"gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

func main() {
	showPorts := flag.Bool("show-ports", false, "show midi ports available without recording, if you specify this flag, you will not need to specify any other flags.")

	output := flag.String("output", "output", "the path to the folder where all midi recordings will be saved.")
	portNumber := flag.Int("port", -1, "the midi port number to record on.")

	meterNumerator := flag.Int("meter-numerator", 4, "the numerator of the time signature.")
	meterDenominator := flag.Int("meter-denominator", 4, "the denominator of the time signature.")
	tempo := flag.Float64("tempo", 120, "the tempo of the midi file.")
	instrument := flag.String("instrument", "Piano", "the instrument to use in the midi file.")

	timeout := flag.Int("timeout", 10, "how many seconds to wait before saving the current midi recording.")

	flag.Parse()

	drv, err := rtmididrv.New()
	if err != nil {
		log.Fatal(err)
	}
	drivers.Register(drv)

	if *showPorts {
		ports := midi.GetInPorts()
		for _, p := range ports {
			fmt.Printf("%d | %s\n", p.Number(), p.String())
		}
		return
	}

	if *portNumber < 0 {
		log.Fatal("you must specify a midi port to record on with the '-port' flag.")
	}

	port, err := midi.InPort(*portNumber)
	if err != nil {
		log.Fatal(err)
	}

	err = os.Mkdir(*output, 0777)
	if err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}

	recorder := NewRecorder(RecorderOptions{
		Timeout: *timeout,
		Port:    port,
		TrackOptions: TrackOptions{
			MeterNumerator:   *meterNumerator,
			MeterDenominator: *meterDenominator,
			Tempo:            *tempo,
			Instrument:       *instrument,
		},
		GetRecordOutput: func() io.WriteCloser {
			now := time.Now()
			year := strconv.Itoa(now.Year())
			month := now.Month().String()
			day := strconv.Itoa(now.Day())

			dayPath := filepath.Join(*output, year, month, day)
			err := os.MkdirAll(dayPath, 0777)
			if err != nil && !os.IsExist(err) {
				log.Fatal("failed to create output directories:", err.Error())
			}

			fpath := filepath.Join(
				dayPath,
				fmt.Sprintf(
					"%dh_%dm_%ds.midi",
					now.Hour(),
					now.Minute(),
					now.Second(),
				),
			)
			out, err := os.Create(fpath)
			if err != nil && !os.IsExist(err) {
				log.Fatal("failed to create file for writing:", err.Error())
			}

			log.Printf("new recording at \"%s\"\n", fpath)

			return out
		},
	})

	signalAccepter := make(chan os.Signal)
	signal.Notify(signalAccepter, os.Interrupt, os.Kill)

	cancelSignal := make(chan struct{})
	go func() {
		<-signalAccepter
		cancelSignal <- struct{}{}
	}()

	err = recorder.AutoRecord(cancelSignal)
	if err != nil {
		log.Fatal(err)
	}
}
