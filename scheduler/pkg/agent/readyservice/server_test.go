/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package readyservice

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/interfaces"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/internal/testing_utils"
)

type FakeReadinessCheckTarget struct {
	isReady bool
}

func (f *FakeReadinessCheckTarget) Ready() bool {
	return f.isReady
}

func TestReadyServiceSmoke(t *testing.T) {
	logger := log.New()
	logger.SetLevel(log.DebugLevel)
	g := NewGomegaWithT(t)

	type test struct {
		name               string
		desiredReadyStatus bool
		target             interfaces.ServiceWithReadinessCheck
	}
	tests := []test{
		{name: "CheckReady", desiredReadyStatus: true, target: &FakeReadinessCheckTarget{isReady: true}},
		{name: "CheckNotReady", desiredReadyStatus: false, target: &FakeReadinessCheckTarget{isReady: false}},
	}

	serverPort, err := testing_utils.GetFreePortForTest()
	if err != nil {
		t.Fatal(err)
	}
	ready := NewReadyService(logger, uint(serverPort))

	t.Logf("Start")
	err = ready.Start()
	g.Expect(err).To(BeNil())
	time.Sleep(time.Millisecond * 100)
	g.Expect(ready.Ready()).To(BeTrue())

	t.Logf("Call endpoints")
	for _, tc := range tests {
		// target dynamically updated while readiness service is running
		ready.SetState(tc.target)

		response, err := http.Get("http://localhost:" + strconv.Itoa(serverPort) + readyEndpoint)
		g.Expect(err).To(BeNil())
		if tc.desiredReadyStatus == true {
			g.Expect(response.StatusCode).To(Equal(http.StatusOK))
		} else {
			g.Expect(response.StatusCode).NotTo(Equal(http.StatusOK))
		}
	}
	t.Logf("Stop")
	err = ready.Stop()
	g.Expect(err).To(BeNil())
	g.Expect(ready.Ready()).To(BeFalse())

}

func TestReadyServiceFailWithoutTargetSet(t *testing.T) {
	logger := log.New()
	logger.SetLevel(log.DebugLevel)
	g := NewGomegaWithT(t)

	serverPort, err := testing_utils.GetFreePortForTest()
	if err != nil {
		t.Fatal(err)
	}
	ready := NewReadyService(logger, uint(serverPort))

	t.Logf("Start")
	err = ready.Start()
	g.Expect(err).To(BeNil())
	time.Sleep(time.Millisecond * 100)
	g.Expect(ready.Ready()).To(BeTrue())

	t.Logf("Call endpoints")
	response, err := http.Get("http://localhost:" + strconv.Itoa(serverPort) + readyEndpoint)
	g.Expect(err).To(BeNil())
	// Since no target service is set via ready.SetState(), the readiness check should fail
	g.Expect(response.StatusCode).NotTo(Equal(http.StatusOK))
	t.Logf("Stop")
	err = ready.Stop()
	g.Expect(err).To(BeNil())
	g.Expect(ready.Ready()).To(BeFalse())
}

func TestDrainerServiceEarlyStop(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := log.New()
	logger.SetLevel(log.DebugLevel)

	serverPort, err := testing_utils.GetFreePortForTest()
	if err != nil {
		t.Fatal(err)
	}
	ready := NewReadyService(logger, uint(serverPort))

	err = ready.Stop()
	g.Expect(err).To(BeNil())

	time.Sleep(time.Millisecond * 100)
	g.Expect(ready.Ready()).To(BeFalse())
}
