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

type InterfaceCreator func(concurrency int, groupId string) interface{}

type Configurable interface {
	Config(params map[string]interface{})
}

type InterfaceFactory struct {
	registry map[string]InterfaceCreator
}

func (factory *InterfaceFactory) Register(name string, creator InterfaceCreator) {
	factory.registry[name] = creator
}

func (factory *InterfaceFactory) Create(name string, concurrency int, groupId string, params map[string]interface{}) interface{} {
	if creator, ok := factory.registry[name]; !ok {
		panic("interface name not found: " + name)
	} else {
		i := creator(concurrency, groupId)
		if len(params) == 0 {
			return i
		}

		if configurable, ok := i.(Configurable); !ok {
			panic("interface not configurable: " + name)
		} else {
			configurable.Config(params)
			return i
		}
	}
}
