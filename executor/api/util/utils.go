package util

import (
	"github.com/seldonio/seldon-core/executor/api/grpc/seldon/proto"
)

func ExtractRouteFromSeldonMessage(msg *proto.SeldonMessage) []int {
	switch msg.GetData().DataOneof.(type) {
	case *proto.DefaultData_Ndarray:
		values := msg.GetData().GetNdarray().GetValues()
		routeArr := make([]int, len(values))
		for i, value := range values {
			routeArr[i] = int(value.GetNumberValue())
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
