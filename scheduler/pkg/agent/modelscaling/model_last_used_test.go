/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package modelscaling

import (
	"strconv"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/interfaces"
)

func setLastUsed(luKeeper *modelReplicaLastUsedKeeper, modelName string, ts uint32) error {
	if err := luKeeper.pq.Update(modelName, -int64(ts)); err != nil {
		return luKeeper.pq.Add(modelName, -int64(ts))
	} else {
		return err
	}
}

func TestModelLastUsedSimple(t *testing.T) {
	type operation uint
	const (
		inc operation = iota
		del
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
				err := setLastUsed(lu, test.modelName, uint32(time.Now().Unix()))
				g.Expect(err).To(BeNil())
			}
			switch test.op {
			case inc:
				err := lu.ModelInferEnter(test.modelName, "")
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
				err := lu.ModelInferEnter(strconv.Itoa(i), "")
				g.Expect(err).To(BeNil())
			}

			time.Sleep(time.Second * 1)

			for i := numModels / 2; i < numModels; i++ {
				err := lu.ModelInferEnter(strconv.Itoa(i), "")
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
