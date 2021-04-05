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

func main() {

	if len(os.Args) != 3 {
		panic("usage: test-suite <subscriber_type> <test_file>")
	}

	subscriberType := os.Args[1]

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

	var subscriber transitions.TransitionSubscriber
	switch subscriberType {
	case ProcessManagerSubscriber:
		state := entities.InitState()
		subscriber = processes.InitManager()
		subscriber.(*processes.ProcessManager).Start(state)
	default:
		panic(fmt.Sprintf("subsriber type %s not supported", subscriberType))
	}

	for _, transition := range sequence.Transitions {
		err := subscriber.Receive(transition.Transition)
		if err != nil {
			panic(err)
		}
	}

	// Takedown if necessary
	switch subscriberType {
	case ProcessManagerSubscriber:
		fmt.Println("stopping")
		subscriber.(*processes.ProcessManager).Stop()
	default:
		panic(fmt.Sprintf("subsriber type %s not supported", subscriberType))
	}
	// Make sure there is enought time to SIGKILL child processes
	time.Sleep(5 * time.Second)
}
