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
	"strings"
	"time"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	"gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

func main() {
	showPorts := flag.Bool("show-ports", false, "show midi ports available without recording, if you specify this flag, you will not need to specify any other flags.")

	output := flag.String("output", "output", "the path to the folder where all midi recordings will be saved.")
	portNumber := flag.Int("port-number", -1, "search for a midi port by its port number, this flag is mutually exclusive with '-port-name'.")
	portName := flag.String("port-name", "", "search for a midi port by a keyword in its lowercased name, this flag is mutually exclusive with '-port-number'.")
	portPollTimeout := flag.Int("port-poll-timeout", 5, "seconds to wait between polling if a midi port exists.")

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

	lowerPortName := strings.ToLower(*portName)

	if *portNumber < 0 && lowerPortName == "" {
		log.Fatal("you must specify either '-port-name' or '-port-number' to record midi.")
	}
	if *portNumber >= 0 && lowerPortName != "" {
		log.Fatal("you cannot specify both '-port-name' and '-port-number'.")
	}
	if lowerPortName != "" {
		log.Printf("searching for port via keyword \"%s\"\n", lowerPortName)
	}
	if *portNumber >= 0 {
		log.Printf("searching for port number %d\n", *portNumber)
	}

	portManager := NewPortManager(PortManagerOptions{
		PortNumber:  *portNumber,
		PortName:    *portName,
		PollTimeout: *portPollTimeout,
	})

	err = os.Mkdir(*output, 0777)
	if err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}

	signalAccepter := make(chan os.Signal)
	signal.Notify(signalAccepter, os.Interrupt, os.Kill)

	cancel := make(chan struct{})

	StartRecording(RecorderOptions{
		Cancel:      cancel,
		Timeout:     *timeout,
		PortManager: portManager,
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

	<-signalAccepter
	cancel <- struct{}{}
}
