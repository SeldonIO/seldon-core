package utils

import (
	"os"
	"testing"

	. "github.com/onsi/gomega"
)

func TestGetEnvAsBool(t *testing.T) {
	g := NewGomegaWithT(t)

	tests := []struct {
		raw      string
		expected bool
	}{
		{
			raw:      "true",
			expected: true,
		},
		{
			raw:      "TRUE",
			expected: true,
		},
		{
			raw:      "1",
			expected: true,
		},
		{
			raw:      "false",
			expected: false,
		},
		{
			raw:      "FALSE",
			expected: false,
		},
		{
			raw:      "0",
			expected: false,
		},
		{
			raw:      "foo",
			expected: false,
		},
		{
			raw:      "",
			expected: false,
		},
		{
			raw:      "345",
			expected: false,
		},
	}

	for _, test := range tests {
		os.Setenv("TEST_FOO", test.raw)
		val := GetEnvAsBool("TEST_FOO", false)

		g.Expect(val).To(Equal(test.expected))
	}
}
