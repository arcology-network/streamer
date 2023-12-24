package sequencer

type FilterInterface interface {
	Check(interface{}) bool
	Less(interface{}, interface{}) bool
	Filter(interface{}) (interface{}, bool, bool)
}

type MessageInterface interface {
	Height() uint64
	Name() string
}
