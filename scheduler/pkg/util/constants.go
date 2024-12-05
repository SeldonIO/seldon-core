/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package util

import (
	"time"
)

// Headers
const (
	SeldonModelHeader          = "seldon-model"
	SeldonPipelineHeader       = "pipeline"
	SeldonInternalModelHeader  = "seldon-internal-model"
	SeldonLoggingHeader        = "Seldon-Logging"
	SeldonRouteHeader          = "x-seldon-route"
	ExternalHeaderPrefix       = "x-"
	SeldonModelHeaderSuffix    = "model"
	SeldonPipelineHeaderSuffix = "pipeline"
)

// REST
const (
	DefaultReverseProxyHTTPPort = 9999
	MaxIdleConnsHTTP            = 10
	MaxIdleConnsPerHostHTTP     = 10
	DisableKeepAlivesHTTP       = false
	MaxConnsPerHostHTTP         = 20
	DefaultTimeoutSeconds       = 5
	IdleConnTimeoutSeconds      = 60 * 30
)

// GRPC
const (
	GRPCRetryBackoff             = 100 * time.Millisecond
	GRPCRetryMaxCount            = 5 // around 3.2s in total wait duration
	GRPCMaxMsgSizeBytes          = 1000 * 1024 * 1024
	GRPCModelServerLoadTimeout   = 60 * time.Minute // How long to wait for a model to load? think of LLM Load, maybe should be a config
	GRPCModelServerUnloadTimeout = 2 * time.Minute
	GRPCControlPlaneTimeout      = 1 * time.Minute // For control plane operations except load/unload
)

const (
	EnvoyUpdateDefaultBatchWait = 250 * time.Millisecond
	// note that we keep client and server keepalive times the same
	// they need to match counterparts in controller client: operator/scheduler/client.go
	// and dataflow-engine: data-flow/src/main/kotlin/io/seldon/dataflow/PipelineSubscriber.kt
	gRPCKeepAliveTime         = 60 * time.Second
	clientKeepAliveTimeout    = 2 * time.Second
	gRPCKeepAlivePermit       = false
	MaxGRPCRetriesOnStream    = 100 // this is at the grpc library level
	backOffExpMaxElapsedTime  = 0   // Never stop due to large time between calls
	backOffExpMaxInterval     = time.Second * 15
	backOffExpInitialInterval = time.Second
)
