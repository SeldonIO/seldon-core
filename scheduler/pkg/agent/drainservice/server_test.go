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

package drainservice

import (
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/internal/testing_utils"
	log "github.com/sirupsen/logrus"
)

func TestDrainerServiceSmoke(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := log.New()
	logger.SetLevel(log.DebugLevel)

	serverPort, err := testing_utils.GetFreePortForTest()
	if err != nil {
		t.Fatal(err)
	}
	drainer := NewDrainerService(logger, uint(serverPort))

	t.Logf("Start")
	err = drainer.Start()
	g.Expect(err).To(BeNil())

	time.Sleep(time.Millisecond * 100)
	g.Expect(drainer.Ready()).To(BeTrue())

	t.Logf("Call endpoints")
	numCalls := 3
	wg := sync.WaitGroup{}
	wg.Add(numCalls)

	for i := 0; i < numCalls; i++ {
		go func() {
			// only one of these http call will trigger the start of the draining process
			_, err = http.Get("http://localhost:" + strconv.Itoa(serverPort) + terminateEndpoint)
			if err == nil {
				wg.Done()
			}
		}()

	}
	drainer.WaitOnTrigger()    // this represents the agent waiting on a termination call from the endpoint
	drainer.SetSchedulerDone() // this represents the agent setting the process as done (after getting a response from the scheduler)
	wg.Wait()
	g.Expect(drainer.triggered).To(BeTrue())

	t.Logf("Stop")
	err = drainer.Stop()
	g.Expect(err).To(BeNil())
	g.Expect(drainer.Ready()).To(BeFalse())

}

func TestDrainerServiceSmokeNoDraining(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := log.New()
	logger.SetLevel(log.DebugLevel)

	serverPort, err := testing_utils.GetFreePortForTest()
	if err != nil {
		t.Fatal(err)
	}
	drainer := NewDrainerService(logger, uint(serverPort))

	t.Logf("Start")
	err = drainer.Start()
	g.Expect(err).To(BeNil())

	time.Sleep(time.Millisecond * 100)
	g.Expect(drainer.Ready()).To(BeTrue())

	t.Logf("Call endpoints")

	response, err := http.Get("http://localhost:" + strconv.Itoa(serverPort) + terminateEndpoint)
	g.Expect(err).To(BeNil())
	g.Expect(response.StatusCode).To(Equal(http.StatusOK))
	g.Expect(drainer.triggered).To(BeFalse())
	g.Expect(drainer.events).To(Equal(uint32(0)))

	t.Logf("Stop")
	_ = drainer.Stop()

}

func TestDrainerServiceEarlyStop(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := log.New()
	logger.SetLevel(log.DebugLevel)

	serverPort, err := testing_utils.GetFreePortForTest()
	if err != nil {
		t.Fatal(err)
	}
	drainer := NewDrainerService(logger, uint(serverPort))

	err = drainer.Stop()
	g.Expect(err).To(BeNil())

	time.Sleep(time.Millisecond * 100)
	g.Expect(drainer.Ready()).To(BeFalse())
}
