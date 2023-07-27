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

const (
	ModelStatKey = "stat"
)

func getModelName(base string, idx int) string {
	return fmt.Sprintf("%s_%d", base, idx)
}

type dummyModelStats struct {
	value uint32
}

func (stats *dummyModelStats) Enter(requestId string) error {
	stats.value += 1
	return nil
}
func (stats *dummyModelStats) Exit(requestId string) error {
	if stats.value != 0 {
		stats.value -= 1
	}
	return nil
}

func (stats *dummyModelStats) Get() uint32 {
	return stats.value
}

func (stats *dummyModelStats) Reset() error {
	stats.value = 0
	return nil
}

func TestModelStatsKeeperSimple(t *testing.T) {
	dummyModel := "dummy_model"
	g := NewGomegaWithT(t)
	t.Logf("Start!")

	type test struct {
		name      string
		modelName string
		initial   uint32
		expected  uint32
	}
	tests := []test{
		{
			name:      "increment",
			modelName: dummyModel,
			initial:   0,
			expected:  1,
		},
		{
			name:      "decrement",
			modelName: dummyModel,
			initial:   1,
			expected:  0,
		},
		{
			name:      "add",
			modelName: dummyModel,
			initial:   0,
			expected:  0,
		},
		{
			name:      "del",
			modelName: dummyModel,
			initial:   1,
			expected:  0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			keeper := NewModelStatsKeeper(ModelStatKey, func() interfaces.ModelStats {
				return &dummyModelStats{
					value: test.initial,
				}
			})

			switch test.name {
			case "increment":
				err := keeper.ModelInferEnter(test.modelName, "")
				g.Expect(err).To(BeNil())
			case "decrement":
				err := keeper.ModelInferExit(test.modelName, "")
				g.Expect(err).To(BeNil())

			case "add":
				err := keeper.Add(test.modelName)
				g.Expect(err).To(BeNil())
			case "del":
				err := keeper.Delete(test.modelName)
				g.Expect(err).To(BeNil())
			}

			if test.name != "del" {
				g.Expect(keeper.Get(test.modelName)).To(Equal(test.expected))
			} else {
				_, err := keeper.Get(test.modelName)
				g.Expect(err).NotTo(BeNil())
			}
		})
	}

	t.Logf("Done!")
}

func TestModelStatsKeeperConcurrent(t *testing.T) {
	dummyModelBase := "dummy_model"
	maxNumModels := 100

	g := NewGomegaWithT(t)
	t.Logf("Start!")

	fnWrapper := func(fn func(modelName, requestId string) error, modelName string, requestId string, wg *sync.WaitGroup) {
		sleepMs := rand.Intn(100)
		time.Sleep(time.Millisecond * time.Duration(sleepMs))
		err := fn(modelName, requestId)
		g.Expect(err).To(BeNil())
		wg.Done()
	}

	type test struct {
		name      string
		numModels int
	}
	tests := []test{
		{
			name:      "increment",
			numModels: maxNumModels,
		},
		{
			name:      "decrement",
			numModels: maxNumModels,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			keeper := NewModelStatsKeeper("stat", func() interfaces.ModelStats {
				return &dummyModelStats{
					value: 0,
				}
			})
			t.Logf("setup")
			// we create state according to the model idx (i) for example
			// dummy_model_10 -> 10
			// up to numModels
			for i := 0; i < test.numModels; i++ {
				for j := 0; j < i; j++ {
					requestId := fmt.Sprintf("request_id_%d_%d", i, j)
					err := keeper.ModelInferEnter(getModelName(dummyModelBase, i), requestId)
					g.Expect(err).To(BeNil())
				}
			}

			var wg sync.WaitGroup
			wg.Add(test.numModels)

			// we then create jobs one for each model according to the operation in the test
			// each operation can wait an arbitrary amount of time before executing.
			for i := 0; i < test.numModels; i++ {
				requestId := fmt.Sprintf("request_id_%d_0", i)
				switch test.name {
				case "increment":
					go fnWrapper(keeper.ModelInferEnter, getModelName(dummyModelBase, i), requestId, &wg)
				case "decrement":
					go fnWrapper(keeper.ModelInferExit, getModelName(dummyModelBase, i), requestId, &wg)
				}
			}

			//wait for all operations to finish
			wg.Wait()

			t.Logf("check")

			for i := 0; i < test.numModels; i++ {
				switch test.name {
				// note: state of the model is at i
				case "increment":
					// we add one to the state of the model, so it should be i + 1
					g.Expect(keeper.Get(getModelName(dummyModelBase, i))).To(Equal(uint32(i + 1)))
				case "decrement":
					// we subtract one from the state of the model, so it should be i
					if i > 0 {
						g.Expect(keeper.Get(getModelName(dummyModelBase, i))).To(Equal(uint32(i - 1)))
					} else {
						g.Expect(keeper.Get(getModelName(dummyModelBase, i))).To(Equal(uint32(0)))
					}
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
					Key:       ModelStatKey,
					ModelName: "2",
				},
				{
					Value:     3,
					Key:       ModelStatKey,
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

			keeper := NewModelStatsKeeper("stat", func() interfaces.ModelStats {
				return &dummyModelStats{
					value: 0,
				}
			})
			for model, lag := range test.init {
				keeper.Add(model)
				if stat, ok := keeper.stats[model]; ok {
					stat.(*dummyModelStats).value = lag
				}
			}

			result, err := keeper.GetAll(test.threshold, test.op, false)

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
				g.Expect(keeper.Get(item.ModelName)).To(Equal(item.Value))
			}

			t.Logf("Check state with reset")
			result, _ = keeper.GetAll(test.threshold, test.op, true)
			for _, item := range result {
				g.Expect(keeper.Get(item.ModelName)).To(Equal(uint32(0)))
			}

		})
	}
}
