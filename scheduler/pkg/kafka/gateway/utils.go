/*
Copyright 2022 Seldon Technologies Ltd.

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

package gateway

import (
	"net/http"
	"strings"

	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/resources"
	"google.golang.org/grpc/metadata"
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
