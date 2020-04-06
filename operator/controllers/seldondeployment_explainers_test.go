/*
Copyright 2020 The Seldon Team.

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

package controllers

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"github.com/seldonio/seldon-core/operator/constants"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

var _ = Describe("createExplainer", func() {
	var r *SeldonDeploymentReconciler
	var mlDep *machinelearningv1.SeldonDeployment
	var p *machinelearningv1.PredictorSpec
	var c *components
	var pSvcName string

	BeforeEach(func() {
		p = &machinelearningv1.PredictorSpec{
			Name: "main",
		}

		mlDep = &machinelearningv1.SeldonDeployment{
			ObjectMeta: metav1.ObjectMeta{
				Name: "my-model",
			},
			Spec: machinelearningv1.SeldonDeploymentSpec{
				Predictors: []machinelearningv1.PredictorSpec{*p},
			},
		}

		c = &components{}

		pSvcName = machinelearningv1.GetPredictorKey(mlDep, p)

		r = &SeldonDeploymentReconciler{
			Client:   k8sManager.GetClient(),
			Log:      ctrl.Log.WithName("controllers").WithName("SeldonDeployment"),
			Scheme:   k8sManager.GetScheme(),
			Recorder: k8sManager.GetEventRecorderFor(constants.ControllerName),
		}
	})

	DescribeTable(
		"Empty explainers should not create any component",
		func(explainer *machinelearningv1.Explainer) {
			p.Explainer = explainer
			err := createExplainer(r, mlDep, p, c, pSvcName, r.Log)

			Expect(err).ToNot(HaveOccurred())
			Expect(c.deployments).To(BeEmpty())
		},
		Entry("nil", nil),
		Entry("empty type", &machinelearningv1.Explainer{}),
	)
})
