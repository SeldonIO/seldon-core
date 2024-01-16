/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package tls

import (
	"testing"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

func TestNewTlsFolderHandler(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name   string
		envs   map[string]string
		caOnly bool
		err    bool
	}

	prefix := "x"
	tests := []test{
		{
			name: "ok",
			envs: map[string]string{
				prefix + EnvCrtLocationSuffix: "/abc",
				prefix + EnvKeyLocationSuffix: "/abc",
				prefix + EnvCaLocationSuffix:  "/abc",
			},
			caOnly: false,
			err:    false,
		},
		{
			name: "crt missing",
			envs: map[string]string{
				prefix + EnvKeyLocationSuffix: "/abc",
				prefix + EnvCaLocationSuffix:  "/abc",
			},
			caOnly: false,
			err:    true,
		},
		{
			name: "key missing",
			envs: map[string]string{
				prefix + EnvCrtLocationSuffix: "/abc",
				prefix + EnvCaLocationSuffix:  "/abc",
			},
			caOnly: false,
			err:    true,
		},
		{
			name: "ca missing",
			envs: map[string]string{
				prefix + EnvCrtLocationSuffix: "/abc",
				prefix + EnvKeyLocationSuffix: "/abc",
			},
			caOnly: false,
			err:    true,
		},
		{
			name: "caonly",
			envs: map[string]string{
				prefix + EnvCaLocationSuffix: "/abc",
			},
			caOnly: true,
			err:    false,
		},
		{
			name:   "ca missing but optional",
			envs:   map[string]string{},
			caOnly: true,
			err:    false,
		},
	}
	for _, test := range tests {
		logger := log.New()
		t.Run(test.name, func(t *testing.T) {
			for k, v := range test.envs {
				t.Setenv(k, v)
			}
			_, err := NewTlsFolderHandler(prefix, test.caOnly, logger)
			if test.err {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
			}
		})

	}
}
