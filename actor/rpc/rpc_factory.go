package rpc

import (
	"errors"
	"sync"
)

type RPCFactory struct {
	m    map[string]interface{}
	once sync.Once
}

var GlobalRPCFactory *RPCFactory

func InitGlobalRPCFactory() {
	GlobalRPCFactory = &RPCFactory{
		m:    map[string]interface{}{},
		once: sync.Once{},
	}
}

func (c *RPCFactory) Register(serviceName string, handle interface{}) error {
	if _, ok := c.m[serviceName]; ok {
		return errors.New("Service duplicate registration: " + serviceName)
	}
	c.m[serviceName] = handle
	return nil
}

func (c *RPCFactory) GetHandle(serviceName string) (interface{}, bool) {
	h, ok := c.m[serviceName]
	return h, ok
}

// func HandleRPC(handle interface{}, methodName string, Payload interface{}) (respPayload interface{}, err error) {
// 	targetValue := reflect.ValueOf(handle)
// 	method := targetValue.MethodByName(methodName)
// 	if !method.IsValid() {
// 		log.Printf("method not found: %v", methodName)
// 		err = errors.New(fmt.Sprintf("method not found: %v", methodName))
// 		return
// 	}

// 	mType := method.Type()
// 	// 强制检查签名：
// 	// func(ctx context.Context, arg X) (Y, error)
// 	if mType.NumIn() != 2 {
// 		log.Printf("method must have exactly 2 input args (ctx, arg)")
// 		err = fmt.Errorf("method must have exactly 2 input args (ctx, arg)")
// 		return
// 	}
// 	if mType.NumOut() != 2 {
// 		log.Printf("method must return (value, error)")
// 		err = fmt.Errorf("method must return (value, error)")
// 		return
// 	}

// 	// 第一个参数必须是 context.Context
// 	ctxInterface := reflect.TypeOf((*context.Context)(nil)).Elem()
// 	if !mType.In(0).Implements(ctxInterface) {
// 		log.Printf("first argument must be context.Context")
// 		err = fmt.Errorf("first argument must be context.Context")
// 		return
// 	}

// 	// 构造 arg
// 	argVal, err := buildArg(Payload, mType.In(1))
// 	if err != nil {
// 		log.Printf("param1 err:%v", err)
// 		return
// 	}

// 	// args := reflect.ValueOf(Payload)
// 	ctx := context.Background()
// 	// ctx = context.WithValue(ctx, "H", "Are you see?")
// 	// 调用方法
// 	ret := method.Call([]reflect.Value{
// 		reflect.ValueOf(ctx),
// 		argVal,
// 	})

// 	if len(ret) == 2 && !ret[1].IsNil() {
// 		err = ret[1].Interface().(error)
// 	} else {
// 		respPayload = ret[0].Interface()
// 	}
// 	fmt.Printf("resp:%v\n", respPayload)
// 	return
// }

// func buildArg(payload interface{}, targetType reflect.Type) (reflect.Value, error) {
// 	// nil payload
// 	if payload == nil {
// 		// 只能传给 interface / pointer / slice / map
// 		switch targetType.Kind() {
// 		case reflect.Interface, reflect.Ptr, reflect.Slice, reflect.Map:
// 			return reflect.Zero(targetType), nil
// 		default:
// 			return reflect.Value{}, fmt.Errorf("nil payload cannot be used for %v", targetType)
// 		}
// 	}

// 	v := reflect.ValueOf(payload)
// 	t := v.Type()

// 	// 完全匹配
// 	if t.AssignableTo(targetType) {
// 		return v, nil
// 	}

// 	// 可转换
// 	if t.ConvertibleTo(targetType) {
// 		return v.Convert(targetType), nil
// 	}

// 	// 值 -> 指针
// 	if targetType.Kind() == reflect.Ptr && t.AssignableTo(targetType.Elem()) {
// 		ptr := reflect.New(targetType.Elem())
// 		ptr.Elem().Set(v)
// 		return ptr, nil
// 	}

// 	// 指针 -> 值
// 	if t.Kind() == reflect.Ptr && t.Elem().AssignableTo(targetType) {
// 		return v.Elem(), nil
// 	}

// 	// interface 参数
// 	if targetType.Kind() == reflect.Interface && t.Implements(targetType) {
// 		return v, nil
// 	}

// 	return reflect.Value{}, fmt.Errorf(
// 		"payload type %v cannot be used as %v",
// 		t, targetType,
// 	)
// }
