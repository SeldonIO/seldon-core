package experiment

type Experiment struct {
	Name              string
	Active            bool
	Deleted           bool
	Baseline          *Candidate
	Candidates        []*Candidate
	Mirror            *Mirror
	Config            *Config
	StatusDescription string
}

type Candidate struct {
	ModelName string
	Weight    uint32
}

type Mirror struct {
	ModelName string
	Percent   int32
}

type Config struct {
	StickySessions bool
}
