package password

import (
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"testing"
)

func TestNewPasswordFolderHandler(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name string
		envs map[string]string
		err  bool
	}

	prefix := "x"
	suffix := "_y"
	tests := []test{
		{
			name: "ok",
			envs: map[string]string{
				prefix + suffix: "/abc",
			},
			err: false,
		},
		{
			name: "no env",
			envs: map[string]string{},
			err:  true,
		},
	}
	for _, test := range tests {
		logger := log.New()
		t.Run(test.name, func(t *testing.T) {
			for k, v := range test.envs {
				t.Setenv(k, v)
			}
			_, err := NewPasswordFolderHandler(prefix, suffix, logger)
			if test.err {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
			}
		})

	}
}
