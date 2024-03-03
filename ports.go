package main

import (
	"log"
	"strings"
	"time"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
)

type PortManagerOptions struct {
	PortNumber  int
	PortName    string
	PollTimeout int
}

type PortListener interface {
	OnMessage(msg midi.Message, currentms int32)
}

type PortManager struct {
	opts      PortManagerOptions
	port      drivers.In
	stop      func()
	listeners []PortListener
}

func NewPortManager(opts PortManagerOptions) *PortManager {
	m := &PortManager{
		opts: opts,
	}
	go m.scanner()
	return m
}

func (m *PortManager) listen(msg midi.Message, currentms int32) {
	for _, l := range m.listeners {
		l.OnMessage(msg, currentms)
	}
}

// specifying a nil port will stop the listen on the current port
func (m *PortManager) switchPort(port drivers.In) {
	if m.stop != nil {
		m.stop()
		m.port.Close()
	}
	if port == nil {
		if m.port != nil {
			log.Println("no active ports for now.")
		}
		m.port = nil
		m.stop = nil
		return
	}

	var err error

	m.port = port
	m.stop, err = midi.ListenTo(m.port, m.listen)
	if err != nil {
		log.Fatal("got error while switching midi ports:", err)
	}

	log.Printf("switched to port: %d | %s\n", port.Number(), port.String())
}

func (m *PortManager) scanner() {
	for {
		ins, err := drivers.Get().Ins()
		if err != nil {
			log.Println("failed to query midi ports:", err.Error())
			time.Sleep(time.Duration(m.opts.PollTimeout) * time.Second)
			continue
		}

		var found drivers.In
		for _, port := range ins {
			if m.opts.PortNumber > 0 && port.Number() == m.opts.PortNumber {
				found = port
				break
			}
			if m.opts.PortName != "" && strings.Contains(strings.ToLower(port.String()), m.opts.PortName) {
				found = port
				break
			}
		}

		if found == nil {
			m.switchPort(nil)
		} else if m.port == nil {
			m.switchPort(found)
		}

		time.Sleep(time.Duration(m.opts.PollTimeout) * time.Second)
	}
}

func (m *PortManager) AddListener(recv PortListener) {
	m.listeners = append(m.listeners, recv)
}

func (m *PortManager) RemoveListener(listener PortListener) {
	var newListeners []PortListener
	for _, l := range m.listeners {
		if l == listener {
			continue
		}
		newListeners = append(newListeners, l)
	}
	m.listeners = newListeners
}

func (m *PortManager) Stop() {
	m.switchPort(nil)
}
