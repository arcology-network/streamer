package broker

import (
	"reflect"
	"runtime"
	"sync"

	log "github.com/sirupsen/logrus"
)

type ActorInterface interface {
	Consume(data interface{})
}

type StreamController interface {
	GetListener(name string) streamListener
	Notify(name string, data interface{})
	Serve()
	GenerateDotLabel() string
	GetActor() ActorInterface
}

type Conjunctions struct {
	triggers       map[string]chan interface{}
	indices        map[string]int
	values         []interface{}
	ActorInterface ActorInterface
}

func NewConjunctions(actor ActorInterface) StreamController {
	return &Conjunctions{
		triggers:       make(map[string]chan interface{}),
		indices:        make(map[string]int),
		ActorInterface: actor,
	}
}

func (conj *Conjunctions) GetListener(name string) streamListener {
	conj.indices[name] = len(conj.indices)
	conj.triggers[name] = make(chan interface{})
	return &defaultListener{
		name:       name,
		controller: conj,
	}
}

func (conj *Conjunctions) Notify(name string, data interface{}) {
	conj.triggers[name] <- data
}

func (conj *Conjunctions) Serve() {
	conj.values = make([]interface{}, len(conj.triggers))
	for {
		var wg sync.WaitGroup
		for name := range conj.triggers {
			tempName := name
			wg.Add(1)
			go func(name string, c chan interface{}) {
				data := <-c
				log.WithFields(log.Fields{
					"channel": name,
				}).Debug("got trigger")
				conj.values[conj.indices[name]] = data
				wg.Done()
			}(tempName, conj.triggers[tempName])
		}
		wg.Wait()
		log.Debug("all the prerequests met")

		valueCopy := make([]interface{}, len(conj.values))
		copy(valueCopy, conj.values)
		conj.ActorInterface.Consume(valueCopy)
	}
}

func (conj *Conjunctions) GenerateDotLabel() string {
	var inputs []string
	var label string
	for name := range conj.indices {
		inputs = append(inputs, name)
	}

	for i := 0; i < len(inputs)-1; i++ {
		label += inputs[i] + " AND "
	}
	label += inputs[len(inputs)-1]
	return label
}

func (conj *Conjunctions) GetActor() ActorInterface {
	return conj.ActorInterface
}

type Aggregated struct {
	Name string
	Data interface{}
}

type Disjunctions struct {
	chans          map[string]chan interface{}
	aggrChan       chan interface{}
	ActorInterface ActorInterface
}

func NewDisjunctions(actor ActorInterface, bufSize int) *Disjunctions {
	return &Disjunctions{
		chans:          make(map[string]chan interface{}),
		aggrChan:       make(chan interface{}, bufSize),
		ActorInterface: actor,
	}
}

func (dis *Disjunctions) GetListener(name string) streamListener {
	dis.chans[name] = make(chan interface{})
	return &defaultListener{
		name:       name,
		controller: dis,
	}
}

func (dis *Disjunctions) Notify(name string, data interface{}) {
	dis.chans[name] <- data
}

func (dis *Disjunctions) Serve() {
	for name := range dis.chans {
		tempName := name
		go func(name string, c chan interface{}) {
			for {
				data := <-c
				dis.aggrChan <- Aggregated{
					Name: name,
					Data: data,
				}
			}
		}(tempName, dis.chans[tempName])
	}

	for aggr := range dis.aggrChan {
		log.WithFields(log.Fields{
			"channel": aggr.(Aggregated).Name,
		}).Debug("new item")
		dis.ActorInterface.Consume(aggr)
	}
}

func (dis *Disjunctions) GenerateDotLabel() string {
	var inputs []string
	var label string
	for name := range dis.chans {
		inputs = append(inputs, name)
	}

	for i := 0; i < len(inputs)-1; i++ {
		label += inputs[i] + " OR "
	}
	label += inputs[len(inputs)-1]
	return label
}

func (dis *Disjunctions) GetActor() ActorInterface {
	return dis.ActorInterface
}

type CustomizableController struct {
	inputIndices   map[string]int
	inputValues    []interface{}
	chans          map[string]chan interface{}
	pf             PassableFunc
	ActorInterface ActorInterface
}

type PassableFunc func(inputs interface{}) bool

func NewCustomizableController(actor ActorInterface, pf PassableFunc) *CustomizableController {
	return &CustomizableController{
		inputIndices:   make(map[string]int),
		chans:          make(map[string]chan interface{}),
		pf:             pf,
		ActorInterface: actor,
	}
}

func (cc *CustomizableController) GetListener(name string) streamListener {
	currIndex := len(cc.inputIndices)
	cc.inputIndices[name] = currIndex
	cc.chans[name] = make(chan interface{})
	return &defaultListener{
		name:       name,
		controller: cc,
	}
}

func (cc *CustomizableController) Notify(name string, data interface{}) {
	cc.chans[name] <- data
}

func (cc *CustomizableController) Serve() {
	cc.inputValues = make([]interface{}, len(cc.inputIndices))

	cases := createCases(cc.chans, cc.inputIndices)
	activeChan := len(cc.chans)
	for {
		chosen, data, _ := reflect.Select(cases)
		cc.inputValues[chosen] = data.Interface()
		if cc.pf(cc.inputValues) {
			log.Debug("passable combination")
			cases = createCases(cc.chans, cc.inputIndices)
			activeChan = len(cc.chans)

			valueCopy := make([]interface{}, len(cc.inputValues))
			copy(valueCopy, cc.inputValues)
			cc.ActorInterface.Consume(valueCopy)
		} else {
			log.WithFields(log.Fields{
				"channel": chosen,
			}).Debug("check failed, close channel")
			cases[chosen].Chan = reflect.ValueOf(nil)
			activeChan--
			if activeChan == 0 {
				log.Warn("no more active channel")
				cases = createCases(cc.chans, cc.inputIndices)
				activeChan = len(cc.chans)
			}
		}
	}
}

func (cc *CustomizableController) GenerateDotLabel() string {
	inputs := make([]string, len(cc.inputIndices))
	label := runtime.FuncForPC(reflect.ValueOf(cc.pf).Pointer()).Name() + "("
	for name, index := range cc.inputIndices {
		inputs[index] = name
	}

	for i := 0; i < len(inputs)-1; i++ {
		label += inputs[i] + ", "
	}
	label += inputs[len(inputs)-1] + ")"
	return label
}

func (cc *CustomizableController) GetActor() ActorInterface {
	return cc.ActorInterface
}

type ShortCircuitActor struct {
	in       chan interface{}
	consumer Consumer
	name     string
}

func NewShortCircuitActor(consumer Consumer, name string) *ShortCircuitActor {
	return &ShortCircuitActor{
		in:       make(chan interface{}),
		consumer: consumer,
		name:     name,
	}
}

func (scw *ShortCircuitActor) Consume(data interface{}) {
	scw.in <- data
}

func (scw *ShortCircuitActor) Serve() {
	for i := range scw.in {
		scw.consumer.StreamController().Notify(scw.name, Aggregated{Name: scw.name, Data: i})
	}
}

func createCases(chans map[string]chan interface{}, indices map[string]int) []reflect.SelectCase {
	cases := make([]reflect.SelectCase, len(chans))
	for name, c := range chans {
		cases[indices[name]] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(c)}
	}
	return cases
}
