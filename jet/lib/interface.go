package jetlib

type JetListener interface {
	Notify(name string, data interface{})
}
