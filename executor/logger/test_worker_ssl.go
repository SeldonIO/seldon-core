package logger

import (
	"fmt"
	"os"

	"net/http"

	_ "net/http/pprof"
	"testing"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func testLoggerKafkaSSL(t *testing.T) {
	os.Setenv("ENV_LOGGER_KAFKA_BROKER", "test-broker")
	os.Setenv("KAFKA_SECURITY_PROTOCOL", "ssl")
	os.Setenv("KAFKA_SSL_CA_CERT_FILE", "test_ca.pem")
	os.Setenv("KAFKA_SSL_CLIENT_CERT_FILE", "test_ca.cert")
	os.Setenv("KAFKA_SSL_CLIENT_KEY_FILE", "test_ca.key")

	defer os.Unsetenv("ENV_LOGGER_KAFKA_BROKER")
	defer os.Unsetenv("KAFKA_SECURITY_PROTOCOL")
	defer os.Unsetenv("KAFKA_SSL_CA_CERT_FILE")
	defer os.Unsetenv("KAFKA_SSL_CLIENT_CERT_FILE")
	defer os.Unsetenv("KAFKA_SSL_CLIENT_KEY_FILE")

	err := StartDispatcher(2, DefaultWorkQueueSize, DefaultWriteTimeoutMilliseconds, logf.Log.WithName("test"), "test-name", "test-namespace", "test-predictor", "test-broker", "")
	if err != nil {
		t.Fatal(err)
	}

	// How can I create a mux(?) thingy but for Kafka and not http?!
	mux := http.NewServeMux()

	mux.HandleFunc("/log", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("received a message to the /log endpoint")
		fmt.Fprint(w, "logged successfully")
	})
	mux.Handle("/debug/pprof/", http.DefaultServeMux)
	// logKafka, err :=
	// err := LogRequest(LogRequest{
	// 	Url:         url,
	// 	Bytes:       &testMessage,
	// 	ContentType: "test",
	// 	ReqType:     InferenceRequest,
	// 	Id:          "test",
	// 	SourceUri:   url,
	// 	ModelId:     "1",
	// 	RequestId:   "1",
	// })
	// if err != nil {
	// 	t.Log(err)
	// }

}
