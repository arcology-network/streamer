package broker

type DefaultProducer struct {
	name          string
	outputs       []string
	bufferLengths []int
}

func NewDefaultProducer(name string, outputs []string, bufferLengths []int) Producer {
	return &DefaultProducer{
		name:          name,
		outputs:       outputs,
		bufferLengths: bufferLengths,
	}
}

func (p *DefaultProducer) Name() string {
	return p.name
}

func (p *DefaultProducer) Outputs() []string {
	return p.outputs
}

func (p *DefaultProducer) BufferLengths() []int {
	return p.bufferLengths
}
