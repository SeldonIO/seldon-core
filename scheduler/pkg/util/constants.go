/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
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
