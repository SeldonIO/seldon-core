/*
Copyright 2023 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package testing_utils

import (
	"fmt"

	"github.com/jarcoal/httpmock"
	log "github.com/sirupsen/logrus"
)

func CreateTestV2ClientMockResponders(host string, port int, modelName string, status int, state *V2State) {

	httpmock.RegisterResponder("POST", fmt.Sprintf("http://%s:%d/v2/repository/models/%s/load", host, port, modelName),
		state.LoadResponder(modelName, status))
	httpmock.RegisterResponder("POST", fmt.Sprintf("http://%s:%d/v2/repository/models/%s/unload", host, port, modelName),
		state.UnloadResponder(modelName, status))
}

func CreateTestV2ClientwithState(models []string, status int) (*V2RestClientForTest, *V2State) {
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	host := "model-server"
	port := 8080
	v2 := NewV2RestClientForTest(host, port, logger)
	state := &V2State{
		Models: make(map[string]bool, len(models)),
	}

	for _, model := range models {
		CreateTestV2ClientMockResponders(host, port, model, status, state)
	}
	// we do not care about ready in tests
	httpmock.RegisterResponder("GET", fmt.Sprintf("http://%s:%d/v2/health/live", host, port),
		httpmock.NewStringResponder(200, `{}`))
	return v2, state
}

func CreateTestV2Client(models []string, status int) *V2RestClientForTest {
	v2, _ := CreateTestV2ClientwithState(models, status)
	return v2
}