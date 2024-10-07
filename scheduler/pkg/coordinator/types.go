/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package coordinator

import "fmt"

type ModelEventUpdateContext int
type ServerEventUpdateContext int

const (
	SERVER_STATUS_UPDATE ServerEventUpdateContext = iota
	SERVER_REPLICA_CONNECTED
)

const (
	SERVER_STATUS_UPDATE ServerEventUpdateContext = iota
	SERVER_REPLICA_CONNECTED
)

const (
	SERVER_STATUS_UPDATE ServerEventUpdateContext = iota
	SERVER_REPLICA_CONNECTED
)

type ModelEventMsg struct {
	ModelName    string
	ModelVersion uint32
}

func (m ModelEventMsg) String() string {
	return fmt.Sprintf("%s:%d", m.ModelName, m.ModelVersion)
}

type ServerEventMsg struct {
	ServerName    string
	ServerIdx     uint32
	Source        string
	UpdateContext ServerEventUpdateContext
}

func (m ServerEventMsg) String() string {
	return fmt.Sprintf("%s:%d", m.ServerName, m.ServerIdx)
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
	Source            string
}

func (p PipelineEventMsg) String() string {
	return fmt.Sprintf("%s:%d (%s)", p.PipelineName, p.PipelineVersion, p.UID)
}
