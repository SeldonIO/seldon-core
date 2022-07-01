package cli

import (
	"net/http"
	"testing"

	. "github.com/onsi/gomega"
)

func TestSaveLoadSessionKey(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name          string
		headers       http.Header
		expectedSaved bool
		expectedKey   string
	}

	tests := []test{
		{
			name:          "ok",
			headers:       http.Header{SeldonRouteHeader: []string{"val"}},
			expectedSaved: true,
			expectedKey:   "val",
		},
		{
			name:          "ok2",
			headers:       http.Header{"foo": []string{"bar"}, SeldonRouteHeader: []string{"val"}},
			expectedSaved: true,
			expectedKey:   "val",
		},
		{
			name:          "no key",
			headers:       http.Header{"foo": []string{"bar"}},
			expectedSaved: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			saved, err := saveStickySessionKey(test.headers)
			g.Expect(err).To(BeNil())
			g.Expect(saved).To(Equal(test.expectedSaved))
			if saved {
				key, _ := getStickySessionKey()
				g.Expect(key).To(Equal(test.expectedKey))
			}
		})
	}
}
