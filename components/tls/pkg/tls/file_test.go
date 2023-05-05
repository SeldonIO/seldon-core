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

package tls

import (
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"testing"
)

func TestNewTlsFolderHandler(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name     string
		envs     map[string]string
		caOnly   bool
		expected *TlsFolderHandler
	}

	prefix := "SCHEDULER"
	tests := []test{
		{
			name: "ok",
			envs: map[string]string{
				prefix + EnvCrtLocationSuffix: "/tls.key",
				prefix + EnvKeyLocationSuffix: "/tls.crt",
				prefix + EnvCaLocationSuffix:  "/ca.crt",
			},
			caOnly: false,
			expected: &TlsFolderHandler{
				prefix:       prefix,
				certFilePath: "/tls.key",
				keyFilePath:  "/tls.crt",
				caFilePath:   "/ca.crt",
			},
		},
		{
			name: "crt missing",
			envs: map[string]string{
				prefix + EnvKeyLocationSuffix: "/tls.key",
				prefix + EnvCaLocationSuffix:  "/ca.crt",
			},
			caOnly: false,
			expected: &TlsFolderHandler{
				prefix:       prefix,
				certFilePath: getDefaultPath(prefix, EnvCrtLocationSuffix),
				keyFilePath:  "/tls.key",
				caFilePath:   "/ca.crt",
			},
		},
		{
			name: "key missing",
			envs: map[string]string{
				prefix + EnvCrtLocationSuffix: "/tls.crt",
				prefix + EnvCaLocationSuffix:  "/ca.crt",
			},
			caOnly: false,
			expected: &TlsFolderHandler{
				prefix:       prefix,
				certFilePath: "/tls.crt",
				keyFilePath:  getDefaultPath(prefix, EnvKeyLocationSuffix),
				caFilePath:   "/ca.crt",
			},
		},
		{
			name: "caonly",
			envs: map[string]string{
				prefix + EnvCaLocationSuffix: "/ca.crt",
			},
			caOnly: true,
			expected: &TlsFolderHandler{
				prefix:       prefix,
				certFilePath: "",
				keyFilePath:  "",
				caFilePath:   "/ca.crt",
			},
		},
		{
			name:   "ca missing",
			envs:   map[string]string{},
			caOnly: true,
			expected: &TlsFolderHandler{
				prefix:       prefix,
				certFilePath: "",
				keyFilePath:  "",
				caFilePath:   getDefaultPath(prefix, EnvCaLocationSuffix),
			},
		},
		{
			name:   "all missing",
			envs:   map[string]string{},
			caOnly: false,
			expected: &TlsFolderHandler{
				prefix:       prefix,
				certFilePath: getDefaultPath(prefix, EnvCrtLocationSuffix),
				keyFilePath:  getDefaultPath(prefix, EnvKeyLocationSuffix),
				caFilePath:   getDefaultPath(prefix, EnvCaLocationSuffix),
			},
		},
	}
	for _, test := range tests {
		logger := log.New()
		t.Run(test.name, func(t *testing.T) {
			for k, v := range test.envs {
				t.Setenv(k, v)
			}
			tls, err := NewTlsFolderHandler(prefix, test.caOnly, logger)
			g.Expect(err).To(BeNil())
			g.Expect(tls.certFilePath).To(Equal(test.expected.certFilePath))
			g.Expect(tls.keyFilePath).To(Equal(test.expected.keyFilePath))
			g.Expect(tls.caFilePath).To(Equal(test.expected.caFilePath))
		})

	}
}
