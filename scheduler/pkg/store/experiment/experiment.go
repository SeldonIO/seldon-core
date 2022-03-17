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

type KubernetesMeta struct {
	Namespace  string
	Generation int64
}

type Candidate struct {
	ModelName string
	Weight    uint32
}

type Mirror struct {
	ModelName string
	Percent   uint32
}

type Config struct {
	StickySessions bool
}
