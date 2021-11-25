package mesh

type SeldonConfig struct {
	Name       string `yaml:"name"`
	SeldonSpec `yaml:"spec"`
}

type SeldonSpec struct {
	Servers []StaticServer `yaml:"servers"`
	Models  []StaticModel  `yaml:"models"`
}

type StaticServer struct {
	Name     string    `yaml:"name"`
	Replicas []Replica `yaml:"replicas"`
}

type Replica struct {
	Address string `yaml:"address"`
	Port    uint32 `yaml:"port"`
}

type StaticModel struct {
	Name        string `yaml:"name"`
	ModelServer string `yaml:"modelServer"`
	Servers     []int  `yaml:"servers"`
}
