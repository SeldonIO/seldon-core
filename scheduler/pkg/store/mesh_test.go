/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package store

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestCleanCapabilities(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name     string
		in       []string
		expected []string
	}

	tests := []test{
		{
			name:     "misc",
			in:       []string{"mlserver", " foo ", " bar", "bar   "},
			expected: []string{"mlserver", "foo", "bar", "bar"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			out := cleanCapabilities(test.in)
			g.Expect(out).To(Equal(test.expected))
		})
	}
}
