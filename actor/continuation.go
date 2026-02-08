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
	callfram  *CallFrame
	actorname string
}

func NewDefaultContinuation(stream *broker.StatefulStreamer, actorname string, callfram *CallFrame) Continuation {
	return &DefaultContinuation{
		stream:    stream,
		actorname: actorname,
		callfram:  callfram,
	}
}

func (c *DefaultContinuation) Resume(comp *broker.RPCCompletion) {
	c.once.Do(func() {
		respmsg := comp.Payload.(*scommon.Message)
		resp := scommon.NewMessageForRPCRESP(c.callfram.ReqMsg, respmsg.Error, respmsg.Data)
		resp.WithRpcTrace(c.callfram.ReqID)

		ctx := scommon.BindLoggerContextFromMessageSafe(context.Background(), resp)
		logger.Log.Debug(ctx, c.actorname, "Continuation send rpc response")

		comp.Payload = resp
		c.stream.Send(resp.ReplyTo, comp)
	})
}

func (c *DefaultContinuation) Cancel(err error) {
	c.once.Do(func() {
		// Construct a failed RPC response message
		resp := scommon.NewMessageForRPCRESP(
			c.callfram.ReqMsg,
			err.Error(),
			nil,
		)

		ctx := scommon.BindLoggerContextFromMessageSafe(context.Background(), resp)
		logger.Log.Error(
			ctx,
			c.actorname, "rpc canceled",
			logger.F("err", err.Error()),
		)

		comp := &broker.RPCCompletion{
			ID:      c.callfram.ReqID,
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
	Frame   *CallFrame
}

type BusinessContinuation struct {
	actor     *Actor
	once      sync.Once
	callframe *CallFrame
}

func NewBusinessContinuation(actor *Actor, callframe *CallFrame) Continuation {
	return &BusinessContinuation{
		actor:     actor,
		callframe: callframe,
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
			Frame:   c.callframe,
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
