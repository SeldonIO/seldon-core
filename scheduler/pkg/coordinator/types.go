/*
Copyright 2022 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
	PipelineName      string
	PipelineVersion   uint32
	UID               string
	ExperimentUpdate  bool
	ModelStatusChange bool
}

func (p PipelineEventMsg) String() string {
	return fmt.Sprintf("%s:%d (%s)", p.PipelineName, p.PipelineVersion, p.UID)
}
