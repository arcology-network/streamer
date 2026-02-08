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
