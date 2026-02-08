package common

import (
	"context"
	"encoding/gob"
	"reflect"
	"time"

	"github.com/arcology-network/streamer/logger"
	"github.com/google/uuid"
)

const (
	MsgTypeRPCReq  = "RPC_REQ"
	MsgTypeRPCResp = "RPC_RESP"
	MsgTypeStream  = "STREAM"
)

type Message struct {
	// ===== Message Basic Attributes =====
	ID        string // Globally unique message ID (uuid)
	Name      string // Business topic / Stream Topic / rpc.Service
	Type      string // "RPC_REQ" | "RPC_RESP" | "STREAM"
	Timestamp int64  // Sending time (UnixNano)

	Height uint64 // Blocknumber
	From   string //sender, workthread name

	// ===== Distributed Tracing =====
	TraceID  string // ID for a complete business process chain
	SpanID   string // Current node1
	ParentID string // Upstream Span
	SpanName string // Fixed Logic Name (Module Name/Function Name)

	// ===== RPC Specific =====
	Service  string //UserService
	Method   string //Add
	ReplyTo  string //Address of Client
	Error    string
	ReqID    string // Used for synchronous to asynchronous scenarios.
	NextStep string // For next Step
	ContId   string // For rpc query

	// ===== Stream Processing Control =====
	Sequence  uint64 // JetStream sequence
	Redeliver int    // Number of deliveries (retry count)
	Key       string // KV Key

	// ===== Message Payload =====
	Data interface{} // ✅ Note: Used as an object in-process, and as []byte when passing through jetStream.
}

func init() {
	gob.Register(&Message{})
}

func NewMessageForStream(name string, payload any) *Message {
	return &Message{
		ID:        uuid.NewString(),
		Name:      name,
		Type:      MsgTypeStream,
		Timestamp: time.Now().UnixNano(),
		Data:      payload,
	}
}

func NewMessageForRPCREQ(service, method string, payload any) *Message {
	return &Message{
		ID:        uuid.New().String(),
		Name:      service,
		Type:      MsgTypeRPCReq,
		Timestamp: time.Now().UnixNano(),
		Service:   service,
		Method:    method,
		// ReqID:     id,
		Data: payload,
	}
}

func NewMessageForRPCRESP(req *Message, errMsg string, respPayload any) *Message {
	return &Message{
		ID:        uuid.New().String(),
		Name:      req.ReplyTo,
		Type:      MsgTypeRPCResp,
		Timestamp: time.Now().UnixNano(),
		TraceID:   req.TraceID,
		Height:    req.Height,
		From:      req.From,
		ParentID:  req.SpanID,
		Service:   req.Service,
		Method:    req.Method,
		ReplyTo:   req.ReplyTo,
		ReqID:     req.ReqID,
		NextStep:  req.NextStep,
		ContId:    req.ContId,
		Key:       req.Key,
		Error:     errMsg,
		Data:      respPayload,
	}
}
func (m *Message) WithRpcTrace(requestID string) {
	m.ReqID = requestID
}
func (m *Message) GetTrace() (TraceId string, ParentId string) {
	return m.TraceID, m.SpanID
}

// DeepCopyMessage
func (m *Message) DeepCopy() *Message {
	if m == nil {
		return nil
	}

	copyMsg := &Message{
		ID:        m.ID,
		Name:      m.Name,
		Type:      m.Type,
		Timestamp: m.Timestamp,
		Height:    m.Height,
		From:      m.From,
		TraceID:   m.TraceID,
		SpanID:    m.SpanID,
		ParentID:  m.ParentID,
		SpanName:  m.SpanName,
		Service:   m.Service,
		Method:    m.Method,
		ReplyTo:   m.ReplyTo,
		Error:     m.Error,
		ReqID:     m.ReqID,
		NextStep:  m.NextStep,
		Sequence:  m.Sequence,
		Redeliver: m.Redeliver,
		Key:       m.Key,
	}

	// Payload uses reflection for deep copy
	copyMsg.Data = deepCopyInterface(m.Data)

	return copyMsg
}

// deepCopyInterface performs a deep copy of an interface{} value (implemented via reflection)
func deepCopyInterface(src interface{}) interface{} {
	if src == nil {
		return nil
	}

	val := reflect.ValueOf(src)
	return deepCopyValue(val).Interface()
}

func deepCopyValue(val reflect.Value) reflect.Value {
	if !val.IsValid() {
		return val
	}

	switch val.Kind() {
	case reflect.Ptr:
		if val.IsNil() {
			return reflect.Zero(val.Type())
		}
		ptrCopy := reflect.New(val.Elem().Type())
		ptrCopy.Elem().Set(deepCopyValue(val.Elem()))
		return ptrCopy

	case reflect.Slice:
		if val.IsNil() {
			return reflect.Zero(val.Type())
		}
		sliceCopy := reflect.MakeSlice(val.Type(), val.Len(), val.Cap())
		for i := 0; i < val.Len(); i++ {
			sliceCopy.Index(i).Set(deepCopyValue(val.Index(i)))
		}
		return sliceCopy

	case reflect.Map:
		if val.IsNil() {
			return reflect.Zero(val.Type())
		}
		mapCopy := reflect.MakeMapWithSize(val.Type(), val.Len())
		for _, key := range val.MapKeys() {
			mapCopy.SetMapIndex(key, deepCopyValue(val.MapIndex(key)))
		}
		return mapCopy

	case reflect.Struct:
		structCopy := reflect.New(val.Type()).Elem()
		for i := 0; i < val.NumField(); i++ {
			if structCopy.Field(i).CanSet() {
				structCopy.Field(i).Set(deepCopyValue(val.Field(i)))
			}
		}
		return structCopy

	default:
		// Basic types are returned directly
		return val
	}
}

func BindLoggerContextFromMessageSafe(
	ctx context.Context,
	msg *Message,
) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if msg == nil {
		return ctx
	}

	// -------- Trace fallback --------
	traceID := msg.TraceID
	if traceID == "" {
		traceID = "trace-unknown"
	}

	spanID := msg.SpanID
	if spanID == "" {
		spanID = "span-unknown"
	}

	parentID := msg.ParentID

	ctx = logger.WithTrace(ctx, logger.TraceContext{
		TraceID:  traceID,
		SpanID:   spanID,
		ParentID: parentID,
		SpanName: msg.SpanName,
	})

	// -------- Message Meta fallback --------
	msgType := msg.Type
	if msgType == "" {
		msgType = "UNKNOWN"
	}

	ctx = logger.WithMsg(ctx, logger.MsgContext{
		MsgType:   msgType,
		Name:      msg.Name,
		Key:       msg.Key,
		Sequence:  msg.Sequence,
		Redeliver: msg.Redeliver,
		Height:    msg.Height,
		From:      msg.From,
		Method:    msg.Method,
		ReqID:     msg.ReqID,
	})

	return ctx
}
