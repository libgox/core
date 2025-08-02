package types

type SourceType int

const (
	Polling SourceType = iota
	Dynamic
)

type Source interface {
	Start() error
	Stop() error
	Type() SourceType
	Read() ([]byte, error)
	SetUpdateCallback(func([]byte))
}

type Listener[T any] interface {
	Update(config T)
}
