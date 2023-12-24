package actor

import "encoding/gob"

var Factory WorkerFactory

func init() {
	Factory.registry = make(map[string]WorkerCreator)
	Factory.Register("cleaner", NewCleaner)
	gob.Register(&BlockStart{})
}
