package actor

import (
	"context"
	"sync"

	"github.com/arcology-network/streamer/broker"
	scommon "github.com/arcology-network/streamer/common"
	"github.com/arcology-network/streamer/logger"
)

type DefaultContinuation struct {
	stream    *broker.StatefulStreamer
	once      sync.Once
	reqMsg    *scommon.Message
	actorname string
}

func NewDefaultContinuation(stream *broker.StatefulStreamer, actorname string, reqMsg *scommon.Message) Continuation {
	return &DefaultContinuation{
		stream:    stream,
		actorname: actorname,
		reqMsg:    reqMsg,
	}
}

func (c *DefaultContinuation) Resume(comp *broker.RPCCompletion) {
	c.once.Do(func() {
		respmsg := comp.Payload.(*scommon.Message)
		resp := scommon.NewMessageForRPCRESP(c.reqMsg, respmsg.Error, respmsg.Data)
		// resp.Service = resp.ReplyTo
		ctx := scommon.BindLoggerContextFromMessageSafe(context.Background(), resp)
		logger.Log.Debug(ctx, c.actorname+" send rpc response")
		comp.Payload = resp
		c.stream.Send(resp.ReplyTo, comp)
	})
}

func (c *DefaultContinuation) Cancel(err error) {
	c.once.Do(func() {
		// Construct a failed RPC response message
		resp := scommon.NewMessageForRPCRESP(
			c.reqMsg,
			err.Error(),
			nil,
		)

		ctx := scommon.BindLoggerContextFromMessageSafe(context.Background(), resp)
		logger.Log.Error(
			ctx,
			c.actorname+" rpc canceled",
			logger.F("err", err.Error()),
		)

		comp := &broker.RPCCompletion{
			ID:      c.reqMsg.ReqID,
			Payload: resp,
			Error:   err.Error(),
		}

		// Notify the caller (critical)
		c.stream.Send(resp.ReplyTo, comp)
	})
}

type ContinuationNext struct {
	FlowID  string // Optional: ID of a business process
	Step    string // nextStep
	Payload any
	Error   string // Optional
}

type BusinessContinuation struct {
	actor *Actor
	once  sync.Once
}

func NewBusinessContinuation(actor *Actor) Continuation {
	return &BusinessContinuation{
		actor: actor,
	}
}

func (c *BusinessContinuation) Resume(comp *broker.RPCCompletion) {
	c.once.Do(func() {
		msg := comp.Payload.(*scommon.Message)

		next := &ContinuationNext{
			FlowID:  msg.ReqID,
			Step:    msg.NextStep,
			Payload: msg,
			Error:   msg.Error,
		}

		c.actor.Consume(next)
	})
}

func (c *BusinessContinuation) Cancel(err error) {
	c.once.Do(func() {
		next := &ContinuationNext{
			Error: err.Error(),
		}
		c.actor.Consume(next)
	})
}
