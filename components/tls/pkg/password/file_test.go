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

package password

import (
	"testing"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

func TestNewPasswordFolderHandler(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name string
		envs map[string]string
		err  bool
	}

	prefix := "x"
	suffix := "_y"
	tests := []test{
		{
			name: "ok",
			envs: map[string]string{
				prefix + suffix: "/abc",
			},
			err: false,
		},
		{
			name: "no env",
			envs: map[string]string{},
			err:  true,
		},
	}
	for _, test := range tests {
		logger := log.New()
		t.Run(test.name, func(t *testing.T) {
			for k, v := range test.envs {
				t.Setenv(k, v)
			}
			_, err := NewPasswordFolderHandler(prefix, suffix, logger)
			if test.err {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
			}
		})

	}
}
