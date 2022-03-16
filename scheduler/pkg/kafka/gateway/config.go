package gateway

type KafkaModelConfig struct {
	ModelName   string
	InputTopic  string
	OutputTopic string
	ErrorTopic  string
}

type KafkaServerConfig struct {
	Host     string
	HttpPort int
	GrpcPort int
}
