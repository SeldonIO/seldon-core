/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
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
