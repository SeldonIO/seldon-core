package coordinator

import "fmt"

type ModelEventMsg struct {
	ModelName    string
	ModelVersion uint32
}

func (m ModelEventMsg) String() string {
	return fmt.Sprintf("%s:%d", m.ModelName, m.ModelVersion)
}

type ExperimentEventMsg struct {
	ExperimentName string
	Status         *ExperimentEventStatus
}

type ExperimentEventStatus struct {
	Active            bool
	StatusDescription string
}

func (e ExperimentEventMsg) String() string {
	return e.ExperimentName
}
