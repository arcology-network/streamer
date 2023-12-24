package actor

type Initializer interface {
	InitMsgs() []*Message
}
