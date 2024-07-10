package scheduler

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
)

func TestGetNumExperiments(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name    string
		results []*scheduler.ExperimentStatusResponse
	}

	tests := []test{
		{
			name: "experiment ok",
			results: []*scheduler.ExperimentStatusResponse{
				{
					ExperimentName: "foo",
				},
				{
					ExperimentName: "bar",
				},
			},
		},
		{
			name:    "experiment ok",
			results: []*scheduler.ExperimentStatusResponse{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := mockSchedulerClient{responses_experiments: test.results}
			num, err := getNumExperimentsFromScheduler(context.Background(), &client)

			g.Expect(err).To(BeNil())
			g.Expect(num).To(Equal(len(test.results)))
		})
	}
}
