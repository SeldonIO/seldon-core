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
	PipelineCreate
	PipelineCreating
	PipelineReady
	PipelineFailed
	PipelineTerminate
	PipelineTerminating
	PipelineTerminated
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
