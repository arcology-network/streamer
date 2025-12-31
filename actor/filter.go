package actor

import (
	"fmt"

	scommon "github.com/arcology-network/streamer/common"
)

type MsgRule func(*scommon.Message) bool

type Filter struct {
	Name  string
	Rules []MsgRule
}

func NewOriginFilter(name string, rules ...MsgRule) *Filter {
	return &Filter{
		Name:  name,
		Rules: rules,
	}
}

// Apply must be:
// 1. pure function
// 2. no side effects
// 3. only filter, not mutate Message
func (f *Filter) Apply(msgs []*scommon.Message) []*scommon.Message {
	if len(msgs) != 1 {
		panic("Filter can only handle one message at a time")
	}

	msg := msgs[0]
	for _, rule := range f.Rules {
		if !rule(msg) {
			fmt.Printf("***************Msg:%v height:%v is filtered\n", msg.Name, msg.Height)
			return nil
		}
	}
	return msgs
}
func OnlyFrom(producer string) MsgRule {
	return func(msg *scommon.Message) bool {
		return msg.From == producer
	}
}

func NotFrom(producer string) MsgRule {
	return func(msg *scommon.Message) bool {
		return msg.From != producer
	}
}

func MsgsOnlyFrom(msgNames []string, producer string) MsgRule {
	blackList := make(map[string]struct{})
	for _, m := range msgNames {
		blackList[m] = struct{}{}
	}
	return func(msg *scommon.Message) bool {
		if _, ok := blackList[msg.Name]; ok {
			return msg.From == producer
		}
		return true
	}
}
