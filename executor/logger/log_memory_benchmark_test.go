package logger

import (
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
	"time"
)

/*
Benchmark for debugging high memory usage as reported in https://github.com/SeldonIO/seldon-core/issues/3726
The actual time to complete is irrelevant, the important information is in the profiles we can get while running it.
It simulates the executor getting a lot of requests and trying to write to a slow request logger.
To run you should limit the number of executions like:
go test -bench=. -run=^$ . -benchtime=10000x
The profile endpoint will be printed to the screen.
*/

var processedChan = make(chan bool, 100)

func BenchmarkLoggerMemoryUsage(b *testing.B) {
	serverPort := startSlowLogListener()

	err := StartDispatcher(5, DefaultWorkQueueSize, DefaultWriteTimeoutMilliseconds, logf.Log.WithName("test"), "test-name", "test-namespace", "test-predictor", "", "")
	if err != nil {
		b.Fatal(err)
	}
	logURL, err := url.Parse(fmt.Sprintf("http://localhost:%v/log", serverPort))
	if err != nil {
		b.Fatal(err)
	}

	go logNTimes(b.N, b, logURL)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		<-processedChan
	}
}

func logNTimes(n int, b *testing.B, url *url.URL) {
	testMessage := []byte("test message")
	for i := 0; i < n; i++ {
		err := QueueLogRequest(LogRequest{
			Url:         url,
			Bytes:       &testMessage,
			ContentType: "test",
			ReqType:     InferenceRequest,
			Id:          "test",
			SourceUri:   url,
			ModelId:     "1",
			RequestId:   "1",
		})
		if err != nil {
			b.Log(err)
		}
	}
}

func startSlowLogListener() int {
	// Use a random free port by specifying port :0
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}

	port := listener.Addr().(*net.TCPAddr).Port

	fmt.Printf("Test server is listening on :%v\n", port)
	fmt.Printf("Get goroutine profile by running:\n go tool pprof http://localhost:%v/debug/pprof/goroutine\n", port)

	mux := http.NewServeMux()

	mux.HandleFunc("/log", func(w http.ResponseWriter, r *http.Request) {
		//fmt.Println("received a message to the /log endpoint")
		time.Sleep(2 * time.Second)
		processedChan <- true
		fmt.Fprint(w, "logged successfully")
	})
	mux.Handle("/debug/pprof/", http.DefaultServeMux)

	go func() {
		panic(http.Serve(listener, mux))
	}()

	return port
}
