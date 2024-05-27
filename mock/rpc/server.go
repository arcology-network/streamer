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

package rpc

import (
	"context"
	"fmt"
	"reflect"
	"testing"
)

var service interface{}

func InitRPCServer(serviceAddr, basepath string, zkAddrs []string, rcvrs, fns []interface{}) {
	t.Log("InitRPCServer")
	service = rcvrs[0]
}

func Call(f string, args, reply interface{}) {
	method := reflect.ValueOf(service).MethodByName(f)
	if method.Type().NumIn() != 3 {
		panic(fmt.Sprintf("expected 3 arguments, got %d", method.Type().NumIn()))
	}

	method.Call([]reflect.Value{reflect.ValueOf(context.Background()), reflect.ValueOf(args), reflect.ValueOf(reply)})
}

var t testing.TB

func NewRPCServerInitializer(logger testing.TB) func(string, string, []string, []interface{}, []interface{}) {
	t = logger
	return InitRPCServer
}
