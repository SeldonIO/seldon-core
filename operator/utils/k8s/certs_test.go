package k8s

import (
	. "github.com/onsi/gomega"
	"testing"
)

func TestGenerateCerts(t *testing.T) {
	g := NewGomegaWithT(t)

	hosts := []string{"seldon-webhook-service.seldon-system", "seldon-webhook-service.seldon-system.svc"}
	cert, err := certSetup(hosts)
	g.Expect(err).To(BeNil())
	g.Expect(cert.certificatePEM[0:27]).To(Equal("-----BEGIN CERTIFICATE-----"))
	g.Expect(cert.caPEM[0:27]).To(Equal("-----BEGIN CERTIFICATE-----"))
	g.Expect(cert.privKeyPEM[0:31]).To(Equal("-----BEGIN RSA PRIVATE KEY-----"))

}
