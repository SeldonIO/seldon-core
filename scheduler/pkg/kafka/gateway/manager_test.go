package gateway

import (
	"testing"

	"github.com/seldonio/seldon-core/scheduler/pkg/kafka/config"

	seldontracer "github.com/seldonio/seldon-core/scheduler/pkg/tracing"

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
			expectedInputTopic:  "seldon.default.model.foo.inputs",
			expectedOutputTopic: "seldon.default.model.foo.outputs",
			expectedErrorTopic:  "seldon.default.errors.errors",
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
			expectedErrorTopic:  "seldon.default.errors.errors",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := log.New()
			tp, err := seldontracer.NewTracer("test")
			g.Expect(err).To(BeNil())
			km := NewKafkaManager(logger, &KafkaServerConfig{}, "default", &config.KafkaConfig{}, tp)
			err = km.AddModel(test.modelName, test.streamSpec)
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
	tp, err := seldontracer.NewTracer("test")
	g.Expect(err).To(BeNil())
	gw1, err := NewInferKafkaGateway(log.New(), 0, &config.KafkaConfig{}, &KafkaModelConfig{ModelName: "foo", InputTopic: "topic1", OutputTopic: "topic2"}, &KafkaServerConfig{}, tp)
	g.Expect(err).To(BeNil())
	gw2, err := NewInferKafkaGateway(log.New(), 0, &config.KafkaConfig{}, &KafkaModelConfig{ModelName: "foo2", InputTopic: "topic2", OutputTopic: "topic3"}, &KafkaServerConfig{}, tp)
	g.Expect(err).To(BeNil())
	gw3, err := NewInferKafkaGateway(log.New(), 0, &config.KafkaConfig{}, &KafkaModelConfig{ModelName: "foo2", InputTopic: "topic2", OutputTopic: "topic3"}, &KafkaServerConfig{}, tp)
	g.Expect(err).To(BeNil())
	tests := []test{
		{
			name: "no current gateways",
			manager: &KafkaManager{
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
