package util

import (
	"testing"

	"github.com/golang/protobuf/jsonpb"
	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/executor/api/grpc/seldon/proto"
)

func TestExtractRouteFromSeldonMessage(t *testing.T) {
	g := NewGomegaWithT(t)

	cases := []struct {
		msg      string
		expected []int
	}{
		{
			msg:      `{"data":{"names":["X1L"],"ndarray":[[1]]}}`,
			expected: []int{1},
		},
		{
			msg:      `{"data":{"ndarray":[2]}}`,
			expected: []int{2},
		},
		{
			msg:      `{"data":{"ndarray":[3,4]}}`,
			expected: []int{3, 4},
		},
		{
			msg:      `{"data":{"names":["X1L","X2L"],"ndarray":[[1,2],[3,4]]}}`,
			expected: []int{1, 3},
		},
		{
			msg:      `{"data":{"ndarray":[]}}`,
			expected: []int{},
		},
	}

	for _, c := range cases {
		var sm proto.SeldonMessage
		jsonpb.UnmarshalString(c.msg, &sm)
		routes := ExtractRouteFromSeldonMessage(&sm)

		g.Expect(routes).To(Equal(c.expected))
	}
}
