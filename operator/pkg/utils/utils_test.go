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
