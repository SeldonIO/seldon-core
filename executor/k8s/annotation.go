package k8s

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

const (
	ANNOTATIONS_FILE                 = "/etc/podinfo/annotations"
	ANNOTATION_GRPC_MAX_MESSAGE_SIZE = "seldon.io/grpc-max-message-size"
	ANNOTATION_GRPC_TIMEOUT          = "seldon.io/grpc-timeout"
	ANNOTATION_REST_TIMEOUT          = "seldon.io/rest-timeout"
)

func trimQuotes(v string) string {
	if strings.HasPrefix(v, "\"") {
		v = strings.TrimPrefix(v, "\"")
	}
	if strings.HasSuffix(v, "\"") {
		v = strings.TrimSuffix(v, "\"")
	}
	return v
}

func GetAnnotations() (map[string]string, error) {
	file, err := os.Open(ANNOTATIONS_FILE)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	return getAnnotationMap(string(b))
}

func getAnnotationMap(data string) (map[string]string, error) {
	annotationMap := make(map[string]string)
	for _, line := range strings.Split(data, "\n") {
		if line == "" {
			continue
		}
		kv := strings.Split(line, "=")
		if len(kv) != 2 {
			return nil, fmt.Errorf("Invalid annotation %s", line)
		}
		k := trimQuotes(kv[0])
		v := trimQuotes(kv[1])
		annotationMap[k] = v
	}
	return annotationMap, nil
}
