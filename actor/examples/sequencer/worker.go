package sequencer

import (
	"fmt"

	"github.com/arcology-network/streamer/actor"
)

const (
	msgA = "a"
	msgB = "b"
	msgC = "c"
)

type Worker struct {
	actor.WorkerThread
}

func NewWorker() *Worker {
	return &Worker{}
}

func (w *Worker) Inputs() ([]string, bool) {
	return []string{msgA, msgB, msgC}, false
}

func (w *Worker) Outputs() map[string]int {
	return map[string]int{}
}

func (w *Worker) OnStart() {}

func (w *Worker) OnMessageArrived(msgs []*actor.Message) error {
	msg := msgs[0]
	fmt.Printf("Message: %v\n", msg)
	return nil
}
