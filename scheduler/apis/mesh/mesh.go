package mesh

import (
	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
)

type Mapping struct {
	Servers map[string]*ServerAssignment
	Models map[string]*ModelAssignment
}

type ServerAssignment struct {
	Server *pb.ServerDetails
	LoadedModels map[string]bool
	Resources []Resources
}

type Resources struct {
	UsedActiveMemory int32
	UsedInactiveMemory int32
}

type ModelAssignment struct {
	Model *pb.ModelDetails
	Server string
	Assignment []int
}

func NewMapping() Mapping {
	m := Mapping{}
	m.Servers = make(map[string]*ServerAssignment)
	m.Models = make(map[string]*ModelAssignment)
	return m
}
