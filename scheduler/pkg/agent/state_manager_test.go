package agent

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	pba "github.com/seldonio/seldon-core/scheduler/apis/mlops/agent"
	pbs "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"

	log "github.com/sirupsen/logrus"

	. "github.com/onsi/gomega"
)

func getModelId(prefix string, suffix int) string {
	return prefix + "_" + strconv.Itoa(suffix)
}

func getDummyModelDetails(modelId string, memBytes uint64, version uint32) *pba.ModelVersion {
	meta := pbs.MetaData{
		Name: modelId,
	}
	model := pbs.Model{
		Meta: &meta,
		ModelSpec: &pbs.ModelSpec{
			Uri:         "gs://dummy",
			MemoryBytes: &memBytes,
		},
	}
	mv := pba.ModelVersion{
		Model:   &model,
		Version: version,
	}
	return &mv
}

func setupLocalTestManager(numModels int, modelPrefix string, v2Client *V2Client, capacity int) *LocalStateManager {

	logger := log.New()
	logger.SetLevel(log.DebugLevel)

	modelState := NewModelState()
	models := make([]string, numModels)
	for i := 0; i < numModels; i++ {
		models[i] = getModelId(modelPrefix, i)
	}
	//create mock v2 client
	if v2Client == nil {
		v2Client = createTestV2Client(models, 200)
	}
	manager := NewLocalStateManager(
		modelState,
		logger,
		v2Client,
		int64(capacity),
	)

	return manager

}

// this mimics LoadModel in client.go with regards to locking
func (manager *LocalStateManager) loadModelFn(modelVersionDetails *pba.ModelVersion) error {
	modelName := modelVersionDetails.GetModel().GetMeta().GetName()

	manager.modelLoadLockCreate(modelName)
	defer manager.modelLoadUnlock(modelName)

	return manager.LoadModelVersion(modelVersionDetails)
}

// this mimics UnoadModel in client.go with regards to locking
func (manager *LocalStateManager) unloadModelFn(modelVersionDetails *pba.ModelVersion) error {
	modelName := modelVersionDetails.GetModel().GetMeta().GetName()

	manager.modelLoadLockCreate(modelName)
	defer manager.modelLoadUnlock(modelName)

	return manager.UnloadModelVersion(modelVersionDetails)
}

func TestLocalStateManagerSmoke(t *testing.T) {
	//activate mock http server for v2
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	numModels := 100
	dummyModelPrefix := "dummy_model"

	manager := setupLocalTestManager(numModels, dummyModelPrefix, nil, numModels-2)

	g := NewGomegaWithT(t)

	for i := 0; i < numModels; i++ {
		modelName := getModelId(dummyModelPrefix, i)
		memBytes := uint64(1)
		err := manager.LoadModelVersion(getDummyModelDetails(modelName, memBytes, uint32(1)))
		g.Expect(err).To(BeNil())
	}

	for i := numModels - 1; i >= 0; i-- {
		modelName := getModelId(dummyModelPrefix, i)
		err := manager.EnsureLoadModel(modelName)
		g.Expect(err).To(BeNil())
	}

}

// Ensures that we have a lock on model reloading
func TestConcurrentReload(t *testing.T) {
	dummyModelPrefix := "dummy_model"

	g := NewGomegaWithT(t)

	type test struct {
		name                    string
		numModels               int
		capacity                int
		expectedAvailableMemory uint64
	}
	tests := []test{
		{
			name:                    "enough capacity",
			numModels:               1000,
			capacity:                1100,
			expectedAvailableMemory: 100,
		},
		{
			name:                    "just enough capacity",
			numModels:               1100,
			capacity:                1100,
			expectedAvailableMemory: 0,
		},
		{
			name:                    "not enough capacity",
			numModels:               1100,
			capacity:                1000,
			expectedAvailableMemory: 0,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			//activate mock http server for v2
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			manager := setupLocalTestManager(test.numModels, dummyModelPrefix, nil, test.capacity)

			// load the first numModels, this will evict in reverse order
			for i := test.numModels - 1; i >= 0; i-- {
				modelName := getModelId(dummyModelPrefix, i)
				memBytes := uint64(1)
				_ = manager.LoadModelVersion(getDummyModelDetails(modelName, memBytes, uint32(1)))
			}

			// parallel load last model
			var wg sync.WaitGroup
			wg.Add(1000)
			for i := 0; i < 1000; i++ {
				modelName := getModelId(dummyModelPrefix, test.numModels-1)

				checkerFn := func(wg *sync.WaitGroup, modelName string) {
					err := manager.EnsureLoadModel(modelName)
					if err != nil {
						t.Logf("Error %s", err)
					}
					g.Expect(err).To(BeNil())
					wg.Done()
				}

				go checkerFn(&wg, modelName)
			}
			wg.Wait()

			g.Expect(manager.availableMemoryBytes).Should(BeNumerically("==", test.expectedAvailableMemory))
			cacheItems, _ := manager.cache.GetItems()
			if test.expectedAvailableMemory == 0 {
				g.Expect(len(cacheItems)).Should(BeNumerically("==", test.capacity))
			} else {
				g.Expect(len(cacheItems)).Should(BeNumerically("==", test.numModels))
			}
		})
	}

}

// We have concurrent load and unload of models from the scheduler.
// This might get us into situations of race conditions, we check concurrently
// that we have available memory and then load the model.
func TestConcurrentLoad(t *testing.T) {
	dummyModelPrefix := "dummy_model"

	g := NewGomegaWithT(t)

	type test struct {
		name                    string
		numModels               int
		capacity                int
		expectedAvailableMemory uint64
	}
	tests := []test{
		{
			name:                    "enough capacity",
			numModels:               1000,
			capacity:                1100,
			expectedAvailableMemory: 100,
		},
		{
			name:                    "just enough capacity",
			numModels:               1100,
			capacity:                1100,
			expectedAvailableMemory: 0,
		},
		{
			name:                    "not enough capacity",
			numModels:               1100,
			capacity:                1000,
			expectedAvailableMemory: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			//activate mock http server for v2
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			manager := setupLocalTestManager(test.numModels, dummyModelPrefix, nil, test.capacity)

			var wg sync.WaitGroup
			wg.Add(test.numModels)
			for i := 0; i < test.numModels; i++ {
				modelName := getModelId(dummyModelPrefix, i)
				memBytes := uint64(1)
				checkerFn := func(wg *sync.WaitGroup, modelName string, modelVersion uint32) {
					err := manager.loadModelFn(getDummyModelDetails(modelName, memBytes, modelVersion))
					for err != nil {
						t.Logf("Error %s", err)
						time.Sleep(10 * time.Millisecond)
						err = manager.loadModelFn(getDummyModelDetails(modelName, memBytes, modelVersion))
					}
					g.Expect(manager.availableMemoryBytes).Should(BeNumerically(">=", 0))
					wg.Done()
				}

				go checkerFn(&wg, modelName, uint32(1))
			}
			wg.Wait()

			// memory available should be zero
			g.Expect(manager.GetAvailableMemoryBytes()).Should(BeNumerically("==", test.expectedAvailableMemory))
			g.Expect(manager.modelVersions.numModels()).Should(BeNumerically("==", test.numModels))

			// then do unload
			wg.Add(test.numModels)
			for i := 0; i < test.numModels; i++ {
				modelName := getModelId(dummyModelPrefix, i)
				memBytes := uint64(1)
				checkerFn := func(wg *sync.WaitGroup, modelName string, modelVersion uint32) {
					err := manager.unloadModelFn(getDummyModelDetails(modelName, memBytes, modelVersion))
					if err != nil {
						t.Logf("Error %s", err)
					}
					g.Expect(manager.availableMemoryBytes).Should(BeNumerically(">=", 0))
					wg.Done()
				}

				go checkerFn(&wg, modelName, uint32(1))
			}
			wg.Wait()

			// should be an empty server
			g.Expect(manager.availableMemoryBytes).Should(BeNumerically("==", test.capacity))
			g.Expect(manager.modelVersions.numModels()).Should(BeNumerically("==", 0))

		})
	}
}

func TestConcurrentLoadWithVersions(t *testing.T) {
	dummyModelPrefix := "dummy_model"

	g := NewGomegaWithT(t)

	numberOfVersionsToAdd := 2

	type test struct {
		name                    string
		numModels               int
		capacity                int
		expectedAvailableMemory uint64
	}
	tests := []test{
		{
			name:                    "enough capacity",
			numModels:               100,
			capacity:                100 * numberOfVersionsToAdd,
			expectedAvailableMemory: 0,
		},
		{
			name:                    "not enough capacity",
			numModels:               100,
			capacity:                (100 * numberOfVersionsToAdd) - (numberOfVersionsToAdd * 5),
			expectedAvailableMemory: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			//activate mock http server for v2
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			manager := setupLocalTestManager(test.numModels, dummyModelPrefix, nil, test.capacity)

			var wg sync.WaitGroup
			wg.Add(test.numModels * numberOfVersionsToAdd)

			checkerFn := func(wg *sync.WaitGroup, modelName string, memBytes uint64, modelVersion uint32) {
				err := manager.loadModelFn(getDummyModelDetails(modelName, memBytes, modelVersion))
				for err != nil {
					t.Logf("Error %s for model %s version %d", err, modelName, modelVersion)
					time.Sleep(10 * time.Millisecond)
					err = manager.loadModelFn(getDummyModelDetails(modelName, memBytes, modelVersion))
				}
				g.Expect(manager.availableMemoryBytes).Should(BeNumerically(">=", 0))
				wg.Done()
			}
			for i := 0; i < test.numModels; i++ {
				modelName := getModelId(dummyModelPrefix, i)
				memBytes := uint64(1)

				for j := 1; j <= numberOfVersionsToAdd; j++ {
					go checkerFn(&wg, modelName, memBytes, uint32(j))
				}
			}
			wg.Wait()

			// because of multiversions there could be cases where we need to evict more models than needed
			g.Expect(manager.GetAvailableMemoryBytes()).Should(BeNumerically(">=", test.expectedAvailableMemory))
			g.Expect(manager.modelVersions.numModels()).Should(BeNumerically("==", test.numModels))
			models := manager.modelVersions.modelNames()
			for _, model := range models {
				g.Expect(manager.modelVersions.numVersions(model)).To(Equal(numberOfVersionsToAdd))
			}

			// then do unload
			wg.Add(test.numModels * numberOfVersionsToAdd)

			checkerFn = func(wg *sync.WaitGroup, modelName string, memBytes uint64, modelVersion uint32) {
				err := manager.unloadModelFn(getDummyModelDetails(modelName, memBytes, modelVersion))
				if err != nil {
					t.Logf("Error %s", err)
				}
				wg.Done()
			}
			for i := 0; i < test.numModels; i++ {
				modelName := getModelId(dummyModelPrefix, i)
				memBytes := uint64(1)
				for j := 1; j <= numberOfVersionsToAdd; j++ {
					go checkerFn(&wg, modelName, memBytes, uint32(j))
				}
			}
			wg.Wait()

			// should be an empty server
			g.Expect(manager.availableMemoryBytes).Should(BeNumerically("==", test.capacity))
			g.Expect(manager.modelVersions.numModels()).Should(BeNumerically("==", 0))

		})
	}
}

func TestDataAndControlPlaneInteractionSmoke(t *testing.T) {
	dummyModelPrefix := "dummy_model"

	g := NewGomegaWithT(t)

	numberOfVersionsToAdd := 2

	type test struct {
		name      string
		numModels int
		capacity  int
	}
	tests := []test{
		{
			name:      "enough capacity",
			numModels: 100,
			capacity:  100 * numberOfVersionsToAdd,
		},
		{
			name:      "not enough capacity",
			numModels: 100,
			capacity:  (100 * numberOfVersionsToAdd) - (numberOfVersionsToAdd * 5),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			//activate mock http server for v2
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			manager := setupLocalTestManager(test.numModels, dummyModelPrefix, nil, test.capacity)

			t.Log("Add a single version for all models")
			// add a single version of all models before actual test
			for i := 0; i < test.numModels; i++ {
				modelName := getModelId(dummyModelPrefix, i)
				memBytes := uint64(1)
				modelVersion := uint32(1)
				_ = manager.loadModelFn(getDummyModelDetails(modelName, memBytes, modelVersion))
			}

			t.Log("Start test")
			var wg sync.WaitGroup
			// load / unload / infer for all models
			// test no dealoack etc.
			wg.Add(test.numModels * numberOfVersionsToAdd * 3)

			fn := func(wg *sync.WaitGroup, modelName string, memBytes uint64, modelVersion uint32) {
				op := rand.Intn(3)
				switch op {
				case 0:
					_ = manager.loadModelFn(getDummyModelDetails(modelName, memBytes, modelVersion))
				case 1:
					_ = manager.unloadModelFn(getDummyModelDetails(modelName, memBytes, modelVersion))
				case 2:
					_ = manager.EnsureLoadModel(modelName) // this can be any model version per test
				}
				wg.Done()
			}

			for i := 0; i < test.numModels*numberOfVersionsToAdd*3; i++ {
				modelName := getModelId(dummyModelPrefix, rand.Intn(test.numModels))
				memBytes := uint64(1)
				modelVersion := rand.Intn(numberOfVersionsToAdd) + 1

				go fn(&wg, modelName, memBytes, uint32(modelVersion))

			}
			wg.Wait()

			g.Expect(manager.availableMemoryBytes).Should(BeNumerically("<=", test.capacity))
			g.Expect(manager.modelVersions.numModels()).Should(BeNumerically("<=", test.numModels))

		})
	}
}

func TestControlAndDataPlaneUseCases(t *testing.T) {
	dummyModelPrefix := "dummy_model"

	g := NewGomegaWithT(t)

	numModels := 2
	capacity := 1
	memBytes := uint64(1)

	type step int
	const (
		controlPlaneLoad step = iota
		controlPlaneUnload
		dataPlaneInfer
	)

	type stepDetails struct {
		stepType      step
		modelIdSuffix int
		modelVersion  int
		inMemory      bool
		isLoaded      bool
	}

	type test struct {
		name                    string
		step1                   stepDetails
		step2                   stepDetails
		isError                 bool
		expectedNumModels       int
		expectedAvailableMemory int64
	}
	tests := []test{
		{
			name:                    "Load then Infer (new model)",
			step1:                   stepDetails{stepType: controlPlaneLoad, modelIdSuffix: 1, modelVersion: 1, inMemory: false, isLoaded: false},
			step2:                   stepDetails{stepType: dataPlaneInfer, modelIdSuffix: 1, modelVersion: 1, inMemory: false, isLoaded: false},
			isError:                 false,
			expectedNumModels:       2,
			expectedAvailableMemory: 0,
		},
		{
			// there could be a race condition here as the model might not be in
			// cache anymore on the data plane path
			name:                    "Infer then Unload (existing model - in memory)",
			step1:                   stepDetails{stepType: dataPlaneInfer, modelIdSuffix: 1, modelVersion: 1, inMemory: true, isLoaded: true},
			step2:                   stepDetails{stepType: controlPlaneUnload, modelIdSuffix: 1, modelVersion: 1, inMemory: true, isLoaded: true},
			isError:                 false,
			expectedNumModels:       1,
			expectedAvailableMemory: 1,
		},
		{
			// should be an error because model is unloaded first
			name:                    "Unload then Infer (existing model - in memory)",
			step1:                   stepDetails{stepType: controlPlaneUnload, modelIdSuffix: 1, modelVersion: 1, inMemory: true, isLoaded: true},
			step2:                   stepDetails{stepType: dataPlaneInfer, modelIdSuffix: 1, modelVersion: 1, inMemory: true, isLoaded: true},
			isError:                 true,
			expectedNumModels:       1,
			expectedAvailableMemory: 1,
		},
		{
			name:                    "Infer then Unload (existing model - not in memory)",
			step1:                   stepDetails{stepType: dataPlaneInfer, modelIdSuffix: 1, modelVersion: 1, inMemory: false, isLoaded: true},
			step2:                   stepDetails{stepType: controlPlaneUnload, modelIdSuffix: 1, modelVersion: 1, inMemory: false, isLoaded: true},
			isError:                 false,
			expectedNumModels:       1,
			expectedAvailableMemory: 1,
		},
		{
			name:                    "Infer then Unload other model being evicted",
			step1:                   stepDetails{stepType: dataPlaneInfer, modelIdSuffix: 1, modelVersion: 1, inMemory: false, isLoaded: true},
			step2:                   stepDetails{stepType: controlPlaneUnload, modelIdSuffix: 0, modelVersion: 1, inMemory: false, isLoaded: true},
			isError:                 false,
			expectedNumModels:       1,
			expectedAvailableMemory: 0,
		},
		{
			// note that this can sometimes be true under heavy load
			name:                    "Unload other model being evicted then Infer",
			step1:                   stepDetails{stepType: controlPlaneUnload, modelIdSuffix: 0, modelVersion: 1, inMemory: true, isLoaded: true},
			step2:                   stepDetails{stepType: dataPlaneInfer, modelIdSuffix: 1, modelVersion: 1, inMemory: false, isLoaded: true},
			isError:                 false,
			expectedNumModels:       1,
			expectedAvailableMemory: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			//activate mock http server for v2
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			manager := setupLocalTestManager(numModels, dummyModelPrefix, nil, capacity)
			var barrier sync.WaitGroup
			barrier.Add(1)

			// fill capacity
			for i := 0; i < capacity; i++ {
				modelName := getModelId(dummyModelPrefix, i)
				modelVersion := uint32(1)
				_ = manager.LoadModelVersion(getDummyModelDetails(modelName, memBytes, modelVersion))
			}

			var wg sync.WaitGroup
			repeats := 1 // for step 2
			if test.step2.stepType != controlPlaneUnload {
				// as otherwise we might have errors
				repeats = 10
			}
			wg.Add(1 + repeats)
			errors := make(chan error, 1+repeats)

			fn := func(wg *sync.WaitGroup, modelName string, memBytes uint64, modelVersion uint32, step step, sleep bool, errors chan<- error) {
				if sleep {
					time.Sleep(50 * time.Microsecond)
					barrier.Wait()
					if step == controlPlaneLoad || step == controlPlaneUnload {
						// mimics control plane locking
						manager.modelLoadLockCreate(modelName)
						defer manager.modelLoadUnlock(modelName)
					}
				} else {

					if step == controlPlaneLoad || step == controlPlaneUnload {
						// mimics control plane locking
						manager.modelLoadLockCreate(modelName)
						defer manager.modelLoadUnlock(modelName)
					}
					barrier.Done()
				}
				switch step {
				case controlPlaneLoad:
					errors <- manager.LoadModelVersion(getDummyModelDetails(modelName, memBytes, modelVersion))
				case controlPlaneUnload:
					errors <- manager.UnloadModelVersion(getDummyModelDetails(modelName, memBytes, modelVersion))
				case dataPlaneInfer:
					errors <- manager.EnsureLoadModel(modelName)
				}
				wg.Done()
			}

			t.Log("Setup step1")

			if test.step1.isLoaded {
				_ = manager.LoadModelVersion(getDummyModelDetails(getModelId(dummyModelPrefix, test.step1.modelIdSuffix), memBytes, uint32(test.step1.modelVersion)))
				if !test.step1.inMemory {
					// ensure load the other model 0, so evicts model_1 if in memory
					_ = manager.EnsureLoadModel(getModelId(dummyModelPrefix, 0))
				}
			}

			t.Log("Setup step2")

			if test.step2.isLoaded {
				_ = manager.LoadModelVersion(getDummyModelDetails(getModelId(dummyModelPrefix, test.step2.modelIdSuffix), memBytes, uint32(test.step2.modelVersion)))
				if !test.step2.inMemory {
					// ensure load the other model 0, so evicts model_1 if in memory
					_ = manager.EnsureLoadModel(getModelId(dummyModelPrefix, 0))
				}
			}

			t.Log("Start actual test")

			for i := 0; i < repeats; i++ {
				go fn(&wg, getModelId(dummyModelPrefix, test.step2.modelIdSuffix), memBytes, uint32(test.step2.modelVersion), test.step2.stepType, true, errors)
			}
			go fn(&wg, getModelId(dummyModelPrefix, test.step1.modelIdSuffix), memBytes, uint32(test.step1.modelVersion), test.step1.stepType, false, errors)

			wg.Wait()

			isErrorActual := false
			for i := 0; i <= repeats; i++ {
				err := <-errors
				if err != nil {
					t.Log(err)
				}
				isErrorActual = (err != nil) || isErrorActual
			}

			g.Expect(isErrorActual).To(Equal(test.isError))
			g.Expect(manager.modelVersions.numModels()).To(Equal(test.expectedNumModels))
			g.Expect(manager.availableMemoryBytes).To(Equal(test.expectedAvailableMemory))
		})
	}
}
