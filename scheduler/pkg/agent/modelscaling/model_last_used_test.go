package modelscaling

import (
	"strconv"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/scheduler/pkg/agent/interfaces"
)

func TestModelLastUsedSimple(t *testing.T) {

	const (
		inc operation = iota
		del
		set
		add
	)

	dummyModel := "dummy_model"
	g := NewGomegaWithT(t)
	t.Logf("Start!")

	type test struct {
		name      string
		modelName string
		op        operation
		initial   uint32
	}
	tests := []test{
		{
			name:      "increment",
			modelName: dummyModel,
			op:        inc,
			initial:   0,
		},
		{
			name:      "increment new",
			modelName: dummyModel,
			op:        inc,
			initial:   0,
		},
		{
			name:      "set",
			modelName: dummyModel,
			op:        set,
			initial:   0,
		},
		{
			name:      "set new",
			modelName: dummyModel,
			op:        set,
			initial:   1,
		},
		{
			name:      "delete not there",
			modelName: dummyModel,
			op:        del,
			initial:   0,
		},
		{
			name:      "delete",
			modelName: dummyModel,
			op:        del,
			initial:   1,
		},
		{
			name:      "add not there",
			modelName: dummyModel,
			op:        add,
			initial:   0,
		},
		{
			name:      "add there",
			modelName: dummyModel,
			op:        add,
			initial:   1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lu := NewModelReplicaLastUsedKeeper()

			if test.initial > 0 {
				err := lu.Set(test.modelName, uint32(time.Now().Unix()))
				g.Expect(err).To(BeNil())
			}
			switch test.op {
			case inc:
				err := lu.IncDefault(test.modelName)
				g.Expect(err).To(BeNil())
			case set:
				err := lu.Set(test.modelName, uint32(time.Now().Unix()))
				g.Expect(err).To(BeNil())
			case del:
				err := lu.Delete(test.modelName)
				if test.initial > 0 {
					g.Expect(err).To(BeNil())
				} else {
					g.Expect(err).NotTo(BeNil())
				}
			case add:
				err := lu.Add(test.modelName)
				g.Expect(err).To(BeNil())
			}

			val, err := lu.Get(test.modelName)
			if test.op == del {
				// we should get an error when we call Get after Delete
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(val).Should(BeNumerically("<=", uint32(time.Now().Unix())))
				g.Expect(val).NotTo(Equal(uint32(0)))
			}
		})
	}

	t.Logf("Done!")
}

func TestModelLastUsedThreshold(t *testing.T) {
	numModels := 20
	g := NewGomegaWithT(t)
	t.Logf("Start!")

	type test struct {
		name           string
		threshold      uint32
		op             interfaces.LogicOperation
		expectedResult int
	}
	tests := []test{
		{
			name:           "withrange",
			threshold:      1,
			expectedResult: 10,
		},
		{
			name:           "outofrange",
			threshold:      5,
			op:             interfaces.Gte,
			expectedResult: 0,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lu := NewModelReplicaLastUsedKeeper()
			for i := 0; i < numModels/2; i++ {
				err := lu.IncDefault(strconv.Itoa(i))
				g.Expect(err).To(BeNil())
			}

			time.Sleep(time.Second * 1)

			for i := numModels / 2; i < numModels; i++ {
				err := lu.IncDefault(strconv.Itoa(i))
				g.Expect(err).To(BeNil())
			}

			models, err := lu.GetAll(test.threshold, test.op, false)
			g.Expect(err).To(BeNil())
			g.Expect(len(models)).To(Equal(test.expectedResult))

			for _, model := range models {
				g.Expect(strconv.Atoi(model.ModelName)).Should(BeNumerically("<=", test.expectedResult)) // only triggered when we get results back
				g.Expect(model.Key).To(Equal(ModelLastUsedKey))
			}

			models, err = lu.GetAll(0, test.op, true)
			g.Expect(err).To(BeNil())
			g.Expect(len(models)).To(Equal(numModels))

		})
	}

	t.Logf("Done!")
}
