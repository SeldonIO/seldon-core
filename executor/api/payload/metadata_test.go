package payload

import (
	. "github.com/onsi/gomega"
	"testing"
)

func TestNewFromtMeta(t *testing.T) {
	g := NewGomegaWithT(t)

	const k = "foo"
	const v = "bar"
	meta := NewFromMap(map[string][]string{k: []string{v}})

	g.Expect(meta.Meta[k][0]).To(Equal(v))
}

func TestGetAsBoolean(t *testing.T) {
	g := NewGomegaWithT(t)

	tests := []struct {
		val      []string
		def      bool
		expected bool
	}{
		{
			val:      []string{"true"},
			def:      false,
			expected: true,
		},
		{
			val:      []string{"on"},
			def:      false,
			expected: true,
		},
		{
			val:      []string{"1"},
			def:      false,
			expected: true,
		},
		{
			val:      []string{"TRUE", "0"},
			def:      false,
			expected: true,
		},
		{
			val:      []string{"false", "1"},
			def:      false,
			expected: true,
		},
		{
			val:      []string{"false"},
			def:      true,
			expected: false,
		},
		{
			val:      []string{"0"},
			def:      true,
			expected: false,
		},
		{
			val:      []string{"asdasd"},
			def:      true,
			expected: false,
		},
		{
			val:      []string{},
			def:      true,
			expected: true,
		},
		{
			val:      []string{},
			def:      false,
			expected: false,
		},
	}

	for _, test := range tests {
		m := map[string][]string{}
		if len(test.val) > 0 {
			m["foo"] = test.val
		}
		meta := NewFromMap(m)

		asBool := meta.GetAsBoolean("foo", test.def)
		g.Expect(asBool).To(Equal(test.expected))
	}
}
