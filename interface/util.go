package intf

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/smallnest/rpcx/protocol"
	"github.com/smallnest/rpcx/share"

	"github.com/smallnest/rpcx/client"
	rpcxserver "github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/serverplugin"
)

const (
	servicePath = "backend.service"
)

func InitZookeeperRpcClient(basepath string, zkAddrs []string) client.XClient {
	tryNums := 100
	for i := 0; i < tryNums; i++ {
		client := newZookeeperRpcClient(basepath, zkAddrs)
		if client != nil {
			return client
		}
		time.Sleep(time.Second * 1)
	}
	panic(basepath + " not found in zookeeper")
}
func newZookeeperRpcClient(basepath string, zkAddrs []string) client.XClient {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("InitZookeeperRpcClient err:%v\n", err)
		}
	}()
	share.Codecs[protocol.SerializeType(4)] = &GobCodec{}
	opt := client.DefaultOption
	opt.SerializeType = protocol.SerializeType(4)
	opt.Retries = math.MaxInt

	d := client.NewZookeeperDiscovery("/"+basepath, servicePath, zkAddrs, nil)
	return client.NewXClient(servicePath, client.Failover, client.RoundRobin, d, opt)
}

var (
	lock sync.RWMutex
)

func addprotocol() {
	lock.Lock()
	defer lock.Unlock()
	share.Codecs[protocol.SerializeType(4)] = &GobCodec{}
}
func InitZookeeperRpcServer(serviceAddr, basepath string, zkAddrs []string, rcvrs, fns []interface{}) {
	go func() {
		// rpcx service
		addprotocol()
		rpcxServer := rpcxserver.NewServer()
		register := &serverplugin.ZooKeeperRegisterPlugin{
			ServiceAddress:   "tcp@" + serviceAddr,
			ZooKeeperServers: zkAddrs,
			BasePath:         "/" + basepath,
			UpdateInterval:   time.Minute,
		}
		register.Start()
		rpcxServer.Plugins.Add(register)
		if rcvrs != nil {
			for _, svc := range rcvrs {
				rpcxServer.RegisterName(servicePath, svc, "")
			}
		}
		if fns != nil {
			for _, fn := range fns {
				rpcxServer.RegisterFunction(servicePath, fn, "")
			}
		}

		rpcxServer.Serve("tcp", serviceAddr)
	}()
}

type GobCodec struct {
}

func (c *GobCodec) Decode(data []byte, i interface{}) error {
	enc := gob.NewDecoder(bytes.NewBuffer(data))
	err := enc.Decode(i)
	return err
}

func (c *GobCodec) Encode(i interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(i)
	return buf.Bytes(), err
}
