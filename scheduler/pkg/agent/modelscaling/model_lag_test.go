/*
Copyright 2022 Seldon Technologies Ltd.

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

package modelscaling

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/interfaces"
)

func getModelName(base string, idx int) string {
	return fmt.Sprintf("%s_%d", base, idx)
}

func TestModelLagsSimple(t *testing.T) {
	g := NewGomegaWithT(t)
	t.Logf("Start!")

	type test struct {
		name     string
		op       operation
		initial  uint32
		expected uint32
	}
	tests := []test{
		{
			name:     "increment",
			op:       inc,
			initial:  0,
			expected: 1,
		},
		{
			name:     "decrement",
			op:       dec,
			initial:  1,
			expected: 0,
		},
		{
			name:     "decrement",
			op:       dec,
			initial:  0,
			expected: 0,
		},
		{
			name:     "reset",
			op:       reset,
			initial:  2,
			expected: 0,
		},
		{
			name:     "reset",
			op:       reset,
			initial:  0,
			expected: 0,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lag := &lagKeeper{
				lag: test.initial,
			}

			switch test.op {
			case inc:
				lag.inc()
			case dec:
				lag.dec()
			case reset:
				lag.reset()
			}

			g.Expect(lag.get()).To(Equal(test.expected))
		})
	}

	t.Logf("Done!")
}

func TestModelReplicaLagsSimple(t *testing.T) {
	dummyModel := "dummy_model"
	g := NewGomegaWithT(t)
	t.Logf("Start!")

	const (
		inc operation = iota
		dec
		reset
		add
		del
	)

	type test struct {
		name      string
		modelName string
		op        operation
		initial   uint32
		expected  uint32
	}
	tests := []test{
		{
			name:      "increment",
			modelName: dummyModel,
			op:        inc,
			initial:   0,
			expected:  1,
		},
		{
			name:      "decrement",
			modelName: dummyModel,
			op:        dec,
			initial:   1,
			expected:  0,
		},
		{
			name:      "reset",
			modelName: dummyModel,
			op:        reset,
			initial:   2,
			expected:  0,
		},
		{
			name:      "add",
			modelName: dummyModel,
			op:        add,
			initial:   0,
			expected:  1,
		},
		{
			name:      "del",
			modelName: dummyModel,
			op:        del,
			initial:   1,
			expected:  0,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lags := NewModelReplicaLagsKeeper()

			err := lags.Set(test.modelName, test.initial)
			g.Expect(err).To(BeNil())

			switch test.op {
			case inc:
				err := lags.IncDefault(test.modelName)
				g.Expect(err).To(BeNil())
			case dec:
				err := lags.DecDefault(test.modelName)
				g.Expect(err).To(BeNil())
			case reset:
				err := lags.Reset(test.modelName)
				g.Expect(err).To(BeNil())
			case add:
				err := lags.Add(test.modelName)
				g.Expect(err).To(BeNil())
			case del:
				err := lags.Delete(test.modelName)
				g.Expect(err).To(BeNil())
			}

			if test.op != del {
				g.Expect(lags.Get(test.modelName)).To(Equal(test.expected))
			} else {
				_, err := lags.Get(test.modelName)
				g.Expect(err).NotTo(BeNil())
			}
		})
	}

	t.Logf("Done!")
}

func TestModelReplicaLagsConcurrent(t *testing.T) {
	dummyModelBase := "dummy_model"
	maxNumModels := 100

	g := NewGomegaWithT(t)
	t.Logf("Start!")

	fnWrapper := func(fn func(modelName string) error, modelName string, wg *sync.WaitGroup) {
		sleepMs := rand.Intn(100)
		time.Sleep(time.Millisecond * time.Duration(sleepMs))
		err := fn(modelName)
		g.Expect(err).To(BeNil())
		wg.Done()
	}

	type test struct {
		name      string
		numModels int
		op        operation
	}
	tests := []test{
		{
			name:      "increment",
			numModels: maxNumModels,
			op:        inc,
		},
		{
			name:      "decrement",
			numModels: maxNumModels,
			op:        dec,
		},
		{
			name:      "reset",
			numModels: maxNumModels,
			op:        reset,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lags := NewModelReplicaLagsKeeper()

			t.Logf("setup")
			// we create state according to the model idx (i) for example
			// dummy_model_10 -> 10
			// up to numModels
			for i := 0; i < test.numModels; i++ {
				for j := 0; j < i; j++ {
					err := lags.IncDefault(getModelName(dummyModelBase, i))
					g.Expect(err).To(BeNil())
				}
			}

			var wg sync.WaitGroup
			wg.Add(test.numModels)

			// we then create jobs one for each model according to the operation in the test
			// each operation can wait an arbitrary amount of time before executing.
			for i := 0; i < test.numModels; i++ {
				switch test.op {
				case inc:
					go fnWrapper(lags.IncDefault, getModelName(dummyModelBase, i), &wg)
				case dec:
					go fnWrapper(lags.DecDefault, getModelName(dummyModelBase, i), &wg)
				case reset:
					go fnWrapper(lags.Reset, getModelName(dummyModelBase, i), &wg)
				}
			}

			//wait for all operations to finish
			wg.Wait()

			t.Logf("check")

			for i := 0; i < test.numModels; i++ {
				switch test.op {
				// note: state of the model is at i
				case inc:
					// we add one to the state of the model, so it should be i + 1
					g.Expect(lags.Get(getModelName(dummyModelBase, i))).To(Equal(uint32(i + 1)))
				case dec:
					// we subtract one from the state of the model, so it should be i
					if i > 0 {
						g.Expect(lags.Get(getModelName(dummyModelBase, i))).To(Equal(uint32(i - 1)))
					} else {
						g.Expect(lags.Get(getModelName(dummyModelBase, i))).To(Equal(uint32(0)))
					}
				case reset:
					// we reset, so it should be 0
					g.Expect(lags.Get(getModelName(dummyModelBase, i))).To(Equal(uint32(0)))
				}
			}

		})
	}

	t.Logf("Done!")
}

func TestThreholdFilter(t *testing.T) {
	g := NewGomegaWithT(t)
	t.Logf("Start!")

	type test struct {
		name           string
		init           map[string]uint32
		threshold      uint32
		op             interfaces.LogicOperation
		expectedResult []interfaces.ModelStatsKV
	}
	tests := []test{
		{
			name: "withinrange",
			init: map[string]uint32{
				"1": 1,
				"2": 2,
				"3": 3,
			},
			op:        interfaces.Gte,
			threshold: 2,
			expectedResult: []interfaces.ModelStatsKV{
				{
					Value:     2,
					Key:       ModelLagKey,
					ModelName: "2",
				},
				{
					Value:     3,
					Key:       ModelLagKey,
					ModelName: "3",
				},
			},
		},
		{
			name: "outofrange",
			init: map[string]uint32{
				"1": 1,
				"2": 2,
				"3": 3,
			},
			op:             interfaces.Gte,
			threshold:      4,
			expectedResult: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lags := NewModelReplicaLagsKeeper()

			for model, lag := range test.init {
				err := lags.Set(model, lag)
				g.Expect(err).To(BeNil())
			}

			result, err := lags.GetAll(test.threshold, test.op, false)

			if test.expectedResult != nil {
				for _, item := range result {
					found := false
					for _, expectedItem := range test.expectedResult {
						if item.ModelName == expectedItem.ModelName {
							g.Expect(*item).To(Equal(expectedItem))
							found = true
						}
					}
					g.Expect(found).To(BeTrue())

				}
				g.Expect(len(result)).To(Equal(len(test.expectedResult)))
			} else {
				g.Expect(len(result)).To(Equal(0))
			}
			g.Expect(err).To(BeNil())

			t.Logf("Check state without reset")
			for _, item := range result {
				g.Expect(lags.Get(item.ModelName)).To(Equal(item.Value))
			}

			t.Logf("Check state with reset")
			result, _ = lags.GetAll(test.threshold, test.op, true)
			for _, item := range result {
				g.Expect(lags.Get(item.ModelName)).To(Equal(uint32(0)))
			}

		})
	}
}
