/*
Copyright 2022 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"strings"

	v2 "github.com/seldonio/seldon-core/scheduler/apis/mlops/v2_dataplane"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

const (
	pipelineHeader = "pipeline"
	modelName      = "model.test" // Should be ignored by data-flow engine
	modelVersion   = "1"
	pipelineId     = "4312"

	numMessages = uint(5)
)

type Args struct {
	bootstrapServers string
	securityProtocol string
	inputTopics      []string
	outputTopics     []string
	pipelineHeader   string
}

func parseArguments(logger log.FieldLogger) *Args {
	b := flag.String("bootstrap-servers", "localhost:9092", "Kafka bootstrap servers, comma-separated")
	s := flag.String("security-protocol", "PLAINTEXT", "Kafka security protocol, e.g. PLAINTEXT")
	h := flag.String("pipeline-header", "some-pipeline", "Value to write for the 'pipeline' header")
	its := flag.String("input-topics", "", "Topics to produce values for")
	ots := flag.String("output-topics", "", "Topics to consume values for")

	flag.Parse()

	if strings.TrimSpace(*its) == "" && strings.TrimSpace(*ots) == "" {
		logger.Fatalln("input or output topics need to be specified")
	}

	inputTopics := strings.Split(*its, ",")
	for idx, t := range inputTopics {
		inputTopics[idx] = strings.TrimSpace(t)
	}

	outputTopics := strings.Split(*ots, ",")
	for idx, t := range outputTopics {
		outputTopics[idx] = strings.TrimSpace(t)
	}

	return &Args{
		bootstrapServers: *b,
		securityProtocol: *s,
		inputTopics:      inputTopics,
		outputTopics:     outputTopics,
		pipelineHeader:   *h,
	}
}

// TODO - knowledge of this should be passed into the producer,
//	as right now it's awkward this is defined here but used solely
//  by the producer
func makeV2Response(data int32) *[]byte {
	req := &v2.ModelInferResponse{
		ModelName:    modelName,
		ModelVersion: modelVersion,
		Id:           pipelineId,
		Outputs: []*v2.ModelInferResponse_InferOutputTensor{
			{
				Name:     "tensor0",
				Datatype: "INT32",
				Shape:    []int64{1, 2},
				Contents: &v2.InferTensorContents{
					IntContents: []int32{data, data},
				},
			},
			{
				Name:     "tensor1",
				Datatype: "INT32",
				Shape:    []int64{1, 2},
				Contents: &v2.InferTensorContents{
					IntContents: []int32{data, data},
				},
			},
		},
	}

	bs, err := proto.Marshal(req)
	if err != nil {
		return nil
	}

	return &bs
}

// TODO - construct _expected_ v2 request
func parseV2Request(data []byte) (*v2.ModelInferRequest, error) {
	msg := &v2.ModelInferRequest{}
	err := proto.Unmarshal(data, msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func makeV2Request(tensors []string, data int32) *v2.ModelInferRequest {
	msg := &v2.ModelInferRequest{
		ModelName:    modelName,
		ModelVersion: modelVersion,
		Id:           pipelineId,
	}

	for _, t := range tensors {
		msg.Inputs = append(
			msg.Inputs,
			&v2.ModelInferRequest_InferInputTensor{
				Name:     t,
				Datatype: "INT32",
				Shape:    []int64{1, 2},
				Contents: &v2.InferTensorContents{
					IntContents: []int32{data, data},
				},
			},
		)
	}

	return msg
}
