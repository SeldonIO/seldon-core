/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
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
