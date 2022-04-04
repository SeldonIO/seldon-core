package pipeline

import (
	"testing"

	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name                 string
		header               string
		expectedResourceName string
		isModel              bool
		error                bool
	}
	tests := []test{
		{
			name:                 "model no suffix",
			header:               "foo",
			expectedResourceName: "foo",
			isModel:              true,
		},
		{
			name:                 "pipeline",
			header:               "foo.pipeline",
			expectedResourceName: "foo",
			isModel:              false,
		},
		{
			name:                 "model with suffix",
			header:               "foo.model",
			expectedResourceName: "foo",
			isModel:              true,
		},
		{
			name:   "model with too many parts",
			header: "foo.bar.model",
			error:  true,
		},
		{
			name:   "bad suffix",
			header: "foo.bar",
			error:  true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resouceName, isModel, err := createResourceNameFromHeader(test.header)
			if test.error {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(resouceName).To(Equal(test.expectedResourceName))
				g.Expect(isModel).To(Equal(test.isModel))
			}
		})
	}
}
