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
