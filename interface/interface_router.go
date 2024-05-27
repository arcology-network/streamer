/*
 *   Copyright (c) 2024 Arcology Network

 *   This program is free software: you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation, either version 3 of the License, or
 *   (at your option) any later version.

 *   This program is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *   GNU General Public License for more details.

 *   You should have received a copy of the GNU General Public License
 *   along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package intf

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/smallnest/rpcx/client"
)

type RPCServerCreator func(serviceAddr, basepath string, zkAddrs []string, rcvrs, fns []interface{})

var (
	RPCCreator RPCServerCreator = InitZookeeperRpcServer
)

type InterfaceRouter struct {
	guard             sync.RWMutex
	locals            map[string]interface{}
	remotes           map[string]client.XClient
	zkServers         []string
	availableServices []string
}

func (router *InterfaceRouter) SetZkServers(servers []string) {
	router.zkServers = servers
}

func (router *InterfaceRouter) SetAvailableServices(services []string) {
	router.availableServices = services
}

func (router *InterfaceRouter) GetAvailableServices() []string {
	return router.availableServices
}

func (router *InterfaceRouter) Register(name string, callee interface{}, addr string, zookeeper string) {
	router.guard.Lock()
	defer router.guard.Unlock()

	router.locals[name] = callee
	RPCCreator(addr, name, []string{zookeeper}, []interface{}{callee}, nil)
}

func (router *InterfaceRouter) Call(name string, f string, args, reply interface{}) error {
	router.guard.RLock()

	if local, ok := router.locals[name]; ok {
		method := reflect.ValueOf(local).MethodByName(f)
		if method.Type().NumIn() != 3 {
			panic(fmt.Sprintf("expected 3 arguments, got %d", method.Type().NumIn()))
		}
		res := method.Call([]reflect.Value{reflect.ValueOf(context.Background()), reflect.ValueOf(args), reflect.ValueOf(reply)})
		router.guard.RUnlock()

		if err, ok := res[0].Interface().(error); ok && err != nil {
			return err
		}
	} else {
		if remote, ok := router.remotes[name]; ok {
			router.guard.RUnlock()
			err := remote.Call(context.Background(), f, args, reply)
			if err != nil {
				fmt.Printf("router.Call(%s, %s) err: %v\n", name, f, err)
			}
			return err
		} else {
			router.guard.RUnlock()
			router.guard.Lock()
			router.remotes[name] = InitZookeeperRpcClient(name, router.zkServers)
			router.guard.Unlock()
			err := router.remotes[name].Call(context.Background(), f, args, reply)
			if err != nil {
				fmt.Printf("router.Call(%s, %s) err: %v\n", name, f, err)
			}
			return err
		}
	}
	return nil
}
