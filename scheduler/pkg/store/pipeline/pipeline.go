/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package pipeline

import (
	"fmt"
	"time"
)

type Pipeline struct {
	Name        string
	LastVersion uint32
	Versions    []*PipelineVersion
	Deleted     bool
}

func (p *Pipeline) GetPipelineVersion(versionNumber uint32) *PipelineVersion {
	for _, pv := range p.Versions {
		if pv.Version == versionNumber {
			return pv
		}
	}
	return nil
}

func (p *Pipeline) GetLatestPipelineVersion() *PipelineVersion {
	if len(p.Versions) > 0 {
		return p.Versions[len(p.Versions)-1]
	}
	return nil
}

func (p *Pipeline) GetPreviousPipelineVersion() *PipelineVersion {
	if len(p.Versions) > 1 {
		return p.Versions[len(p.Versions)-2]
	}
	return nil
}

type PipelineVersion struct {
	Name           string
	Version        uint32
	UID            string
	Input          *PipelineInput
	Steps          map[string]*PipelineStep
	State          *PipelineState
	Output         *PipelineOutput
	KubernetesMeta *KubernetesMeta
}

func (pv *PipelineVersion) String() string {
	return fmt.Sprintf("%s:%d (%s)", pv.Name, pv.Version, pv.UID)
}

type KubernetesMeta struct {
	Namespace  string
	Generation int64
}

type PipelineStatus uint32

const (
	PipelineStatusUnknown PipelineStatus = iota
	PipelineCreate                       // Received signal to create pipeline.
	PipelineCreating                     // In the process of creating pipeline.
	PipelineReady                        // Pipeline is ready to be used.
	PipelineFailed                       // Pipeline creation/deletion failed.
	PipelineTerminate                    // Received signal that pipeline should be terminated.
	PipelineTerminating                  // In the process of doing cleanup/housekeeping for pipeline termination.
	PipelineTerminated                   // Pipeline has been terminated.
)

type PipelineState struct {
	Status      PipelineStatus
	ModelsReady bool
	Reason      string
	Timestamp   time.Time
}

func (ps PipelineStatus) String() string {
	return [...]string{"PipelineStatusUnknown", "PipelineCreate", "PipelineCreating", "PipelineReady", "PipelineFailed", "PipelineTerminate", "PipelineTerminating", "PipelineTerminated"}[ps]
}

func (ps *PipelineState) setState(status PipelineStatus, reason string) {
	ps.Status = status
	ps.Reason = reason
	ps.Timestamp = time.Now()
}

type JoinType uint32

const (
	JoinInner = iota
	JoinOuter
	JoinAny
)

type PipelineStep struct {
	Name             string
	Inputs           []string
	Triggers         []string
	TensorMap        map[string]string
	JoinWindowMs     *uint32
	InputsJoinType   JoinType
	TriggersJoinType JoinType
	Batch            *Batch
	Available        bool
}

type Batch struct {
	Size     *uint32
	WindowMs *uint32
}

type PipelineOutput struct {
	Steps         []string
	JoinWindowMs  uint32
	StepsJoinType JoinType
	TensorMap     map[string]string
}

type PipelineInput struct {
	ExternalInputs   []string
	ExternalTriggers []string
	TensorMap        map[string]string
	JoinWindowMs     *uint32
	InputsJoinType   JoinType
	TriggersJoinType JoinType
}
