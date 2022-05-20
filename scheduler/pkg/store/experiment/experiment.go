package experiment

type Experiment struct {
	Name              string
	Active            bool
	Deleted           bool
	DefaultModel      *string
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
	} else {
		return true
	}
}

type KubernetesMeta struct {
	Namespace  string
	Generation int64
}

type Candidate struct {
	ModelName string
	Weight    uint32
	Ready     bool
}

type Mirror struct {
	ModelName string
	Percent   uint32
	Ready     bool
}

type Config struct {
	StickySessions bool
}
