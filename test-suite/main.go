package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/eagraf/habitat-node/entities"
	"github.com/eagraf/habitat-node/orchestrator/processes"
)

// subscriber types
const (
	ProcessManagerSubscriber string = "process_manager"
)

type Sequence struct {
	Transitions []*entities.TransitionWrapper `json:"transitions"`
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

	var subscriber entities.TransitionSubscriber
	switch subscriberType {
	case ProcessManagerSubscriber:
		subscriber = processes.InitManager()
	default:
		panic(fmt.Sprintf("subsriber type %s not supported", subscriberType))

	}

	for _, transition := range sequence.Transitions {
		err := subscriber.Receive(transition.Transition)
		if err != nil {
			panic(err)
		}
	}
}
