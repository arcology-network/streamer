package kafka

import (
	"strings"

	"github.com/arcology-network/common-lib/types"
	"github.com/arcology-network/component-lib/actor"
	"github.com/arcology-network/component-lib/kafka/lib"
)

type ExecutingLogServer struct {
	outgoing *lib.ComOutgoing
	mqaddr   string
	topic    string
	msgtype  string
	logsChan chan *types.ExecutingLogsMessage
}

// return a Subscriber struct
func NewExecutingLogServer(logsChan chan *types.ExecutingLogsMessage, mqaddr, topic, msgtype string) *ExecutingLogServer {
	server := ExecutingLogServer{}
	server.logsChan = logsChan
	server.topic = topic
	server.msgtype = msgtype
	server.mqaddr = mqaddr
	return &server
}

func (els *ExecutingLogServer) Start() {
	els.outgoing = new(lib.ComOutgoing)
	relations := map[string]string{}
	relations[els.msgtype] = els.topic
	if err := els.outgoing.Start(strings.Split(els.mqaddr, ","), relations, "ExecutingLogServer"); err != nil {
		panic(err)
	}
	exitChan := make(chan bool, 0)
	go func() {
		for {
			select {
			case exectorlogs := <-els.logsChan:
				msg := actor.Message{
					Name:   els.msgtype,
					Data:   exectorlogs.Logs,
					Msgid:  exectorlogs.Msgid,
					Height: exectorlogs.Height,
					Round:  exectorlogs.Round,
				}
				els.outgoing.Send(&msg)
			case <-exitChan:
				break
			}
		}
	}()

}

func (ku *ExecutingLogServer) OnMessageArrived(msgs []*actor.Message) error {
	ku.outgoing.Send(msgs[0])
	return nil
}
