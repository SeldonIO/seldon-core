package interfaces

// TODO: use generics
// TODO: define more logic operations

type LogicOperation int

const (
	Gte LogicOperation = iota
)

type ModelStatsKV struct {
	Value     uint32
	Key       string
	ModelName string
}

type ModelScalingStats interface {
	Inc(string, uint32) error
	IncDefault(string) error
	Dec(string, uint32) error
	DecDefault(string) error
	Reset(string) error
	Set(string, uint32) error
	Get(string) (uint32, error)
	GetAll(uint32, LogicOperation, bool) ([]*ModelStatsKV, error)
	Info() string
	Delete(string) error
	Add(string) error
}
