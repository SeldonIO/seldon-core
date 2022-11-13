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

package cli

import (
	"net/http"
	"net/textproto"
	"testing"

	. "github.com/onsi/gomega"
)

func TestSaveLoadSessionKey(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name          string
		headers       http.Header
		expectedSaved bool
		expectedKeys  []string
	}

	tests := []test{
		{
			name:          "ok",
			headers:       http.Header{textproto.CanonicalMIMEHeaderKey(SeldonRouteHeader): []string{"val"}},
			expectedSaved: true,
			expectedKeys:  []string{"val"},
		},
		{
			name:          "multiple values",
			headers:       http.Header{textproto.CanonicalMIMEHeaderKey(SeldonRouteHeader): []string{"val", "val2"}},
			expectedSaved: true,
			expectedKeys:  []string{"val", "val2"},
		},
		{
			name:          "ok2",
			headers:       http.Header{"foo": []string{"bar"}, textproto.CanonicalMIMEHeaderKey(SeldonRouteHeader): []string{"val"}},
			expectedSaved: true,
			expectedKeys:  []string{"val"},
		},
		{
			name:          "no key",
			headers:       http.Header{"foo": []string{"bar"}},
			expectedSaved: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			saved, err := saveStickySessionKeyHttp(test.headers)
			g.Expect(err).To(BeNil())
			g.Expect(saved).To(Equal(test.expectedSaved))
			if saved {
				keys, _ := getStickySessionKeys()
				for _, key := range keys {
					found := false
					for _, keyExpected := range test.expectedKeys {
						if key == keyExpected {
							found = true
						}
					}
					g.Expect(found).To(BeTrue())
				}

			}
		})
	}
}
