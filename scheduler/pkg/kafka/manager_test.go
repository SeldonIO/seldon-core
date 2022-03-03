package kafka

import (
	"testing"

	"github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

func TestManagerAddModel(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name                string
		modelName           string
		streamSpec          *scheduler.StreamSpec
		expectedInputTopic  string
		expectedOutputTopic string
		expectedErrorTopic  string
	}
	tests := []test{
		{
			name:                "basic with empty stream spec",
			modelName:           "foo",
			streamSpec:          nil,
			expectedInputTopic:  "foo-in",
			expectedOutputTopic: "foo-out",
			expectedErrorTopic:  "foo-err",
		},
		{
			name:      "basic with stream spec",
			modelName: "foo",
			streamSpec: &scheduler.StreamSpec{
				InputTopic:  "input",
				OutputTopic: "output",
			},
			expectedInputTopic:  "input",
			expectedOutputTopic: "output",
			expectedErrorTopic:  "foo-err",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := log.New()
			km := NewKafkaManager(logger, &KafkaServerConfig{})
			km.active = true
			err := km.AddModel(test.modelName, test.streamSpec)
			g.Expect(err).To(BeNil())
			g.Expect(km.gateways[test.modelName].modelConfig.ModelName).To(Equal(test.modelName))
			g.Expect(km.gateways[test.modelName].modelConfig.InputTopic).To(Equal(test.expectedInputTopic))
			g.Expect(km.gateways[test.modelName].modelConfig.OutputTopic).To(Equal(test.expectedOutputTopic))
			g.Expect(km.gateways[test.modelName].modelConfig.ErrorTopic).To(Equal(test.expectedErrorTopic))
			km.RemoveModel("foo1")
			err = km.Stop()
			g.Expect(err).To(BeNil())
		})
	}
}

func TestManagerRemoveModel(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name             string
		manager          *KafkaManager
		modelName        string
		expectedGateways int
	}
	gw1, err := NewInferKafkaGateway(log.New(), 0, "", &KafkaModelConfig{ModelName: "foo", InputTopic: "topic1", OutputTopic: "topic2"}, &KafkaServerConfig{})
	g.Expect(err).To(BeNil())
	gw2, err := NewInferKafkaGateway(log.New(), 0, "", &KafkaModelConfig{ModelName: "foo2", InputTopic: "topic2", OutputTopic: "topic3"}, &KafkaServerConfig{})
	g.Expect(err).To(BeNil())
	gw3, err := NewInferKafkaGateway(log.New(), 0, "", &KafkaModelConfig{ModelName: "foo2", InputTopic: "topic2", OutputTopic: "topic3"}, &KafkaServerConfig{})
	g.Expect(err).To(BeNil())
	tests := []test{
		{
			name: "no current gateways",
			manager: &KafkaManager{
				active:   true,
				logger:   log.New(),
				gateways: map[string]*InferKafkaGateway{},
				broker:   "",
			},
			modelName:        "foo",
			expectedGateways: 0,
		},
		{
			name: "gateway exists",
			manager: &KafkaManager{
				active: true,
				logger: log.New(),
				gateways: map[string]*InferKafkaGateway{
					"foo": gw1,
				},
				broker: "",
			},
			modelName:        "foo",
			expectedGateways: 0,
		},
		{
			name: "multiple kafka gateways",
			manager: &KafkaManager{
				active: true,
				logger: log.New(),
				gateways: map[string]*InferKafkaGateway{
					"foo":  gw2,
					"foo2": gw3,
				},
				broker: "",
			},
			modelName:        "foo",
			expectedGateways: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.manager.RemoveModel(test.modelName)
			g.Expect(test.manager.gateways[test.modelName]).To(BeNil())
			g.Expect(len(test.manager.gateways)).To(Equal(test.expectedGateways))
		})
	}
}
