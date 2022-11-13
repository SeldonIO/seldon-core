/*
Copyright 2022 Seldon Technologies Ltd.

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

package util

import (
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
)

func TestGetOriginalModelAndVersion(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name               string
		versionedModelName string
		originalModelName  string
		version            uint32
		err                error
	}

	tests := []test{
		{
			name:               "valid model with one separator",
			versionedModelName: "model_1",
			originalModelName:  "model",
			version:            1,
			err:                nil,
		},
		{
			name:               "valid model with more than one separator",
			versionedModelName: "model_x_1",
			originalModelName:  "model_x",
			version:            1,
			err:                nil,
		},
		{
			name:               "bad model no separator",
			versionedModelName: "model",
			originalModelName:  "",
			version:            0,
			err:                fmt.Errorf("cannot convert to original model"),
		},
		{
			name:               "bad model, version is not int",
			versionedModelName: "model_x",
			originalModelName:  "",
			version:            0,
			err:                fmt.Errorf("cannot convert to original model"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			orginalModelName, originalModelVersion, err := GetOrignalModelNameAndVersion(test.versionedModelName)
			g.Expect(orginalModelName).To(Equal(test.originalModelName))
			g.Expect(originalModelVersion).To(Equal(test.version))
			if test.err != nil {
				g.Expect(err).To(Equal(test.err))
			} else {
				g.Expect(err).To(BeNil())
			}
		})
	}

}
