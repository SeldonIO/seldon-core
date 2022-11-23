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

package util

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func GetHttpClientFromTLSOptions(tlsOptions *TLSOptions) *http.Client {
	if tlsOptions.TLS {
		t := &http.Transport{
			TLSClientConfig: tlsOptions.Cert.CreateClientTLSConfig(),
		}
		return &http.Client{Transport: otelhttp.NewTransport(t)}
	} else {
		return &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
	}
}

func GetHttpClient() (*http.Client, error) {
	tlsOptions, err := CreateTLSClientOptions()
	if err != nil {
		return nil, err
	}
	return GetHttpClientFromTLSOptions(tlsOptions), nil
}
