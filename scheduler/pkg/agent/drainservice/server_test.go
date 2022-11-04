package drainservice

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

// TODO: refactor as this is a copy from another test
func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func TestDrainerServiceSmoke(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := log.New()
	logger.SetLevel(log.DebugLevel)

	serverPort, err := getFreePort()
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
	numCalls := 5
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

	t.Logf("Stop")
	err = drainer.Stop()
	g.Expect(err).To(BeNil())
	g.Expect(drainer.Ready()).To(BeFalse())

}

func TestDrainerServiceEarlyStop(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := log.New()
	logger.SetLevel(log.DebugLevel)

	serverPort, err := getFreePort()
	if err != nil {
		t.Fatal(err)
	}
	drainer := NewDrainerService(logger, uint(serverPort))

	err = drainer.Stop()
	g.Expect(err).To(BeNil())

	time.Sleep(time.Millisecond * 100)
	g.Expect(drainer.Ready()).To(BeFalse())
}
