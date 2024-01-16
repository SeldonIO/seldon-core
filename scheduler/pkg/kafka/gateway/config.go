/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package gateway

type KafkaModelConfig struct {
	ModelName   string
	InputTopic  string
	OutputTopic string
	ErrorTopic  string
}

type InferenceServerConfig struct {
	Host     string
	HttpPort int
	GrpcPort int
}
