package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/eagraf/habitat-node/entities"
	"github.com/eagraf/habitat-node/entities/transitions"
	"github.com/eagraf/habitat-node/orchestrator/processes"
)

// subscriber types
const (
	ProcessManagerSubscriber string = "process_manager"
)

type Sequence struct {
	Transitions []*transitions.TransitionWrapper `json:"transitions"`
}

type Sequencer interface {
	Start() error
	Next(*transitions.TransitionWrapper) error
	Stop() error
}

type TransitionSubscriberSequencer struct {
	subscriber     transitions.TransitionSubscriber
	subscriberType string
}

func (s *TransitionSubscriberSequencer) Start() error {
	switch s.subscriberType {
	case ProcessManagerSubscriber:
		state := entities.InitState()
		s.subscriber = processes.InitManager()
		s.subscriber.(*processes.ProcessManager).Start(state)
	default:
		panic(fmt.Sprintf("subsriber type %s not supported", s.subscriberType))
	}

	return nil
}

func (s *TransitionSubscriberSequencer) Next(transition *transitions.TransitionWrapper) error {
	return s.subscriber.Receive(transition.Transition)
}

func (s *TransitionSubscriberSequencer) Stop() error {
	switch s.subscriberType {
	case ProcessManagerSubscriber:
		s.subscriber.(*processes.ProcessManager).Stop()
	default:
		panic(fmt.Sprintf("subsriber type %s not supported", s.subscriberType))
	}
	return nil
}

func main() {

	if len(os.Args) != 3 {
		panic("usage: test-suite <sequencer_type> <test_file>")
	}

	sequencerType := os.Args[1]

	file := os.Args[2]
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}

	sequence := &Sequence{}
	err = json.Unmarshal(buffer, sequence)
	if err != nil {
		panic(err)
	}

	var sequencer Sequencer
	switch sequencerType {
	case ProcessManagerSubscriber:
		sequencer = &TransitionSubscriberSequencer{
			subscriberType: sequencerType,
		}
	}

	err = sequencer.Start()
	if err != nil {
		panic(err)
	}

	for _, transition := range sequence.Transitions {
		//err := subscriber.Receive(transition.Transition)
		err := sequencer.Next(transition)
		if err != nil {
			panic(err)
		}
	}

	err = sequencer.Stop()
	if err != nil {
		panic(err)
	}

	// Make sure there is enought time to SIGKILL child processes
	time.Sleep(5 * time.Second)
}
