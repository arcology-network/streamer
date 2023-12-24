package actor

import (
	streamer "github.com/arcology-network/component-lib/broker"
)

const (
	RenamerPrefix = "__rename__"
)

func RenamerName(from, to string) string {
	return RenamerPrefix + from + "-" + to
}

type Renamer struct {
	WorkerThread
	from string
	to   string
}

func Rename(from string) *Renamer {
	return &Renamer{from: from}
}

func (r *Renamer) To(to string) *Renamer {
	r.to = to
	return r
}

func (r *Renamer) On(broker *streamer.StatefulBroker) *Renamer {
	renamer := NewActorEx(
		RenamerName(r.from, r.to),
		broker,
		r,
	)
	renamer.Connect(streamer.NewDisjunctions(renamer, 1))
	return r
}

func (r *Renamer) Inputs() ([]string, bool) {
	return []string{r.from}, false
}

func (r *Renamer) Outputs() map[string]int {
	return map[string]int{
		r.to: 1,
	}
}

func (r *Renamer) OnStart() {

}

func (r *Renamer) OnMessageArrived(msgs []*Message) error {
	for _, v := range msgs {
		switch v.Name {
		case r.from:
			r.MsgBroker.Send(r.to, v.Data)
		}
	}

	return nil
}
