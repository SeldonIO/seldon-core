package mesh

type SeldonConfig struct {
	Name string `yaml:"name"`
	SeldonSpec `yaml:"spec"`
}

type SeldonSpec struct {
	Servers []Server `yaml:"servers"`
	Models  []Model  `yaml:"models"`
}

type Server struct {
	Name string        `yaml:"name"`
	Replicas []Replica `yaml:"replicas"`
}

type Replica struct {
	Address string  `yaml:"address"`
	Port    uint32  `yaml:"port"`
}

type Model struct {
	Name string `yaml:"name"`
	ModelServer string `yaml:"modelServer"`
	Servers []int `yaml:"servers"`
}
