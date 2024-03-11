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

package modelscaling

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

func TestModelDelaysSimple(t *testing.T) {
	g := NewGomegaWithT(t)

	dummyModelBase := "dummy_model_0"
	t.Run("simple", func(t *testing.T) {

		requestId := "request_id_0"
		delays := NewModelReplicaDelaysKeeper()

		_, err := delays.Get(dummyModelBase)
		g.Expect(err).To(Not(BeNil()))

		delays.ModelInferEnter(dummyModelBase, requestId)
		time.Sleep(time.Millisecond * time.Duration(10))
		delays.ModelInferExit(dummyModelBase, requestId)

		delay, err := delays.Get(dummyModelBase)
		g.Expect(err).To(BeNil())
		g.Expect(delay > 0).To(BeTrue())
	})
}
