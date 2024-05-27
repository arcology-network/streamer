package actor

import (
	"errors"

	"github.com/arcology-network/streamer/broker"
	"github.com/arcology-network/streamer/log"
	"go.uber.org/zap/zapcore"
)

type IWorker interface {
	Init(workThreadName string, broker *broker.StatefulStreamer)
	ChangeEnvironment(message *Message)
	OnStart()
	OnMessageArrived(msgs []*Message) error
}

type IWorkerEx interface {
	IWorker
	Inputs() ([]string, bool)
	Outputs() map[string]int
}

type WorkerThreadLogger struct {
	LatestMessage *Message
	Logger        *log.LogWraper
}

func (workerThdLogger *WorkerThreadLogger) Log(level string, info string, fields ...zapcore.Field) uint64 {
	return workerThdLogger.Logger.Log(level, workerThdLogger.LatestMessage, info, fields...)
}
func (workerThdLogger *WorkerThreadLogger) CheckPoint(info string, fields ...zapcore.Field) uint64 {
	return workerThdLogger.Logger.CheckPoint(log.LogLevel_Info, workerThdLogger.LatestMessage, info, fields...)
}

func (workerThdLogger *WorkerThreadLogger) GetLogger(refid uint64) *WorkerThreadLogger {
	newLog := *workerThdLogger
	newLog.LatestMessage.Msgid = refid
	return &newLog
}

type WorkerThread struct {
	Name          string
	Groupid       string
	Concurrency   int
	MsgBroker     *MessageWrapper
	LatestMessage *Message
	Log           *log.LogWraper
}

// use when function is called
func (workerThd *WorkerThread) GetLogger(refid uint64) *WorkerThreadLogger {
	innerMessage := *workerThd.LatestMessage
	return &WorkerThreadLogger{
		LatestMessage: &innerMessage,
		Logger:        workerThd.Log,
	}
}

// use where is in worklthread
func (workerThd *WorkerThread) AddLog(level string, info string, fields ...zapcore.Field) uint64 {
	return workerThd.Log.Log(level, workerThd.LatestMessage, info, fields...)
}

func (workerThd *WorkerThread) CheckPoint(info string, fields ...zapcore.Field) uint64 {
	return workerThd.Log.CheckPoint(log.LogLevel_Info, workerThd.LatestMessage, info, fields...)
}

func (workerThd *WorkerThread) Init(workThreadName string, broker *broker.StatefulStreamer) {
	workerThd.Name = workThreadName
	latestMessage := NewMessage()

	workerThd.MsgBroker = &MessageWrapper{
		MsgBroker:      broker,
		LatestMessage:  latestMessage,
		WorkThreadName: workThreadName,
	}

	workerThd.LatestMessage = latestMessage
	workerThd.Log = log.Logger.GetLogger(workThreadName)
}

func (workerThd *WorkerThread) Set(concurrency int, groupid string) {
	workerThd.Concurrency = concurrency
	workerThd.Groupid = groupid
}

func (workerThd *WorkerThread) ChangeEnvironment(message *Message) {
	// if message.Height > workerThd.LatestMessage.Height ||
	// 	(message.Height == workerThd.LatestMessage.Height && message.Round > workerThd.LatestMessage.Round) {
	workerThd.LatestMessage = message
	workerThd.MsgBroker.LatestMessage = message
	// }
}

func (workerThd *WorkerThread) IsNil(param interface{}, paramName string) (bool, error) {
	if param == nil {
		workerThd.AddLog(log.LogLevel_Error, workerThd.Name+" received params err:"+paramName+" is nil")
		return true, errors.New(workerThd.Name + " received params err:" + paramName + " is nil")
	}
	return false, nil
}
