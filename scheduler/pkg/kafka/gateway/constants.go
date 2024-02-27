/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package gateway

import "time"

const (
	HeaderKeyType       = "seldon-infer-type"
	HeaderValueJsonReq  = "json/inferModelRequest"
	HeaderValueJsonRes  = "json/inferModelResponse"
	HeaderValueProtoReq = "proto/InferModelRequest"
	HeaderValueProtoRes = "proto/InferModelResponse"

	// Topic creation retries
	TopicCreateTimeout      = time.Minute
	TopicDescribeTimeout    = time.Second
	TopicDescribeMaxRetries = 60
	TopicDescribeRetryDelay = time.Second
)
