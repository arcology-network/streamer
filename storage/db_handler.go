package storage

import (
	"fmt"
	"time"

	cmntyp "github.com/arcology-network/common-lib/types"
	"github.com/arcology-network/component-lib/actor"
	"github.com/arcology-network/component-lib/log"
	ccurl "github.com/arcology-network/concurrenturl"

	"github.com/arcology-network/concurrenturl/interfaces"
	ccdb "github.com/arcology-network/concurrenturl/storage"
)

type DBOperation interface {
	Init(db interfaces.Datastore, url *ccurl.ConcurrentUrl, broker *actor.MessageWrapper)
	Import(transitions []interfaces.Univalue)
	PostImport(euResults []*cmntyp.EuResult, height uint64)
	PreCommit(euResults []*cmntyp.EuResult, height uint64)
	PostCommit(euResults []*cmntyp.EuResult, height uint64)
	Finalize()
	Outputs() map[string]int
	Config(params map[string]interface{})
}

type BasicDBOperation struct {
	DB        interfaces.Datastore
	URL       *ccurl.ConcurrentUrl
	MsgBroker *actor.MessageWrapper
}

func (op *BasicDBOperation) Init(db interfaces.Datastore, url *ccurl.ConcurrentUrl, broker *actor.MessageWrapper) {
	op.DB = db
	op.URL = url
	op.MsgBroker = broker
}

func (op *BasicDBOperation) Import(transitions []interfaces.Univalue) {
	op.URL.Import(transitions)
}

func (op *BasicDBOperation) PostImport(euResults []*cmntyp.EuResult, height uint64) {
	if height == 0 {
		_, transitions := GetTransitions(euResults)
		op.URL.Import(transitions)
	}
	op.URL.Sort()
}

func (op *BasicDBOperation) PreCommit(euResults []*cmntyp.EuResult, height uint64) {
	op.URL.Finalize(GetTransitionIds(euResults))
}

func (op *BasicDBOperation) PostCommit(euResults []*cmntyp.EuResult, height uint64) {
	// if len(euResults) > 0 {
	// 	fmt.Printf("stop\n")
	// }
	op.URL.WriteToDbBuffer()
}

func (op *BasicDBOperation) Finalize() {
	op.URL.SaveToDB()
}

func (op *BasicDBOperation) Outputs() map[string]int {
	return map[string]int{}
}

func (op *BasicDBOperation) Config(params map[string]interface{}) {}

const (
	dbStateUninit = iota
	dbStateInit
	dbStateDone
)

type DBHandler struct {
	actor.WorkerThread

	db          interfaces.Datastore
	url         *ccurl.ConcurrentUrl
	state       int
	commitMsg   string
	finalizeMsg string
	op          DBOperation
}

func NewDBHandler(concurrency int, groupId string, commitMsg, finalizeMsg string, op DBOperation) *DBHandler {
	handler := &DBHandler{
		state:       dbStateUninit,
		commitMsg:   commitMsg,
		finalizeMsg: finalizeMsg,
		op:          op,
	}
	handler.Set(concurrency, groupId)
	return handler
}

func (handler *DBHandler) Inputs() ([]string, bool) {
	msgs := []string{actor.MsgEuResults, handler.commitMsg, handler.finalizeMsg}
	if handler.state == dbStateUninit {
		msgs = append(msgs, actor.MsgInitDB)
	}
	return msgs, false
}

func (handler *DBHandler) Outputs() map[string]int {
	return handler.op.Outputs()
}

func (handler *DBHandler) Config(params map[string]interface{}) {
	if v, ok := params["init_db"]; !ok {
		panic("parameter not found: init_db")
	} else {
		if !v.(bool) {			

			handler.db = ccdb.NewParallelEthMemDataStore()

			handler.url = ccurl.NewConcurrentUrl(handler.db)

			handler.op.Init(handler.db, handler.url, handler.MsgBroker)
			handler.state = dbStateDone
		}
	}
	handler.op.Config(params)
}

func (handler *DBHandler) OnStart() {
	handler.op.Init(handler.db, handler.url, handler.MsgBroker)
}

func (handler *DBHandler) OnMessageArrived(msgs []*actor.Message) error {
	msg := msgs[0]
	switch handler.state {
	case dbStateUninit:
		if msg.Name == actor.MsgInitDB {
			handler.db = msg.Data.(interfaces.Datastore)
			handler.url = ccurl.NewConcurrentUrl(handler.db)

			handler.op.Init(handler.db, handler.url, handler.MsgBroker)
			handler.state = dbStateDone
		}
	case dbStateInit:
		if msg.Name == actor.MsgEuResults {
			data := msg.Data.(*cmntyp.Euresults)
			t1 := time.Now()
			_, transitions := GetTransitions(*data)
			handler.op.Import(transitions)
			fmt.Printf("DBHandler Euresults import height:%v,tim:%v\n", msg.Height, time.Since(t1))
		} else if msg.Name == handler.commitMsg {
			var data []*cmntyp.EuResult
			if msg.Data != nil {
				for _, item := range msg.Data.([]interface{}) {
					data = append(data, item.(*cmntyp.EuResult))
				}
			}

			handler.AddLog(log.LogLevel_Info, "Before PostImport.")
			handler.op.PostImport(data, msg.Height)
			handler.AddLog(log.LogLevel_Info, "Before PreCommit.")
			handler.op.PreCommit(data, msg.Height)
			handler.AddLog(log.LogLevel_Info, "Before PostCommit.")
			handler.op.PostCommit(data, msg.Height)
			handler.AddLog(log.LogLevel_Info, "After PostCommit.")
			handler.state = dbStateDone
		}
	case dbStateDone:
		if msg.Name == handler.finalizeMsg {
			handler.op.Finalize()
			handler.state = dbStateInit
		}
	}
	return nil
}

func (handler *DBHandler) GetStateDefinitions() map[int][]string {
	return map[int][]string{
		dbStateUninit: {actor.MsgInitDB},
		dbStateInit:   {actor.MsgEuResults, handler.commitMsg},
		dbStateDone:   {actor.MsgEuResults, handler.finalizeMsg},
	}
}

func (handler *DBHandler) GetCurrentState() int {
	return handler.state
}
