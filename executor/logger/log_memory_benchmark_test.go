package logger

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
	"time"
)

func BenchmarkLoggerMemoryUsage(b *testing.B) {
	//fmt.Println("starting benchmark..")
	serverPort := startSlowLogListener()

	err := StartDispatcher(5, logf.Log.WithName("test"), "test-name", "test-namespace", "test-predictor", "", "")
	if err != nil {
		b.Fatal(err)
	}
	logURL, err := url.Parse(fmt.Sprintf("http://localhost:%v/log",serverPort))
	if err != nil {
		b.Fatal(err)
	}
	testMessage := []byte("test message")


	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		err = QueueLogRequest(LogRequest{
			Url:         logURL,
			Bytes:       &testMessage,
			ContentType: "test",
			ReqType:     InferenceRequest,
			Id:          "test",
			SourceUri:   logURL,
			ModelId:     "1",
			RequestId:   "1",
		})
		if err != nil {
			b.Log(err)
		}
	}
}
//
//func TestAsd(t *testing.T) {
//	serverPort := startSlowLogListener()
//
//	err := StartDispatcher(5, logf.Log.WithName("test"), "test-name", "test-namespace", "test-predictor", "", "")
//	if err != nil {
//		t.Fatal(err)
//	}
//
//
//	logURL, err := url.Parse(fmt.Sprintf("http://localhost:%v/log",serverPort))
//	if err != nil {
//		t.Fatal(err)
//	}
//	testMessage := []byte("test message")
//	err = QueueLogRequest(LogRequest{
//		Url:         logURL,
//		Bytes:       &testMessage,
//		ContentType: "test",
//		ReqType:     InferenceRequest,
//		Id:          "test",
//		SourceUri:   logURL,
//		ModelId:     "1",
//		RequestId:   "1",
//	})
//	if err != nil {
//		panic(err)
//	}
//}

func startSlowLogListener() int {
	// Use a random free port by specifying port :0
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}

	port := listener.Addr().(*net.TCPAddr).Port

	fmt.Printf("listening on %v\n", port)

	mux := http.NewServeMux()

	mux.HandleFunc("/log", func(w http.ResponseWriter, r *http.Request) {
		//fmt.Println("received a message to the /log endpoint")
		time.Sleep(2 * time.Second)
		fmt.Fprint(w, "logged successfully")
	})

	go func() {
		// todo: cleanup
		panic(http.Serve(listener, mux))
	}()

	return port
}