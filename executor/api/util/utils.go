package util

import (
	"os"
	"strconv"

	"github.com/seldonio/seldon-core/executor/api/grpc/seldon/proto"
	"os"
)

// Get an environment variable given by key or return the fallback.
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func ExtractRouteFromSeldonMessage(msg *proto.SeldonMessage) []int {
	switch msg.GetData().DataOneof.(type) {
	case *proto.DefaultData_Ndarray:
		values := msg.GetData().GetNdarray().GetValues()
		routeArr := make([]int, len(values))
		for i, value := range values {
			if listValue := value.GetListValue(); listValue != nil {
				routeArr[i] = int(listValue.GetValues()[0].GetNumberValue())
			} else {
				routeArr[i] = int(value.GetNumberValue())
			}
		}
		return routeArr
	case *proto.DefaultData_Tensor:
		values := msg.GetData().GetTensor().Values
		routeArr := make([]int, len(values))
		for i, value := range values {
			routeArr[i] = int(value)
		}
		return routeArr
	case *proto.DefaultData_Tftensor:
		values := msg.GetData().GetTftensor().GetIntVal()
		routeArr := make([]int, len(values))
		for i, value := range values {
			routeArr[i] = int(value)
		}
		return routeArr
	}
	return []int{-1}
}

// Get an environment variable given by key or return the fallback.
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// Get an environment variable given by key or return the fallback.
func GetEnvAsBool(key string, fallback bool) bool {
	if raw, ok := os.LookupEnv(key); ok {
		val, err := strconv.ParseBool(raw)
		if err == nil {
			return val
		}
	}

	return fallback
}
