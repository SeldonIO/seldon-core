package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	"github.com/seldonio/seldon-core/scheduler/apis/mlops/v2_dataplane"
	"google.golang.org/protobuf/proto"
)

const (
	SeldonPrefix        = "seldon"
	InputsSpecifier     = "inputs"
	OutputsSpecifier    = "outputs"
	PipelineSpecifier   = "pipeline"
	ModelSpecifier      = "model"
	kafkaTimeoutSeconds = 2
)

type KafkaClient struct {
	consumer        *kafka.Consumer
	schedulerClient *SchedulerClient
}

type PipelineTopics struct {
	pipeline string
	topics   []string
	tensor   string
}

type KafkaInspect struct {
	Topics []*KafkaInspectTopic `json:"topics"`
}

type KafkaInspectTopic struct {
	Name string                      `json:"name"`
	Msgs []*KafkaInspectTopicMessage `json:"msgs"`
}

type KafkaInspectTopicMessage struct {
	Headers map[string]string `json:"headers"`
	Key     string            `json:"key"`
	Value   json.RawMessage   `json:"value"`
}

func NewKafkaClient(kafkaBroker string, kafkaBrokerIsSet bool, schedulerHost string, schedulerHostIsSet bool) (*KafkaClient, error) {
	config, err := LoadSeldonCLIConfig()
	if err != nil {
		return nil, err
	}

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	// Overwrite broker if set in config
	if !kafkaBrokerIsSet && config.Kafka != nil && config.Kafka.Bootstrap != "" {
		kafkaBroker = config.Kafka.Bootstrap
	}
	consumerConfig := kafka.ConfigMap{
		"bootstrap.servers": kafkaBroker,
		"group.id":          fmt.Sprintf("seldon-cli-%d", r1.Int()),
		"auto.offset.reset": "largest",
	}

	if config.Kafka != nil {
		consumerConfig["security.protocol"] = "ssl"
		consumerConfig["ssl.ca.location"] = config.Kafka.CaPath
		consumerConfig["ssl.key.location"] = config.Kafka.KeyPath
		consumerConfig["ssl.certificate.location"] = config.Kafka.CrtPath
	}

	consumer, err := kafka.NewConsumer(&consumerConfig)
	if err != nil {
		return nil, err
	}

	scheduler, err := NewSchedulerClient(schedulerHost, schedulerHostIsSet, "")
	if err != nil {
		return nil, err
	}
	kc := &KafkaClient{
		consumer:        consumer,
		schedulerClient: scheduler,
	}
	return kc, nil
}

func (kc *KafkaClient) subscribeAndSetOffset(pipelineStep string, offset int64) error {

	md, err := kc.consumer.GetMetadata(&pipelineStep, false, 1000)
	if err != nil {
		return err
	}

	for _, partitionMeta := range md.Topics[pipelineStep].Partitions {
		err := kc.consumer.Assign([]kafka.TopicPartition{
			{
				Topic:     &pipelineStep,
				Partition: partitionMeta.ID,
				//Note will get more messages than requested when multiple partitions available
				Offset: kafka.OffsetTail(kafka.Offset(offset)),
			},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func hasStep(stepName string, response *scheduler.PipelineStatusResponse) bool {
	version := response.Versions[len(response.Versions)-1]
	for _, step := range version.GetPipeline().Steps {
		if step.Name == stepName {
			return true
		}
	}
	return false
}

func createPipelineInspectTopics(pipelineSpec string, response *scheduler.PipelineStatusResponse, namespace string) (*PipelineTopics, error) {
	parts := strings.Split(pipelineSpec, ".")
	switch len(parts) {
	case 1: //Just pipeline - show all steps and pipeline itself
		var topics []string
		for _, step := range response.Versions[len(response.Versions)-1].Pipeline.Steps {
			topics = append(topics, fmt.Sprintf("%s.%s.%s.%s.%s", SeldonPrefix, namespace, ModelSpecifier, step.Name, InputsSpecifier))
			topics = append(topics, fmt.Sprintf("%s.%s.%s.%s.%s", SeldonPrefix, namespace, ModelSpecifier, step.Name, OutputsSpecifier))
		}
		topics = append(topics, fmt.Sprintf("%s.%s.%s.%s.%s", SeldonPrefix, namespace, PipelineSpecifier, parts[0], InputsSpecifier))
		topics = append(topics, fmt.Sprintf("%s.%s.%s.%s.%s", SeldonPrefix, namespace, PipelineSpecifier, parts[0], OutputsSpecifier))
		return &PipelineTopics{
			pipeline: pipelineSpec,
			topics:   topics,
		}, nil
	case 2:
		if parts[1] == InputsSpecifier || parts[1] == OutputsSpecifier {
			return &PipelineTopics{
				pipeline: parts[0],
				topics:   []string{fmt.Sprintf("%s.%s.%s.%s.%s", SeldonPrefix, namespace, PipelineSpecifier, parts[0], parts[1])},
			}, nil
		} else {
			if hasStep(parts[1], response) {
				return &PipelineTopics{
					pipeline: parts[0],
					topics: []string{
						fmt.Sprintf("%s.%s.%s.%s.%s", SeldonPrefix, namespace, ModelSpecifier, parts[1], InputsSpecifier),
						fmt.Sprintf("%s.%s.%s.%s.%s", SeldonPrefix, namespace, ModelSpecifier, parts[1], OutputsSpecifier),
					},
				}, nil
			} else {
				return nil, fmt.Errorf("Failed to find step with name %s in pipeline %s", parts[1], parts[0])
			}
		}
	case 3:
		if hasStep(parts[1], response) {
			if parts[2] == InputsSpecifier || parts[2] == OutputsSpecifier {
				return &PipelineTopics{
					pipeline: parts[0],
					topics: []string{
						fmt.Sprintf("%s.%s.%s.%s.%s", SeldonPrefix, namespace, ModelSpecifier, parts[1], parts[2]),
					},
				}, nil
			} else {
				return nil, fmt.Errorf("Need to specify either %s or %s for a step", InputsSpecifier, OutputsSpecifier)
			}
		} else {
			return nil, fmt.Errorf("Failed to find step with name %s in pipeline %s", parts[1], parts[0])
		}
	case 4:
		if hasStep(parts[1], response) {
			if parts[2] == InputsSpecifier || parts[2] == OutputsSpecifier {
				return &PipelineTopics{
					pipeline: parts[0],
					topics: []string{
						fmt.Sprintf("%s.%s.%s.%s.%s", SeldonPrefix, namespace, ModelSpecifier, parts[1], parts[2]),
					},
					tensor: parts[3],
				}, nil
			} else {
				return nil, fmt.Errorf("Need to specify either %s or %s for a step", InputsSpecifier, OutputsSpecifier)
			}
		} else {
			return nil, fmt.Errorf("Failed to find step with name %s in pipeline %s", parts[1], parts[0])
		}
	default:
		return nil, fmt.Errorf("Bad pipeline specifier %s", pipelineSpec)
	}
}

func (kc *KafkaClient) getPipelineStatus(pipelineSpec string) (*scheduler.PipelineStatusResponse, error) {
	parts := strings.Split(pipelineSpec, ".")
	pipeline := parts[0]
	conn, err := kc.schedulerClient.newConnection()
	if err != nil {
		return nil, err
	}
	grpcClient := scheduler.NewSchedulerClient(conn)
	res, err := kc.schedulerClient.getPipelineStatus(grpcClient, &scheduler.PipelineStatusRequest{SubscriberName: "cli", Name: &pipeline})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (kc *KafkaClient) InspectStep(pipelineStep string, offset int64, key string, format string, verbose bool, namespace string) error {
	status, err := kc.getPipelineStatus(pipelineStep)
	if err != nil {
		return err
	}
	pipelineTopics, err := createPipelineInspectTopics(pipelineStep, status, namespace)
	if err != nil {
		return err
	}

	ki := KafkaInspect{}
	for _, topic := range pipelineTopics.topics {
		kit, err := kc.createInspectTopic(topic, pipelineTopics.tensor, offset, key, verbose)
		if err != nil {
			return err
		}
		ki.Topics = append(ki.Topics, kit)
	}

	if format == InspectFormatJson {
		b, err := json.Marshal(ki)
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", string(b))
	} else {
		for _, topic := range ki.Topics {
			for _, msg := range topic.Msgs {
				if verbose {
					fmt.Printf("%s\t%s\t%s\t", topic.Name, msg.Key, msg.Value)
					for k, v := range msg.Headers {
						fmt.Printf("\t%s=%s", k, v)
					}
					fmt.Println("")
				} else {
					fmt.Printf("%s\t%s\t%s\n", topic.Name, msg.Key, msg.Value)
				}
			}
		}
	}

	// Fast close requires maybe: https://github.com/confluentinc/confluent-kafka-go/pull/757
	//_ = kc.consumer.Close()
	return nil
}

func (kc *KafkaClient) createInspectTopic(topic string, tensor string, offset int64, key string, verbose bool) (*KafkaInspectTopic, error) {
	kit := KafkaInspectTopic{
		Name: topic,
	}
	err := kc.subscribeAndSetOffset(topic, offset)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), kafkaTimeoutSeconds*time.Second)
	defer cancel()

	run := true
	var seen int64
	for run {
		select {
		case <-ctx.Done():
			run = false
		default:
			ev := kc.consumer.Poll(1000)
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *kafka.Message:
				seen = seen + 1
				if (string(e.Key) == key) || key == "" {
					kitm, err := createKafkaMsg(e, topic, tensor, verbose)
					if err != nil {
						return nil, err
					}
					kit.Msgs = append(kit.Msgs, kitm)
				}
				if seen >= offset {
					run = false
				}
			case kafka.Error:
				return nil, fmt.Errorf(e.Error())
			default:
				continue
			}
		}
	}

	return &kit, nil
}

func addInspectHeaders(e *kafka.Message, kitm *KafkaInspectTopicMessage) {
	kitm.Headers = make(map[string]string)
	for _, header := range e.Headers {
		kitm.Headers[header.Key] = string(header.Value)
	}
}

func createKafkaMsg(e *kafka.Message, topic string, tensor string, verbose bool) (*KafkaInspectTopicMessage, error) {
	kitm := KafkaInspectTopicMessage{
		Key: string(e.Key),
	}
	if verbose { // Only add headers in verbose mode
		addInspectHeaders(e, &kitm)
	}
	var err error
	if strings.HasSuffix(topic, OutputsSpecifier) {
		err = addInspectKafkaOutputMsg(e, tensor, &kitm)
	} else {
		err = addInspectKafkaInputMsg(e, tensor, &kitm)
	}
	return &kitm, err
}

func protoTojson(msg proto.Message) (json.RawMessage, error) {
	b, err := protojson.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func addInspectKafkaOutputMsg(e *kafka.Message, tensor string, kitm *KafkaInspectTopicMessage) error {
	res := &v2_dataplane.ModelInferResponse{}
	err := proto.Unmarshal(e.Value, res)
	if err != nil {
		kitm.Value = json.RawMessage(e.Value)
		return nil
	}
	err = updateResponseFromRawContents(res)
	if err != nil {
		return err
	}
	if tensor != "" {
		for _, output := range res.Outputs {
			if output.Name == tensor {
				kitm.Value, err = protoTojson(output)
				if err != nil {
					return err
				}
			}
		}

	} else {
		kitm.Value, err = protoTojson(res)
		if err != nil {
			return err
		}
	}
	return nil
}

func addInspectKafkaInputMsg(e *kafka.Message, tensor string, kitm *KafkaInspectTopicMessage) error {
	req := &v2_dataplane.ModelInferRequest{}
	err := proto.Unmarshal(e.Value, req)
	if err != nil {
		kitm.Value = json.RawMessage(e.Value)
		return nil
	}
	err = updateRequestFromRawContents(req)
	if err != nil {
		return err
	}
	if tensor != "" {
		for _, input := range req.Inputs {
			if input.Name == tensor {
				kitm.Value, err = protoTojson(input)
				if err != nil {
					return err
				}
			}
		}

	} else {
		kitm.Value, err = protoTojson(req)
		if err != nil {
			return err
		}
	}
	return nil
}
