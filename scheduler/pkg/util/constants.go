/*
Copyright 2023 Seldon Technologies Ltd.

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

package util

import "time"

const (
	// REST constants
	DefaultReverseProxyHTTPPort = 9999
	MaxIdleConnsHTTP            = 10
	MaxIdleConnsPerHostHTTP     = 10
	DisableKeepAlivesHTTP       = false
	MaxConnsPerHostHTTP         = 20
	DefaultTimeoutSeconds       = 5
	IdleConnTimeoutSeconds      = 60 * 30
)

const (
	GrpcRetryBackoffMillisecs         = 100
	GrpcRetryMaxCount                 = 5 // around 3.2s in total wait duration
	GrpcMaxMsgSizeBytes               = 1000 * 1024 * 1024
	EnvoyUpdateDefaultBatchWaitMillis = 250 * time.Millisecond
)
