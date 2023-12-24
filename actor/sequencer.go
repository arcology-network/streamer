package actor

import streamer "github.com/arcology-network/component-lib/broker"

type sequencer struct {
	worker IWorkerEx

	defs  map[int][]string
	state int
}

func AsSequencer(worker IWorkerEx, sequence []string) *sequencer {
	defs := make(map[int][]string)
	for index, item := range sequence {
		defs[index] = []string{item}
	}

	return &sequencer{
		worker: worker,
		defs:   defs,
	}
}

func (s *sequencer) Init(wtName string, broker *streamer.StatefulBroker) {
	s.worker.Init(wtName, broker)
}

func (s *sequencer) ChangeEnvironment(message *Message) {
	s.worker.ChangeEnvironment(message)
}

func (s *sequencer) Inputs() ([]string, bool) {
	return s.worker.Inputs()
}

func (s *sequencer) Outputs() map[string]int {
	return s.worker.Outputs()
}

func (s *sequencer) Config(params map[string]interface{}) {
	if c, ok := s.worker.(Configurable); ok {
		c.Config(params)
	} else {
		panic("not configurable")
	}
}

func (s *sequencer) OnStart() {
	s.worker.OnStart()
}

func (s *sequencer) OnMessageArrived(msgs []*Message) error {
	if s.state == len(s.defs)-1 {
		s.state = 0
	} else {
		s.state++
	}
	return s.worker.OnMessageArrived(msgs)
}

func (s *sequencer) GetStateDefinitions() map[int][]string {
	return s.defs
}

func (s *sequencer) GetCurrentState() int {
	return s.state
}
