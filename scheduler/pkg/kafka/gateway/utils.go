/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package gateway

import (
	"net/http"
	"strings"

	"google.golang.org/grpc/metadata"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/envoy/resources"
)

func extractHeadersHttp(headers http.Header) map[string][]string {
	filteredHeaders := make(map[string][]string)
	for k, v := range headers {
		if strings.HasPrefix(k, resources.ExternalHeaderPrefix) {
			filteredHeaders[k] = v
		}
	}
	return filteredHeaders
}

func extractHeadersGrpc(headers metadata.MD, trailers metadata.MD) map[string][]string {
	filteredHeaders := make(map[string][]string)
	for k, v := range headers {
		if strings.HasPrefix(k, resources.ExternalHeaderPrefix) {
			filteredHeaders[k] = v
		}
	}
	for k, v := range trailers {
		if strings.HasPrefix(k, resources.ExternalHeaderPrefix) {
			filteredHeaders[k] = v
		}
	}
	return filteredHeaders
}
