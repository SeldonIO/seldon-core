package rest

import (
	"github.com/onsi/gomega"
	"testing"
)

func TestConversions(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	val := "[1,2]"
	arr, err := ExtractRouteAsJsonArray([]byte(val))
	g.Expect(err).Should(gomega.BeNil())
	g.Expect(arr[0]).Should(gomega.Equal(1))
	g.Expect(arr[1]).Should(gomega.Equal(2))
}
