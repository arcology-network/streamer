package actor

import "fmt"

type WorkerCreator func(concurrency int, groupId string) IWorkerEx

type Configurable interface {
	Config(params map[string]interface{})
}

type WorkerFactory struct {
	registry map[string]WorkerCreator
}

func (factory *WorkerFactory) Register(name string, creator WorkerCreator) {
	factory.registry[name] = creator
}

func (factory *WorkerFactory) Create(name string, concurrency int, groupId string, params map[string]interface{}) IWorkerEx {
	if creator, ok := factory.registry[name]; !ok {
		panic("worker name not found: " + name)
	} else {
		worker := creator(concurrency, groupId)
		if len(params) == 0 {
			return worker
		}

		if configurable, ok := worker.(Configurable); !ok {
			panic("worker type not configurable: " + name)
		} else {
			configurable.Config(params)
			return worker
		}
	}
}

func (factory *WorkerFactory) Registry() map[string]WorkerCreator {
	return factory.registry
}

func (factory *WorkerFactory) Print() {
	for name, creator := range factory.registry {
		worker := creator(1, "printer")
		fmt.Printf("%v { ", name)
		if _, ok := worker.(FSMCompatible); ok {
			fmt.Print("FSMCompatible ")
		}
		if _, ok := worker.(HeightSensitive); ok {
			fmt.Print("HeightSensitive ")
		}
		if _, ok := worker.(Configurable); ok {
			fmt.Print("Configurable ")
		}
		if _, ok := worker.(Initializer); ok {
			fmt.Print("Initializer ")
		}
		fmt.Print("}\n")

		inputs, isConjunction := worker.Inputs()
		fmt.Print("\tIN")
		if isConjunction {
			fmt.Print("[AND] { ")
		} else {
			fmt.Print("[OR] { ")
		}
		for _, input := range inputs {
			fmt.Printf("%v ", input)
		}
		outputs := worker.Outputs()
		fmt.Print("}, OUT { ")
		for output := range outputs {
			fmt.Printf("%v ", output)
		}
		fmt.Print("}\n")
	}
}
