/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
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
