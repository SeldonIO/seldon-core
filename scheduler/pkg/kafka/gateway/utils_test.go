package gateway

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestConsistentHash(t *testing.T) {
	g := NewGomegaWithT(t)

	numBuckets := 10

	type test struct {
		name    string
		modelId string
		hash    uint32
	}
	tests := []test{
		{
			name:    "smoke test",
			modelId: "dumm1",
			hash:    modelIdToConsumerBucket("dumm1", numBuckets),
		},
		{
			name:    "smoke empty test",
			modelId: "",
			hash:    modelIdToConsumerBucket("", numBuckets),
		},
		{
			name:    "smoke escape chars test",
			modelId: "x%$\\",
			hash:    modelIdToConsumerBucket("x%$\\", numBuckets),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g.Expect(modelIdToConsumerBucket(test.modelId, numBuckets)).To(Equal(test.hash))

		})
	}
}
