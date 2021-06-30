package seldon

import (
	"context"
	"fmt"
	"sync"
	"math/rand"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/seldonio/seldon-core/executor/api/client"

	"github.com/golang/protobuf/ptypes/empty"
	grpc2 "github.com/seldonio/seldon-core/executor/api/grpc"
	"github.com/seldonio/seldon-core/executor/api/grpc/seldon/proto"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/executor/api/util"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"google.golang.org/grpc"
	"io"
	"math"
	"net/http"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

// TODO: make this configurable
var numConns = 10

type SeldonMessageGrpcClient struct {
    sync.RWMutex
	Log            logr.Logger
	callOptions    []grpc.CallOption
	conns          map[string][]*grpc.ClientConn
	Predictor      *v1.PredictorSpec
	DeploymentName string
	annotations    map[string]string
}

func (s *SeldonMessageGrpcClient) IsGrpc() bool {
	return true
}

func NewSeldonGrpcClient(spec *v1.PredictorSpec, deploymentName string, annotations map[string]string) client.SeldonApiClient {
	opts := []grpc.CallOption{
		grpc.MaxCallSendMsgSize(math.MaxInt32),
		grpc.MaxCallRecvMsgSize(math.MaxInt32),
	}
	smgc := SeldonMessageGrpcClient{
		Log:            logf.Log.WithName("SeldonGrpcClient"),
		callOptions:    opts,
		conns:          make(map[string][]*grpc.ClientConn),
		Predictor:      spec,
		DeploymentName: deploymentName,
		annotations:    annotations,
	}
	return &smgc
}

// TODO: Re-examine locks here, make sure concurrent reads/writes aren't overwriting each other
// TODO: Verify this logic is necessary, or if we can just create numConns connections at once (simplifying logic)
// TODO: Investigate a TLL for conns--may result in better load balancing if connections occasionally re-connect
func (s *SeldonMessageGrpcClient) getConnection(host string, port int32, modelName string) (*grpc.ClientConn, error) {
    s.RLock()
    randNum := rand.Intn(numConns)
	k := fmt.Sprintf("%s:%d", host, port)
	if nodeConns, ok := s.conns[k]; ok {
	    if c := nodeConns[randNum]; c != nil {
	        defer s.RUnlock()
	        return c, nil
	    }
	    s.RUnlock()

	    c, err := s.createNewConn(modelName, host, port)
	    if err != nil {
	        return nil, err
	    }

	    s.Lock()
	    s.conns[k][randNum] = c
	    s.Unlock()

		return s.conns[k][randNum], nil
	} else {
	    s.RUnlock()
	    connList := make([]*grpc.ClientConn, numConns)

	    c, err := s.createNewConn(modelName, host, port)
	    if err != nil {
	        return nil, err
	    }

	    connList[randNum] = c

	    s.Lock()
	    s.conns[k] = connList
	    s.Unlock()

		return s.conns[k][randNum], nil
	}
}

func (s *SeldonMessageGrpcClient) createNewConn(modelName, host string, port int32) (*grpc.ClientConn, error) {
    opts := []grpc.DialOption{
        grpc.WithInsecure(),
    }

    opts = append(opts, grpc2.AddClientInterceptors(s.Predictor, s.DeploymentName, modelName, s.annotations, s.Log))
    conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, port), opts...)
    if err != nil {
        return nil, err
    }
    return conn, nil
}

func (s *SeldonMessageGrpcClient) Chain(ctx context.Context, modelName string, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	return msg, nil
}

func (s *SeldonMessageGrpcClient) Predict(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	conn, err := s.getConnection(host, port, modelName)
	if err != nil {
		return s.CreateErrorPayload(err), err
	}
	grpcClient := proto.NewModelClient(conn)
	resp, err := grpcClient.Predict(grpc2.AddMetadataToOutgoingGrpcContext(ctx, meta), msg.GetPayload().(*proto.SeldonMessage), s.callOptions...)
	if err != nil {
		return s.CreateErrorPayload(err), err
	}
	resPayload := payload.ProtoPayload{Msg: resp}
	return &resPayload, nil
}

func (s *SeldonMessageGrpcClient) TransformInput(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	conn, err := s.getConnection(host, port, modelName)
	if err != nil {
		return s.CreateErrorPayload(err), err
	}
	grpcClient := proto.NewTransformerClient(conn)
	resp, err := grpcClient.TransformInput(grpc2.AddMetadataToOutgoingGrpcContext(ctx, meta), msg.GetPayload().(*proto.SeldonMessage), s.callOptions...)
	if err != nil {
		return s.CreateErrorPayload(err), err
	}
	resPayload := payload.ProtoPayload{Msg: resp}
	return &resPayload, nil
}

func (s *SeldonMessageGrpcClient) Route(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (int, error) {
	conn, err := s.getConnection(host, port, modelName)
	if err != nil {
		return 0, err
	}
	grpcClient := proto.NewRouterClient(conn)
	resp, err := grpcClient.Route(grpc2.AddMetadataToOutgoingGrpcContext(ctx, meta), msg.GetPayload().(*proto.SeldonMessage), s.callOptions...)
	if err != nil {
		return 0, err
	}
	routes := util.ExtractRouteFromSeldonMessage(resp)
	//Only returning first route. API could be extended to allow multiple routes
	return routes[0], nil
}

func (s *SeldonMessageGrpcClient) Combine(ctx context.Context, modelName string, host string, port int32, msgs []payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	conn, err := s.getConnection(host, port, modelName)
	if err != nil {
		return s.CreateErrorPayload(err), err
	}
	sms := make([]*proto.SeldonMessage, len(msgs))
	for i, sm := range msgs {
		sms[i] = sm.GetPayload().(*proto.SeldonMessage)
	}
	grpcClient := proto.NewCombinerClient(conn)
	sml := proto.SeldonMessageList{SeldonMessages: sms}
	resp, err := grpcClient.Aggregate(grpc2.AddMetadataToOutgoingGrpcContext(ctx, meta), &sml, s.callOptions...)
	if err != nil {
		return s.CreateErrorPayload(err), err
	}
	resPayload := payload.ProtoPayload{Msg: resp}
	return &resPayload, nil
}

func (s *SeldonMessageGrpcClient) TransformOutput(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	conn, err := s.getConnection(host, port, modelName)
	if err != nil {
		return s.CreateErrorPayload(err), err
	}
	grpcClient := proto.NewOutputTransformerClient(conn)
	resp, err := grpcClient.TransformOutput(grpc2.AddMetadataToOutgoingGrpcContext(ctx, meta), msg.GetPayload().(*proto.SeldonMessage), s.callOptions...)
	if err != nil {
		return s.CreateErrorPayload(err), err
	}
	resPayload := payload.ProtoPayload{Msg: resp}
	return &resPayload, nil
}

func (s *SeldonMessageGrpcClient) Feedback(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	conn, err := s.getConnection(host, port, modelName)
	if err != nil {
		return s.CreateErrorPayload(err), err
	}
	grpcClient := proto.NewModelClient(conn)
	resp, err := grpcClient.SendFeedback(grpc2.AddMetadataToOutgoingGrpcContext(ctx, meta), msg.GetPayload().(*proto.Feedback), s.callOptions...)
	if err != nil {
		return s.CreateErrorPayload(err), err
	}
	resPayload := payload.ProtoPayload{Msg: resp}
	return &resPayload, nil
}

func (s *SeldonMessageGrpcClient) Unmarshall(msg []byte, contentType string) (payload.SeldonPayload, error) {
	panic("Not implemented")
}

func (s *SeldonMessageGrpcClient) Marshall(out io.Writer, msg payload.SeldonPayload) error {
	panic("Not implemented")
}

func (s *SeldonMessageGrpcClient) CreateErrorPayload(err error) payload.SeldonPayload {
	respFailed := proto.SeldonMessage{Status: &proto.Status{Code: http.StatusInternalServerError, Info: err.Error()}}
	res := payload.ProtoPayload{Msg: &respFailed}
	return &res
}

func (s *SeldonMessageGrpcClient) Status(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	return nil, errors.Errorf("Not implemented")
}

// Return model's metadata as payload.SeldonPaylaod (to expose as received on corresponding executor endpoint)
func (s *SeldonMessageGrpcClient) Metadata(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	conn, err := s.getConnection(host, port, modelName)
	if err != nil {
		return s.CreateErrorPayload(err), err
	}
	grpcClient := proto.NewModelClient(conn)
	resp, err := grpcClient.Metadata(grpc2.AddMetadataToOutgoingGrpcContext(ctx, meta), &empty.Empty{}, s.callOptions...)
	if err != nil {
		return s.CreateErrorPayload(err), err
	}

	resPayload := payload.ProtoPayload{Msg: resp}
	return &resPayload, nil
}

// Return model's metadata decoded to payload.ModelMetadata (to build GraphMetadata)
func (s *SeldonMessageGrpcClient) ModelMetadata(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.ModelMetadata, error) {
	resPayload, err := s.Metadata(ctx, modelName, host, port, msg, meta)
	if err != nil {
		return payload.ModelMetadata{}, err
	}

	protoPayload, ok := resPayload.GetPayload().(*proto.SeldonModelMetadata)
	if !ok {
		return payload.ModelMetadata{}, errors.New("Wrong Payload")
	}
	output := payload.ModelMetadata{
		Name:     protoPayload.GetName(),
		Platform: protoPayload.GetPlatform(),
		Versions: protoPayload.GetVersions(),
		Inputs:   protoPayload.GetInputs(),
		Outputs:  protoPayload.GetOutputs(),
		Custom:   protoPayload.GetCustom(),
	}
	return output, nil
}
