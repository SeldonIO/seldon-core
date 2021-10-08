package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/seldonio/seldon-core/scheduler/apis/mesh"
	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/processor"
	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"math/rand"
	"net"
	"sort"

)

var (
	ErrAddServerEmptyServerName = status.Errorf(codes.FailedPrecondition, "Empty server name passed")
	ErrRemoveServerServerNotFound = status.Errorf(codes.FailedPrecondition,"Server name not found")
	ErrLoadModelServerNameNotFound = status.Errorf(codes.InvalidArgument,"Server name does not exist")
	ErrLoadModelRequirementFailed = status.Errorf(codes.FailedPrecondition, "Compatible server was not found")
	ErrLoadModelUnableToFindServerInstance = status.Errorf(codes.Internal, "Unable to find enough server instances for model")
	ErrModelStatusModelNotFound = status.Errorf(codes.FailedPrecondition,"Model not found")
	ErrServerStatusServerNotFound = status.Errorf(codes.FailedPrecondition,"Server name not found")
)

type SchedulerServer struct {
	pb.UnimplementedSchedulerServer
	Mesh mesh.Mapping
	EnvoyProcessor *processor.IncrementalProcessor
	logger log.FieldLogger
}

func(s SchedulerServer) StartGrpcServer(schedulerPort uint) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", schedulerPort))
	if err != nil {
		log.Fatalf("failed to create listener: %v", err)
	}
	opts := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterSchedulerServer(grpcServer, s)
	s.logger.Printf("Scheduler server running on %d", schedulerPort)
	return  grpcServer.Serve(lis)
}

func NewScheduler(cache cache.SnapshotCache, nodeID string, logger log.FieldLogger) *SchedulerServer {

	s := &SchedulerServer{
		Mesh: mesh.NewMapping(),
		EnvoyProcessor: processor.NewIncrementalProcessor(cache, nodeID, logger),
		logger: logger,
	}
	s.EnvoyProcessor.SetListener("seldon_http")
	return s
}

func (s SchedulerServer) AddServer(ctx context.Context, details *pb.ServerDetails) (*pb.AddServerResponse, error) {
	if details.Name == "" {
		return nil, ErrAddServerEmptyServerName
	}
	s.Mesh.Servers[details.Name] = &mesh.ServerAssignment{
		Server: details,
		LoadedModels: make(map[string]bool),
	}
	return &pb.AddServerResponse{}, nil
}

func (s SchedulerServer) RemoveServer(ctx context.Context, reference *pb.ServerReference) (*pb.RemoveServerResponse, error) {
	_, ok := s.Mesh.Servers[reference.Name]
	if !ok {
		return nil, ErrRemoveServerServerNotFound
	}
	delete(s.Mesh.Servers, reference.Name)
	return &pb.RemoveServerResponse{}, nil
}

func (s SchedulerServer) LoadModel(ctx context.Context, details *pb.ModelDetails) (*pb.LoadModelResponse, error) {
	// find model assignment
	modelAssignment, ok := s.Mesh.Models[details.Name]
	var serverAssignment *mesh.ServerAssignment
	if !ok { // no mapping so find one
		if details.Server != nil {
			serverAssignment, ok = s.Mesh.Servers[*details.Server]
			if !ok {
				return nil, ErrLoadModelServerNameNotFound
			}
			modelAssignment = &mesh.ModelAssignment{
				Model: details,
				Server: serverAssignment.Server.Name,
			}
		} else { // need to schedule to a server
			serverAssignment = s.findServerForModel(details)
			if serverAssignment == nil {
				// error a compatible server was not found
				return nil, ErrLoadModelRequirementFailed
			}
			modelAssignment = &mesh.ModelAssignment{
				Model: details,
				Server: serverAssignment.Server.Name,
			}
			s.Mesh.Models[details.Name] = modelAssignment
		}
	} else {
		serverAssignment = s.Mesh.Servers[modelAssignment.Server]
	}
	serverAssignment.LoadedModels[modelAssignment.Model.Name] = true
	//assign model to server replicas
	if len(modelAssignment.Assignment) == int(details.Replicas) {
		// Do nothing
	} else if int(details.Replicas) < len(modelAssignment.Assignment) {
		// Scale down needed
		numReplicasToRemove := len(modelAssignment.Assignment) - int(details.Replicas)
		modelAssignment.Assignment = modelAssignment.Assignment[:len(modelAssignment.Assignment)-numReplicasToRemove]
	} else {
		// Scale up needed
		numReplicasToAdd := int(details.Replicas) - len(modelAssignment.Assignment)
		for i := 0; i < numReplicasToAdd; i++ {
			serverIdx, err := getNextServerInstance(serverAssignment, modelAssignment)
			if err != nil {
				return nil, ErrLoadModelUnableToFindServerInstance
			}
			modelAssignment.Assignment = append(modelAssignment.Assignment, serverIdx)
			sort.Ints(modelAssignment.Assignment)
		}
	}

	// Update Envoy
	err := s.EnvoyProcessor.SetModelForServerInEnvoy(modelAssignment, serverAssignment)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &pb.LoadModelResponse{}, nil
}

func printMesh(mesh mesh.Mapping) error {
	b, err := json.Marshal(mesh)
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}

func getNextServerInstance(server *mesh.ServerAssignment, model *mesh.ModelAssignment) (int, error) {
	randomServerOrdering := rand.Perm(len(server.Server.Replicas))
	for _,possibleServerIdx := range randomServerOrdering {
		found := false
		for _, existingServerIdx := range model.Assignment {
			if possibleServerIdx == existingServerIdx {
				found = true
				break
			}
		}
		if !found {
			return possibleServerIdx, nil
		}
	}
	return 0, fmt.Errorf("Failed to find available server instance for model")
}

func (s SchedulerServer) findServerForModel(details *pb.ModelDetails) *mesh.ServerAssignment {
	for _, v := range s.Mesh.Servers {
		requirementsFound := true
		for _,requirement := range details.Requirements {
			requirementFound := false
			for _,capability := range v.Server.Capabilities {
				if requirement == capability {
					if int(details.Replicas) <= len(v.Server.Replicas) {
						requirementFound = true
						break
					}
				}
			}
			if !requirementFound {
				requirementsFound = false
				break
			}
		}
		if requirementsFound {
			return v
		}
	}
	return nil
}

func (s SchedulerServer) UnloadModel(ctx context.Context, reference *pb.ModelReference) (*pb.UnloadModelResponse, error) {
	delete(s.Mesh.Models, reference.Name)
	return &pb.UnloadModelResponse{}, nil
}

func (s SchedulerServer) ModelStatus(ctx context.Context, reference *pb.ModelReference) (*pb.ModelStatusResponse, error) {
	modelAssignment, ok := s.Mesh.Models[reference.Name]
	if !ok {
		return nil, ErrModelStatusModelNotFound
	}
	assignment := make([]int32,len(modelAssignment.Assignment))
	for i:=0; i< len(modelAssignment.Assignment); i++ {
		assignment[i] = int32(modelAssignment.Assignment[i])
	}
	return &pb.ModelStatusResponse{
		ModelName: modelAssignment.Model.Name,
		ServerName: modelAssignment.Server,
		Assignment: assignment,
	}, nil
}

func (s SchedulerServer) ServerStatus(ctx context.Context, reference *pb.ServerReference) (*pb.ServerStatusResponse, error) {
	serverAssignment, ok := s.Mesh.Servers[reference.Name]
	if !ok {
		return nil, ErrServerStatusServerNotFound
	}
	var loadedModels []string
	for k,_ := range serverAssignment.LoadedModels {
		loadedModels = append(loadedModels,k)
	}
	return &pb.ServerStatusResponse{
		ServerName: reference.Name,
		LoadedModels: loadedModels,
	}, nil
}