package actor

import (
	"context"

	"github.com/arcology-network/streamer/actor/rpc"
	"github.com/arcology-network/streamer/broker"
	scommon "github.com/arcology-network/streamer/common"
	"github.com/arcology-network/streamer/logger"
	"github.com/google/uuid"
)

type Sendable interface {
	SetSender(sender OutboundSender)
}

type OutboundSender interface {
	Send(name string, data interface{}, height uint64)
	SendSync(serverName, methodName string, payload interface{}, height uint64) (respPayload interface{}, err error)
}

type SendAdaptor struct {
	msgSender *broker.StatefulStreamer
	rpcCaller *rpc.RPCClient
}

func NewSendAdaptor(msgsender *broker.StatefulStreamer, rpcsender *rpc.RPCClient) *SendAdaptor {
	return &SendAdaptor{
		msgSender: msgsender,
		rpcCaller: rpcsender,
	}
}

func (s *SendAdaptor) SendSync(serverName, methodName string, payload interface{}, height uint64) (respPayload interface{}, err error) {
	allServerName := "rpc." + serverName
	msg := scommon.NewMessageForRPCREQ(allServerName, methodName, payload)
	msg.StartRpcTrace(msg.ID)
	msg.Height = height
	ctxlog := scommon.BindLoggerContextFromMessageSafe(context.Background(), msg)
	logger.Log.Debug(ctxlog, "Send rpc Msg from external")

	resp, err := s.rpcCaller.SyncCall(allServerName, methodName, msg.ID, msg, rpc.RpcTimeout)
	if err != nil {
		return nil, err
	}
	return resp.Payload.(*scommon.Message).Data, nil
}
func (s *SendAdaptor) Send(name string, data interface{}, height uint64) {
	msg := scommon.NewMessageForStream(name, data)
	msg.Height = height
	msg.ReqID = uuid.NewString()
	s.msgSender.Send(name, msg)

	ctxlog := scommon.BindLoggerContextFromMessageSafe(context.Background(), msg)
	logger.Log.Debug(ctxlog, "Send Msg from external")
}
