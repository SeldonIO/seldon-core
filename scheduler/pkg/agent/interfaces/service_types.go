package interfaces

type SubServiceType uint8

const (
	UnknownSubServiceType SubServiceType = iota
	CriticalControlPlaneService
	AuxControlPlaneService
	CriticalDataPlaneService
	AuxDataPlaneService
	OptionalService
)
