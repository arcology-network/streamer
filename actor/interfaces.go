package actor

import (
	"github.com/arcology-network/streamer/broker"
	scommon "github.com/arcology-network/streamer/common"
)

type Business interface {
	Inputs() ([]string, bool)
	Outputs() map[string]int
	RegisterActions(reg ActionRegistrar)
	// BusinessInfo() string
}

type Gateable interface {
	SetActor(act *Actor)
}

type Externalable interface {
	Inject(data interface{})
}

type RpcConfigurable interface {
	RpcConfig() (string, int)
}

type Initializer interface {
	InitMsgs() []*scommon.Message
}

// ---------------------------------------------
type IWorker interface {
	Init(owner IWorker, workThreadName string, broker *broker.StatefulStreamer) *ActionRouter
	ChangeEnvironment(message *scommon.Message)
	OnStart()
	OnMessageArrived(msgs []*scommon.Message) error
	// ActionRegister()
}

type IWorkerEx interface {
	IWorker
	Inputs() ([]string, bool)
	Outputs() map[string]int
	RpcConfig() (string, int)
}
