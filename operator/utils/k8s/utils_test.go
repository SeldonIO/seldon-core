package k8s

import (
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/version"
	discovery "k8s.io/client-go/discovery/fake"
	coretesting "k8s.io/client-go/testing"
	ctrl "sigs.k8s.io/controller-runtime"
	"testing"
)

func TestServerVersion(t *testing.T) {
	g := NewGomegaWithT(t)

	scenarios := map[string]struct {
		major    string
		minor    string
		expected string
	}{
		"1.18 test":             {major: "1", minor: "18", expected: "1.18"},
		"GKE 1.20":              {major: "1", minor: "20+", expected: "1.20"},
		"Unknown minor version": {major: "1", minor: "XYZ", expected: "1.18"},
	}

	for _, scenario := range scenarios {
		discoveryFake := &discovery.FakeDiscovery{Fake: &coretesting.Fake{},
			FakedServerVersion: &version.Info{
				Major: scenario.major,
				Minor: scenario.minor,
			}}
		serverVersion, err := GetServerVersion(discoveryFake, ctrl.Log)
		g.Expect(err).To(BeNil())
		g.Expect(serverVersion).To(Equal(scenario.expected))
	}

}
