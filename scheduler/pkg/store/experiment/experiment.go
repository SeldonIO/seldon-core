/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package experiment

import "time"

type ResourceType uint32

const (
	ModelResourceType ResourceType = iota
	PipelineResourceType
)

type Experiment struct {
	Name              string
	Active            bool
	Deleted           bool
	DeletedAt         time.Time
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
