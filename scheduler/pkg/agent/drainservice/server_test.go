/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package drainservice

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/internal/testing_utils"
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

	t.Logf("Call endpoint")

	go func() {
		_, err := http.Get("http://localhost:" + strconv.Itoa(serverPort) + "/terminate")
		g.Expect(err).To(BeNil())
	}()

	drainer.WaitOnTrigger()    // this represents the agent waiting on a termination call from the endpoint
	drainer.SetSchedulerDone() // this represents the agent setting the process as done (after getting a response from the scheduler)

	g.Expect(drainer.triggered).To(BeTrue())

	t.Logf("Stop")
	err = drainer.Stop()
	g.Expect(err).To(BeNil())
	g.Expect(drainer.Ready()).To(BeFalse())
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
