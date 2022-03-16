package pipeline

import (
	"fmt"
	"time"
)

type Pipeline struct {
	Name        string
	LastVersion uint32
	Versions    []*PipelineVersion
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
	Name    string
	Version uint32
	UID     string
	Steps   map[string]*PipelineStep
	State   *PipelineState
	Output  *PipelineOutput
}

func (pv *PipelineVersion) String() string {
	return fmt.Sprintf("%s:%d (%s)", pv.Name, pv.Version, pv.UID)
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
	Status    PipelineStatus
	Reason    string
	Timestamp time.Time
}

func (ps PipelineStatus) String() string {
	return [...]string{"PipelineStatusUnknown", "PipelineCreate", "PipelineCreating", "PipelineReady", "PipelineFailed", "PipelineTerminate", "PipelineTerminating", "PipelineTerminated"}[ps]
}

func (ps *PipelineState) setState(status PipelineStatus, reason string) {
	ps.Status = status
	ps.Reason = reason
	ps.Timestamp = time.Now()
}

type PipelineStep struct {
	Name             string
	Inputs           []string
	JoinWindowMs     *uint32
	PassEmptyOutputs bool
}

type PipelineOutput struct {
	Inputs       []string
	JoinWindowMs uint32
}
