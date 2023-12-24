package sequencer

type HeightFilter struct {
	height    uint64
	whitelist []string
	signalMsg string
}

func NewHeightFilter(names []string, signalMsg string, height uint64) *HeightFilter {
	filter := &HeightFilter{
		height:    height,
		signalMsg: signalMsg,
		whitelist: names,
	}

	return filter
}

func (this *HeightFilter) Less(lhs interface{}, rhs interface{}) bool {
	return lhs.(MessageInterface).Height() < rhs.(MessageInterface).Height()
}

func (this *HeightFilter) Check(msg interface{}) bool {
	if len(this.whitelist) == 0 {
		return true
	}

	if this.height > msg.(MessageInterface).Height() ||
		(msg.(MessageInterface).Name() == this.signalMsg && (this.height == msg.(MessageInterface).Height() || this.height+1 < msg.(MessageInterface).Height())) {
		panic("Error: Wrong height !!!")
	}

	for _, v := range this.whitelist {
		if msg.(MessageInterface).Name() == v {
			return true
		}
	}
	return this.signalMsg == msg.(MessageInterface).Name()
}

func (this *HeightFilter) Filter(msg interface{}) (interface{}, bool, bool) {
	if msg.(MessageInterface).Name() == this.signalMsg {
		if this.height+1 == msg.(MessageInterface).Height() {
			this.height = msg.(MessageInterface).Height()
			return msg, true, true // Change the internal state and forward message as downstream entity may need it
		}
		panic("Error: Duplicate signal message !!!") // One signal for a height only
	}
	return msg, this.height == msg.(MessageInterface).Height(), false
}
