package interfaces

type DependencyServiceInterface interface {
	SetState(state interface{})
	Start() error
	Ready() bool
	Stop() error
	Name() string
}
