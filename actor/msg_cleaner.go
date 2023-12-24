package actor

import "fmt"

type CheckFunc func(msg *Message) bool

type MsgCleaner struct {
	BaseLinkedActor

	isOk CheckFunc
}

func NewMsgCleaner(isOk CheckFunc) *MsgCleaner {
	cleaner := &MsgCleaner{
		isOk: isOk,
	}
	cleaner.BaseLinkedActor.SetDerived(cleaner)
	return cleaner
}

func NewCleaner(concurrency int, groupId string) IWorkerEx {
	cleaner := &MsgCleaner{}
	cleaner.Set(concurrency, groupId)
	cleaner.BaseLinkedActor.SetDerived(cleaner)
	return cleaner
}

func (cleaner *MsgCleaner) Config(params map[string]interface{}) {
	if _, ok := params["func"]; !ok {
		panic("invalid params")
	}

	f := params["func"].(string)
	if len(f) < 5 || f[:4] != "not:" {
		panic("invalid cleaner func " + f)
	}
	cleaner.isOk = NotFrom(f[4:])
}

func (cleaner *MsgCleaner) Preprocess(msgs []*Message) ([]*Message, error) {
	if len(msgs) != 1 {
		panic("cannot handle more than one message.")
	}
	msg := msgs[0]

	if cleaner.isOk(msg) {
		return msgs, nil
	}

	fmt.Printf("MC: Ignore %v\n", msg)
	return nil, nil
}

func (cleaner *MsgCleaner) Postprocess(msgs []*Message) error {
	return nil
}

func (cleaner *MsgCleaner) OnStart() {
	cleaner.BaseLinkedActor.SetDerived(cleaner)
	cleaner.BaseLinkedActor.OnStart()
}

func OnlyFrom(producer string) CheckFunc {
	return func(msg *Message) bool {
		return msg.From == producer
	}
}

func NotFrom(producer string) CheckFunc {
	return func(msg *Message) bool {
		return msg.From != producer
	}
}

func MsgsOnlyFrom(msgs []string, producer string) CheckFunc {
	blackList := make(map[string]struct{})
	for _, msg := range msgs {
		blackList[msg] = struct{}{}
	}

	return func(msg *Message) bool {
		if _, ok := blackList[msg.Name]; ok {
			return msg.From == producer
		}
		return true
	}
}
