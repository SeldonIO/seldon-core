package k8s

import (
	. "github.com/onsi/gomega"
	"testing"
)

func TestGetAnnotationsSimple(t *testing.T) {
	g := NewGomegaWithT(t)
	const data = `a=b`
	amap, err := getAnnotationMap(data)
	g.Expect(err).To(BeNil())
	g.Expect(amap["a"]).To(Equal("b"))
}

func TestGetAnnotationsQuotes(t *testing.T) {
	g := NewGomegaWithT(t)
	const data = `"a"="b"`
	amap, err := getAnnotationMap(data)
	g.Expect(err).To(BeNil())
	g.Expect(amap["a"]).To(Equal("b"))
}

func TestGetAnnotationsEmptyLine(t *testing.T) {
	g := NewGomegaWithT(t)
	const data = `a=b
`
	amap, err := getAnnotationMap(data)
	g.Expect(err).To(BeNil())
	g.Expect(amap["a"]).To(Equal("b"))
}

func TestGetAnnotationsBad(t *testing.T) {
	g := NewGomegaWithT(t)
	const data = `bad`
	amap, err := getAnnotationMap(data)
	g.Expect(err).NotTo(BeNil())
	g.Expect(amap).To(BeNil())
}
