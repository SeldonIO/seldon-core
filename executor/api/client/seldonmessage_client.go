package client

import (
	api "github.com/seldonio/seldon-core/executor/api/grpc"
)
func ExtractRoute(msg *api.SeldonMessage) []int {
	switch msg.GetData().DataOneof.(type) {
	case *api.DefaultData_Ndarray:
		values := msg.GetData().GetNdarray().GetValues()
		routeArr := make([]int,len(values))
		for i,value := range(values) {
			routeArr[i] = int(value.GetNumberValue())
		}
		return routeArr
	case *api.DefaultData_Tensor:
		values := msg.GetData().GetTensor().Values
		routeArr := make([]int,len(values))
		for i,value := range(values) {
			routeArr[i] = int(value)
		}
		return routeArr
	case *api.DefaultData_Tftensor:
		values := msg.GetData().GetTftensor().GetIntVal()
		routeArr := make([]int,len(values))
		for i,value := range(values) {
			routeArr[i] = int(value)
		}
		return routeArr
	}
	return []int{-1}
}