/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package utils

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestMergeMaps(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name     string
		m1       map[string]string
		m2       map[string]string
		expected map[string]string
	}

	tests := []test{
		{
			name:     "same mappings",
			m1:       map[string]string{"a": "a"},
			m2:       map[string]string{},
			expected: map[string]string{"a": "a"},
		},
		{
			name:     "different mappings",
			m1:       map[string]string{"a": "a"},
			m2:       map[string]string{"b": "b"},
			expected: map[string]string{"a": "a", "b": "b"},
		},
		{
			name:     "overrides",
			m1:       map[string]string{"a": "2"},
			m2:       map[string]string{"a": "1"},
			expected: map[string]string{"a": "2"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m3 := MergeMaps(test.m1, test.m2)
			for k, v := range test.expected {
				g.Expect(m3[k]).To(Equal(v))
			}
		})
	}
}

func TestHasMappings(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name     string
		m1       map[string]string
		m2       map[string]string
		expected bool
	}

	tests := []test{
		{
			name:     "same mappings",
			m1:       map[string]string{"a": "a"},
			m2:       map[string]string{"a": "a"},
			expected: true,
		},
		{
			name:     "subset",
			m1:       map[string]string{"a": "a"},
			m2:       map[string]string{"a": "a", "b": "b"},
			expected: true,
		},
		{
			name:     "fail value",
			m1:       map[string]string{"a": "a"},
			m2:       map[string]string{"a": "b", "b": "b"},
			expected: false,
		},
		{
			name:     "fail key",
			m1:       map[string]string{"a": "a"},
			m2:       map[string]string{"c": "c", "b": "b"},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g.Expect(HasMappings(test.m1, test.m2)).To(Equal(test.expected))
		})
	}
}
