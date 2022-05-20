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
	ExperimentName    string
	UpdatedExperiment bool
	Status            *ExperimentEventStatus
	KubernetesMeta    *KubernetesMeta
}

type ExperimentEventStatus struct {
	Active            bool
	CandidatesReady   bool
	MirrorReady       bool
	StatusDescription string
}

type KubernetesMeta struct {
	Namespace  string
	Generation int64
}

func (e ExperimentEventMsg) String() string {
	return e.ExperimentName
}

type PipelineEventMsg struct {
	PipelineName    string
	PipelineVersion uint32
	UID             string
}

func (p PipelineEventMsg) String() string {
	return fmt.Sprintf("%s:%d (%s)", p.PipelineName, p.PipelineVersion, p.UID)
}
