package aggrselector

import (
	"fmt"

	"github.com/arcology-network/component-lib/actor"
)

type DataPreprocessor struct {
	actor.WorkerThread
}

func NewDataPreprocessor() *DataPreprocessor {
	return &DataPreprocessor{}
}

func (dpp *DataPreprocessor) Inputs() ([]string, bool) {
	return []string{msgData}, false
}

func (dpp *DataPreprocessor) Outputs() map[string]int {
	return map[string]int{}
}

func (dpp *DataPreprocessor) OnStart() {}

func (dpp *DataPreprocessor) OnMessageArrived(msgs []*actor.Message) error {
	msg := msgs[0]
	if msg.Name == msgData {
		fmt.Printf("DPP: %v\n", msg)
	}
	return nil
}

type DataPreprocessorV2 struct {
	actor.WorkerThread
}

func NewDataPreprocessorV2() *DataPreprocessorV2 {
	return &DataPreprocessorV2{}
}

func (dpp *DataPreprocessorV2) Inputs() ([]string, bool) {
	return []string{msgData}, false
}

func (dpp *DataPreprocessorV2) Outputs() map[string]int {
	return map[string]int{}
}

func (dpp *DataPreprocessorV2) OnStart() {}

func (dpp *DataPreprocessorV2) OnMessageArrived(msgs []*actor.Message) error {
	msg := msgs[0]
	if msg.Name == msgData {
		fmt.Printf("DPP: %v\n", msg)
		dpp.MsgBroker.Send(msg.Name, msg.Data)
	}
	return nil
}
