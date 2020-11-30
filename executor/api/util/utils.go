package util

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/golang/protobuf/jsonpb"
	"github.com/seldonio/seldon-core/executor/api/grpc/seldon/proto"
	"github.com/seldonio/seldon-core/executor/api/payload"
)

// Assumes the byte array is a json list of ints
func ExtractRouteAsJsonArray(msg []byte) ([]int, error) {
	var routes []int
	err := json.Unmarshal(msg, &routes)
	if err == nil {
		return routes, err
	} else {
		return nil, err
	}
}

func ExtractRouteFromSeldonJson(sp payload.SeldonPayload) (int, error) {
	var routes []int
	msg := sp.GetPayload().([]byte)

	var sm proto.SeldonMessage
	value := string(msg)
	err := jsonpb.UnmarshalString(value, &sm)
	if err == nil {
		//Remove in future
		routes = ExtractRouteFromSeldonMessage(&sm)
	} else {
		routes, err = ExtractRouteAsJsonArray(msg)
		if err != nil {
			return 0, err
		}
	}

	//Only returning first route. API could be extended to allow multiple routes
	return routes[0], nil
}

func RouteFromFeedbackJsonMeta(sp payload.SeldonPayload, predictorName string) int {
	msg := sp.GetPayload().([]byte)

	var fm proto.Feedback
	value := string(msg)
	err := jsonpb.UnmarshalString(value, &fm)
	if err != nil {
		return -1
	}

	return RouteFromFeedbackMessageMeta(&fm, predictorName)
}

func RouteFromFeedbackMessageMeta(msg *proto.Feedback, predictorName string) int {
	routing := msg.GetResponse().GetMeta().GetRouting()
	if route, ok := routing[predictorName]; ok {
		return int(route)
	}
	return -1
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
