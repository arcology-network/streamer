package broker

type DefaultConsumer struct {
	name       string
	inputs     []string
	controller StreamController
}

func NewDefaultConsumer(name string, inputs []string, controller StreamController) Consumer {
	return &DefaultConsumer{
		name:       name,
		inputs:     inputs,
		controller: controller,
	}
}

func (c *DefaultConsumer) Name() string {
	return c.name
}

func (c *DefaultConsumer) Inputs() []string {
	return c.inputs
}

func (c *DefaultConsumer) StreamController() StreamController {
	return c.controller
}
