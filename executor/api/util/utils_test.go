package util

import (
	"os"
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
