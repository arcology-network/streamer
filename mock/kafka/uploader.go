package kafka

import (
	"testing"

	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/streamer/actor"
	"github.com/arcology-network/streamer/broker"
)

type OnMessageArrivedCallback func(msgs []*actor.Message) error

type Uploader struct {
	messages        map[string]string
	broker          *actor.MessageWrapper
	counter         map[string]int
	encChan         chan *actor.Message
	dataSizeCounter map[string]int
	cb              OnMessageArrivedCallback
}

func NewUploader(_ int, _ string, messages map[string]string, _ string) actor.IWorkerEx {
	tu.Log("NewUploader")
	return &Uploader{
		messages:        messages,
		counter:         make(map[string]int),
		encChan:         make(chan *actor.Message, 1000),
		dataSizeCounter: make(map[string]int),
	}
}

func (u *Uploader) SetCallback(cb OnMessageArrivedCallback) {
	u.cb = cb
}

func (u *Uploader) Init(wtName string, broker *broker.StatefulStreamer) {
	tu.Log("Uploader.Init")
	u.broker = &actor.MessageWrapper{
		MsgBroker:      broker,
		LatestMessage:  actor.NewMessage(),
		WorkThreadName: wtName,
	}
}

func (u *Uploader) ChangeEnvironment(_ *actor.Message) {

}

func (u *Uploader) Inputs() ([]string, bool) {
	var inputs []string
	for input := range u.messages {
		inputs = append(inputs, input)
	}
	return inputs, false
}

func (u *Uploader) Outputs() map[string]int {
	return map[string]int{}
}

func (u *Uploader) OnStart() {
	tu.Log("Uploader.OnStart")
	go func() {
		for msg := range u.encChan {
			bs, err := common.GobEncode(msg)
			if err != nil {
				panic(err)
			}
			u.dataSizeCounter[msg.Name] += len(bs)
		}
	}()
}

func (u *Uploader) OnMessageArrived(msgs []*actor.Message) error {
	if u.cb != nil {
		u.cb(msgs)
	}

	for _, m := range msgs {
		u.counter[m.Name]++
		u.encChan <- m
	}
	return nil
}

// For test.
func (u *Uploader) GetCounter() map[string]int {
	return u.counter
}

func (u *Uploader) GetDataSizeCounter() map[string]int {
	return u.dataSizeCounter
}

var tu testing.TB

func NewUploaderCreator(logger testing.TB) func(int, string, map[string]string, string) actor.IWorkerEx {
	tu = logger
	return NewUploader
}
