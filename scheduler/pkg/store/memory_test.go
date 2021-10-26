package store

import (
	"errors"
	. "github.com/onsi/gomega"
	pba "github.com/seldonio/seldon-core/scheduler/apis/mlops/agent"
	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	log "github.com/sirupsen/logrus"
	"testing"
)

func createTestMemoryScheduler() *MemorySchedulerStore {
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	agentChan := make(chan string, 10)
	envoyChan := make(chan string, 10)
	m := NewMemoryScheduler(logger, agentChan, envoyChan)
	return m
}

func TestMatchingReplicas(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	type test struct {
		c1 []string
		servers map[string]*Server
		replicas uint32
		memory uint64
		result []string
	}
	smallMemory := uint64(100)
	largeMemory := uint64(10000)
	tests := []test {
		{c1: []string{"foo"}, // simple single match
			servers: map[string]*Server {"server1":{replicas: map[int]*ServerReplica{1:{capabilities: []string{"foo"}, memory: 1000, availableMemory: 1000}}}},
			replicas: 1,
			memory: smallMemory,
			result:[]string{"server1"}},
		{c1: []string{"foo","bar"}, // multiple requirements
			servers: map[string]*Server {"server1":{replicas: map[int]*ServerReplica{1:{capabilities: []string{"foo","foo2","bar"}, memory: 1000, availableMemory: 1000}}}},
			replicas: 1,
			memory: smallMemory,
			result:[]string{"server1"}},
		{c1: []string{"foo"}, // not enough memory for model
			servers: map[string]*Server {"server1":{replicas: map[int]*ServerReplica{1:{capabilities: []string{"foo"}, memory: 1000, availableMemory: 1000}}}},
			replicas: 1,
			memory: largeMemory,
			result:nil},
		{c1: []string{"foo"}, // server with more capabilities than requirements
			servers: map[string]*Server {"server1":{replicas: map[int]*ServerReplica{1:{capabilities: []string{"foo","bar"}, memory: 1000, availableMemory: 1000}}}},
			replicas: 1,
			memory: smallMemory,
			result:[]string{"server1"}},
		{c1: []string{"foo","bar"}, // more requirements than server
			servers: map[string]*Server {"server1":{replicas: map[int]*ServerReplica{1:{capabilities: []string{"foo"}, memory: 1000, availableMemory: 1000}}}},
			replicas: 1,
			memory: smallMemory,
			result:nil},
		{c1: []string{"foo"}, // multiple servers matching
			servers: map[string]*Server {"server1":{replicas: map[int]*ServerReplica{1:{capabilities: []string{"foo"}, memory: 1000, availableMemory: 1000}}},
				"server2":{replicas: map[int]*ServerReplica{1:{capabilities: []string{"foo"}, memory: 1000, availableMemory: 1000}}}},
			replicas: 1,
			memory: smallMemory,
			result:[]string{"server1","server2"}},
		{c1: []string{"foo"}, // replicas not enough
			servers: map[string]*Server {"server1":{replicas: map[int]*ServerReplica{1:{capabilities: []string{"foo"}, memory: 1000, availableMemory: 1000}}},
				"server2":{replicas: map[int]*ServerReplica{1:{capabilities: []string{"foo"}, memory: 1000, availableMemory: 1000}}}},
			replicas: 2,
			memory: smallMemory,
			result:nil},
		{c1: []string{"foo"}, // replicas enough for some servers
			servers: map[string]*Server {"server1":{replicas: map[int]*ServerReplica{
				1:{capabilities: []string{"foo"}, memory: 1000, availableMemory: 1000},
				2:{capabilities: []string{"foo"}, memory: 1000, availableMemory: 1000}}},
				"server2":{replicas: map[int]*ServerReplica{1:{capabilities: []string{"foo"}, memory: 1000}}}},
			replicas: 2,
			memory: smallMemory,
			result:[]string{"server1"}},
	}

	for tidx,test := range tests {
		t.Logf("test %d",tidx)
		res := getMatchingServers(test.servers, test.c1, test.replicas, test.memory)
		g.Expect(res).To(Equal(test.result))
	}
}


func TestUpdateServerReplica(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	m := createTestMemoryScheduler()
	type test struct {
		req *pba.AgentSubscribeRequest
		err error
	}
	tests := []test {
		{req: &pba.AgentSubscribeRequest{ServerName: "server1", ReplicaIdx: 0,
			ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"sklearn"}}},
			err: nil},
		{req: &pba.AgentSubscribeRequest{ServerName: "server1", ReplicaIdx: 0,
			ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"sklearn"}}},
			err: ServerReplicaAlreadyExistsErr}, // Replica exists error
	}
	for _,test := range tests {
		err := m.UpdateServerReplica(test.req)
		if test.err == nil {
			g.Expect(err).To(BeNil())
		} else {
			g.Expect(errors.Is(err,test.err)).To(BeTrue())
		}
	}
}

func TestGetServer(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	m := createTestMemoryScheduler()
	req := &pba.AgentSubscribeRequest{ServerName: "server1", ReplicaIdx: 0,
			ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"sklearn"}}}
	err := m.UpdateServerReplica(req)
	g.Expect(err).To(BeNil())
	s,err := m.GetServer("server1")
	g.Expect(err).To(BeNil())
	g.Expect(s).ToNot(BeNil())
	g.Expect(s.name).To(Equal("server1"))
	s,err = m.GetServer("foo")
	g.Expect(err).ToNot(BeNil())
	g.Expect(s).To(BeNil())
	g.Expect(errors.Is(err, ServerNotFoundErr)).To(BeTrue())
}


func TestGetServerReplica(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	m := createTestMemoryScheduler()
	req := &pba.AgentSubscribeRequest{ServerName: "server1", ReplicaIdx: 0,
		ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"sklearn"}}}
	err := m.UpdateServerReplica(req)
	g.Expect(err).To(BeNil())
	s,err := m.GetServerReplica("server1", 0)
	g.Expect(err).To(BeNil())
	g.Expect(s).ToNot(BeNil())
	g.Expect(s.server.Key()).To(Equal("server1"))
	s,err = m.GetServerReplica("foo", 0)
	g.Expect(err).ToNot(BeNil())
	g.Expect(s).To(BeNil())
	g.Expect(errors.Is(err, ServerReplicaNotFoundErr)).To(BeTrue())
	s,err = m.GetServerReplica("server1", 1)
	g.Expect(err).ToNot(BeNil())
	g.Expect(s).To(BeNil())
	g.Expect(errors.Is(err, ServerReplicaNotFoundErr)).To(BeTrue())
}

func TestCreateModel(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)

	type test struct {
		req []*pba.AgentSubscribeRequest
		model *pb.ModelDetails
		err error
	}
	smallMemory := uint64(100)
	largeMemory := uint64(10000)
	tests := []test {
		{req: []*pba.AgentSubscribeRequest{{ServerName: "server1", ReplicaIdx: 0,
			ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"sklearn"}}}},
			model: &pb.ModelDetails{Name: "model1", Uri: "gs://model", Requirements: []string{"sklearn"}, Memory:&smallMemory, Replicas: 1},
			err: nil},
		{req: []*pba.AgentSubscribeRequest{{ServerName: "server1", ReplicaIdx: 0,
			ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"sklearn"}}}},
			model: &pb.ModelDetails{Name: "model1", Uri: "gs://model", Requirements: []string{"sklearn"}, Memory:&smallMemory, Replicas: 2},
			err: FailedSchedulingErr}, // ask for too many replicas
		{req: []*pba.AgentSubscribeRequest{{ServerName: "server1", ReplicaIdx: 0,
			ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"sklearn"}}}},
			model: &pb.ModelDetails{Name: "model1", Uri: "gs://model", Requirements: []string{"sklearn"}, Memory:&largeMemory, Replicas: 1},
			err: FailedSchedulingErr}, // ask for too much memory
		{req: []*pba.AgentSubscribeRequest{{ServerName: "server1", ReplicaIdx: 0,
			ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"sklearn"}}}},
			model: &pb.ModelDetails{Name: "model1", Uri: "gs://model", Requirements: []string{"xgboost"}, Memory:&smallMemory, Replicas: 1},
			err: FailedSchedulingErr}, // unable to find requirements
		{req: []*pba.AgentSubscribeRequest{{ServerName: "server1", ReplicaIdx: 0,
			ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"sklearn","xgboost"}}}},
			model: &pb.ModelDetails{Name: "model1", Uri: "gs://model", Requirements: []string{"xgboost","sklearn"}, Memory:&smallMemory, Replicas: 1},
			err: nil}, // multiple requirements
		{req: []*pba.AgentSubscribeRequest{{ServerName: "server1", ReplicaIdx: 0,
			ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"sklearn"}}},
			{ServerName: "server1", ReplicaIdx: 1,
				ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"sklearn"}}}},
			model: &pb.ModelDetails{Name: "model1", Uri: "gs://model", Requirements: []string{"sklearn"}, Memory:&smallMemory, Replicas: 2},
			err: nil}, // schedule to 2 replicas
		{req: []*pba.AgentSubscribeRequest{{ServerName: "server1", ReplicaIdx: 0,
			ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"sklearn"}}},
			{ServerName: "server1", ReplicaIdx: 1,
				ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"foo"}}}},
			model: &pb.ModelDetails{Name: "model1", Uri: "gs://model", Requirements: []string{"sklearn"}, Memory:&smallMemory, Replicas: 2},
			err: FailedSchedulingErr}, // schedule to 2 replicas but 1 fails
	}
	for _,test := range tests {
		m := createTestMemoryScheduler()
		for _, repReq := range test.req {
			err := m.UpdateServerReplica(repReq) // Create server and replicas
			g.Expect(err).To(BeNil())
		}
		err := m.CreateModel(test.model.Name, test.model)
		g.Expect(err).To(BeNil())
		err = m.ScheduleModelToServer(test.model.Name)
		if test.err == nil {
			g.Expect(err).To(BeNil())
		} else {
			g.Expect(errors.Is(err,test.err)).To(BeTrue())
			g.Expect(m.isFailedModel(test.model.Name))
		}
	}
}

func TestCreateModelOnServer(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)

	type test struct {
		req []*pba.AgentSubscribeRequest
		model *pb.ModelDetails
		err error
	}
	serverNameUnknown := "foo"
	serverName := "server1"
	smallMemory := uint64(100)
	tests := []test {
		{req: []*pba.AgentSubscribeRequest{{ServerName: serverName, ReplicaIdx: 0,
			ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"sklearn"}}}},
			model: &pb.ModelDetails{Name: "model1", Uri: "gs://model", Requirements: []string{"sklearn"}, Memory:&smallMemory, Replicas: 1, Server: &serverName},
			err: nil},
		{req: []*pba.AgentSubscribeRequest{{ServerName: serverName, ReplicaIdx: 0,
			ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"sklearn"}}}},
			model: &pb.ModelDetails{Name: "model1", Uri: "gs://model", Requirements: []string{"sklearn"}, Memory:&smallMemory, Replicas: 1, Server: &serverNameUnknown},
			err: ServerNotFoundErr},
	}
	for _,test := range tests {
		m := createTestMemoryScheduler()
		for _, repReq := range test.req {
			err := m.UpdateServerReplica(repReq) // Create server and replicas
			g.Expect(err).To(BeNil())
		}
		err := m.CreateModel(test.model.Name, test.model)
		g.Expect(err).To(BeNil())
		err = m.UpdateModelOnServer(test.model.Name, *test.model.Server)
		if test.err == nil {
			g.Expect(err).To(BeNil())
		} else {
			g.Expect(errors.Is(err,test.err)).To(BeTrue())
			g.Expect(m.isFailedModel(test.model.Name))
		}
	}
}

func TestRemoveModel(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)

	type test struct {
		req []*pba.AgentSubscribeRequest
		model *pb.ModelDetails
		err error
	}
	modelName := "model1"
	smallMemory := uint64(100)
	tests := []test {
		{req: []*pba.AgentSubscribeRequest{{ServerName: "server1", ReplicaIdx: 0,
			ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"sklearn"}}}},
			model: &pb.ModelDetails{Name: modelName, Uri: "gs://model", Requirements: []string{"sklearn"}, Memory:&smallMemory, Replicas: 1},
			err: nil}, // simple create
		{req: []*pba.AgentSubscribeRequest{{ServerName: "server1", ReplicaIdx: 0,
			ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"sklearn","xgboost"}}}},
			model: &pb.ModelDetails{Name: modelName, Uri: "gs://model", Requirements: []string{"xgboost","sklearn"}, Memory:&smallMemory, Replicas: 1},
			err: nil}, // multiple requirements
		{req: []*pba.AgentSubscribeRequest{{ServerName: "server1", ReplicaIdx: 0,
			ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"sklearn"}}},
			{ServerName: "server1", ReplicaIdx: 1,
				ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"sklearn"}}}},
			model: &pb.ModelDetails{Name: modelName, Uri: "gs://model", Requirements: []string{"sklearn"}, Memory:&smallMemory, Replicas: 2},
			err: nil}, // schedule to 2 replicas
		{req: []*pba.AgentSubscribeRequest{{ServerName: "server1", ReplicaIdx: 0,
			ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"sklearn"}}}},
			model: nil,
			err: ModelNotFoundErr}, // fail to unload model that does not exist
	}
	for _,test := range tests {
		m := createTestMemoryScheduler()
		for _, repReq := range test.req {
			err := m.UpdateServerReplica(repReq) // Create server and replicas
			g.Expect(err).To(BeNil())
		}
		if test.model != nil {
			err := m.CreateModel(test.model.Name, test.model)
			g.Expect(err).To(BeNil())
			err = m.ScheduleModelToServer(test.model.Name)
			g.Expect(err).To(BeNil())
		}
		err := m.RemoveModel(modelName)
		if test.err == nil {
			g.Expect(err).To(BeNil())
		} else {
			g.Expect(errors.Is(err,test.err)).To(BeTrue())
		}
	}
}

func TestRemoveReplica(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)

	type test struct {
		req []*pba.AgentSubscribeRequest
		model *pb.ModelDetails
		removeReplicas []int
		err error
	}
	smallMemory := uint64(100)
	tests := []test {
		{req: []*pba.AgentSubscribeRequest{{ServerName: "server1", ReplicaIdx: 0,
			ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"sklearn"}}}},
			model: &pb.ModelDetails{Name: "model1", Uri: "gs://model", Requirements: []string{"sklearn"}, Memory:&smallMemory, Replicas: 1},
			removeReplicas: []int{0},
			err: RescheduleModelsFromReplicaErr}, // 1 replica so remove should cause reschedule failure
		{req: []*pba.AgentSubscribeRequest{{ServerName: "server1", ReplicaIdx: 0,
			ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"sklearn"}}},
			{ServerName: "server1", ReplicaIdx: 1,
				ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"sklearn"}}}},
			model: &pb.ModelDetails{Name: "model1", Uri: "gs://model", Requirements: []string{"sklearn"}, Memory:&smallMemory, Replicas: 1},
			removeReplicas: []int{1},
			err: nil}, // 2 replicas so reschedule should succeed
		{req: []*pba.AgentSubscribeRequest{{ServerName: "server1", ReplicaIdx: 0,
			ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"sklearn"}}},
			{ServerName: "server1", ReplicaIdx: 1,
				ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"foo"}}}},
			model: &pb.ModelDetails{Name: "model1", Uri: "gs://model", Requirements: []string{"sklearn"}, Memory:&smallMemory, Replicas: 1},
			removeReplicas: []int{0},
			err: RescheduleModelsFromReplicaErr}, // 2 replicas but reschedule should fail due to capabilities
		{req: []*pba.AgentSubscribeRequest{{ServerName: "server1", ReplicaIdx: 0,
			ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"sklearn"}}},
			{ServerName: "server1", ReplicaIdx: 1,
				ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 10, Capabilities: []string{"sklearn"}}}},
			model: &pb.ModelDetails{Name: "model1", Uri: "gs://model", Requirements: []string{"sklearn"}, Memory:&smallMemory, Replicas: 1},
			removeReplicas: []int{0},
			err: RescheduleModelsFromReplicaErr}, // 2 replicas but reschedule should fail due to memory
	}
	for tidx,test := range tests {
		t.Logf("start test %d", tidx)
		m := createTestMemoryScheduler()
		for _, repReq := range test.req {
			err := m.UpdateServerReplica(repReq) // Create server and replicas
			g.Expect(err).To(BeNil())
		}
		err := m.CreateModel(test.model.Name, test.model)
		g.Expect(err).To(BeNil())
		err = m.ScheduleModelToServer(test.model.Name)

		model, err := m.GetModel(test.model.Name)
		g.Expect(err).To(BeNil())
		for k := range model.replicas {
			err := m.SetModelState(test.model.Name, model.Server(), k, Loaded, nil)
			g.Expect(err).To(BeNil())
		}
		for _,removeReplicaIdx := range test.removeReplicas {
			err := m.RemoveServerReplicaAndRedeployModels(test.req[removeReplicaIdx].ServerName, int(test.req[removeReplicaIdx].GetReplicaIdx()))
			if test.err == nil {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(errors.Is(err,test.err)).To(BeTrue())
				g.Expect(m.isFailedModel(test.model.Name))
				g.Expect(len(m.getFailedModels())).To(Equal(1))
			}
		}
	}
}


func TestModelRequirements(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	type test struct {
		capabilities []string
		requirements []string
		availableMemory uint64
		memory uint64
		expected bool
	}
	tests := []test {
		{capabilities: []string{}, requirements: []string{}, availableMemory: 10, memory: 100, expected: false},
		{capabilities: []string{"sklearn"}, requirements: []string{"sklearn"}, availableMemory: 100, memory: 100, expected: true},
		{capabilities: []string{"sklearn"}, requirements: []string{"xgboost"}, availableMemory: 100, memory: 100, expected: false},
		{capabilities: []string{"sklearn"}, requirements: []string{"sklearn"}, availableMemory: 10, memory: 100, expected: false},
	}
	for tidx,test := range tests {
		t.Logf("start test %d", tidx)
		g.Expect(checkModelRequirements(test.capabilities, test.requirements, test.availableMemory, test.memory)).To(Equal(test.expected))
	}
}


func TestReplicaBounce(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)

	type test struct {
		req []*pba.AgentSubscribeRequest
		model *pb.ModelDetails
		removeReplicas []int
		err1 error
		bounceReq []*pba.AgentSubscribeRequest
		err2 error
	}
	smallMemory := uint64(100)
	tests := []test {
		{req: []*pba.AgentSubscribeRequest{{ServerName: "server1", ReplicaIdx: 0,
			ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"sklearn"}}}},
			model: &pb.ModelDetails{Name: "model1", Uri: "gs://model", Requirements: []string{"sklearn"}, Memory:&smallMemory, Replicas: 1},
			removeReplicas: []int{0},
			err1: RescheduleModelsFromReplicaErr,
			bounceReq: []*pba.AgentSubscribeRequest{{ServerName: "server1", ReplicaIdx: 0,
				ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"sklearn"}}}},
			err2: nil}, // 1 replica so remove should cause reschedule failure
		{req: []*pba.AgentSubscribeRequest{{ServerName: "server1", ReplicaIdx: 0,
			ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"sklearn"}}},
			{ServerName: "server1", ReplicaIdx: 1,
				ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"sklearn"}}}},
			model: &pb.ModelDetails{Name: "model1", Uri: "gs://model", Requirements: []string{"sklearn"}, Memory:&smallMemory, Replicas: 1},
			removeReplicas: []int{1},
			err1: nil,
			bounceReq: []*pba.AgentSubscribeRequest{{ServerName: "server1", ReplicaIdx: 1,
				ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"sklearn"}}}},
			err2: nil}, // 2 replicas so reschedule should succeed
		{req: []*pba.AgentSubscribeRequest{{ServerName: "server1", ReplicaIdx: 0,
			ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"sklearn"}}},
			{ServerName: "server1", ReplicaIdx: 1,
				ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"foo"}}}},
			model: &pb.ModelDetails{Name: "model1", Uri: "gs://model", Requirements: []string{"sklearn"}, Memory:&smallMemory, Replicas: 1},
			removeReplicas: []int{0},
			err1: RescheduleModelsFromReplicaErr,
			bounceReq: []*pba.AgentSubscribeRequest{{ServerName: "server1", ReplicaIdx: 0,
				ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"sklearn"}}}},
			err2: nil}, // 2 replicas but reschedule should fail due to capabilities but then succeed on bounce
		{req: []*pba.AgentSubscribeRequest{
			{ServerName: "server1", ReplicaIdx: 0,
				ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"sklearn"}}},
			{ServerName: "server1", ReplicaIdx: 1,
				ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 10, Capabilities: []string{"sklearn"}}}},
			model: &pb.ModelDetails{Name: "model1", Uri: "gs://model", Requirements: []string{"sklearn"}, Memory:&smallMemory, Replicas: 1},
			removeReplicas: []int{0},
			err1: RescheduleModelsFromReplicaErr,
			bounceReq: []*pba.AgentSubscribeRequest{{ServerName: "server1", ReplicaIdx: 0,
				ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"sklearn"}}}},
			err2: nil}, // 2 replicas but reschedule should fail due to memory but succeed on bounce
	}
	for tidx,test := range tests {
		t.Logf("start test %d", tidx)
		m := createTestMemoryScheduler()
		for _, repReq := range test.req {
			err := m.UpdateServerReplica(repReq) // Create server and replicas
			g.Expect(err).To(BeNil())
		}
		err := m.CreateModel(test.model.Name, test.model)
		g.Expect(err).To(BeNil())
		err = m.ScheduleModelToServer(test.model.Name)
		g.Expect(err).To(BeNil())
		model, err := m.GetModel(test.model.Name)
		g.Expect(err).To(BeNil())
		for k := range model.replicas {
			err := m.SetModelState(test.model.Name, model.Server(), k, Loaded, nil)
			g.Expect(err).To(BeNil())
		}
		for _,removeReplicaIdx := range test.removeReplicas {
			err := m.RemoveServerReplicaAndRedeployModels(test.req[removeReplicaIdx].ServerName, int(test.req[removeReplicaIdx].GetReplicaIdx()))
			if test.err1 == nil {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err).ToNot(BeNil())
				g.Expect(errors.Is(err,test.err1)).To(BeTrue())
				g.Expect(m.isFailedModel(test.model.Name))
				g.Expect(len(m.getFailedModels())).To(Equal(1))
			}
		}
		for _,repReq := range test.bounceReq {
			err := m.UpdateServerReplica(repReq)
			g.Expect(err).To(BeNil())
		}
	}
}