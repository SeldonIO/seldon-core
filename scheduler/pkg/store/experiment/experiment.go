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

package experiment

type ResourceType uint32

const (
	ModelResourceType ResourceType = iota
	PipelineResourceType
)

type Experiment struct {
	Name              string
	Active            bool
	Deleted           bool
	Default           *string
	ResourceType      ResourceType
	Candidates        []*Candidate
	Mirror            *Mirror
	Config            *Config
	StatusDescription string
	KubernetesMeta    *KubernetesMeta
}

func (e *Experiment) AreCandidatesReady() bool {
	for _, candidate := range e.Candidates {
		if !candidate.Ready {
			return false
		}
	}
	return true
}

func (e *Experiment) IsMirrorReady() bool {
	if e.Mirror != nil {
		return e.Mirror.Ready
	}
	return true
}

type KubernetesMeta struct {
	Namespace  string
	Generation int64
}

type Candidate struct {
	Name   string
	Weight uint32
	Ready  bool
}

type Mirror struct {
	Name    string
	Percent uint32
	Ready   bool
}

type Config struct {
	StickySessions bool
}
