/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
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
			_, err := newFileStore(prefix, suffix, logger)
			if test.err {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
			}
		})

	}
}
