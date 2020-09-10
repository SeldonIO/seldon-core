package controllers

import (
	"fmt"

	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"github.com/seldonio/seldon-core/operator/constants"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("MLServer helpers", func() {
	var pu *machinelearningv1.PredictiveUnit

	BeforeEach(func() {
		puType := machinelearningv1.MODEL
		puImplementation := machinelearningv1.PredictiveUnitImplementation(
			machinelearningv1.PrepackSklearnName,
		)

		pu = &machinelearningv1.PredictiveUnit{
			Name:           "my-model",
			Type:           &puType,
			Implementation: &puImplementation,
			Endpoint: &machinelearningv1.Endpoint{
				Type:        machinelearningv1.REST,
				ServicePort: int32(5001),
			},
		}
	})

	Describe("getMLServerPort", func() {
		DescribeTable(
			"returns the right port",
			func(endpointType machinelearningv1.EndpointType, serviceEndpointType machinelearningv1.EndpointType, expected int32) {
				pu.Endpoint.Type = serviceEndpointType

				port := getMLServerPort(pu, endpointType)
				Expect(port).To(Equal(expected))
			},
			Entry(
				"default httpPort",
				machinelearningv1.REST,
				machinelearningv1.GRPC,
				constants.MLServerDefaultHttpPort,
			),
			Entry(
				"default grpcPort",
				machinelearningv1.GRPC,
				machinelearningv1.REST,
				constants.MLServerDefaultGrpcPort,
			),
			Entry(
				"service httpPort",
				machinelearningv1.REST,
				machinelearningv1.REST,
				int32(5001),
			),
			Entry(
				"service grpcPort",
				machinelearningv1.GRPC,
				machinelearningv1.GRPC,
				int32(5001),
			),
		)
	})
})
