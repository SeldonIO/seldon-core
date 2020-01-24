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
