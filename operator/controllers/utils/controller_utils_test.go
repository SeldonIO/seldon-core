package utils

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
)

var _ = Describe("Controller utils", func() {
	DescribeTable(
		"isEmptyExplainer",
		func(explainer *machinelearningv1.Explainer, expected bool) {
			empty := IsEmptyExplainer(explainer)
			Expect(empty).To(Equal(expected))
		},
		Entry("empty if nil", nil, true),
		Entry("empty if unset type", &machinelearningv1.Explainer{}, true),
		Entry(
			"not empty otherwise",
			&machinelearningv1.Explainer{Type: machinelearningv1.AlibiAnchorsImageExplainer},
			false,
		),
	)
})
