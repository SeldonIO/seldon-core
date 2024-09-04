package utils

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestCheckName(t *testing.T) {
	g := NewGomegaWithT(t)
	tests := []struct {
		name           string
		value          string
		expectedResult bool
	}{
		{
			"a valid name",
			"this-NAME_is_valid1234-5",
			true,
		},
		{
			"a valid numerical name",
			"111111111111111111",
			true,
		},
		{
			"an valid dash and underscore name that begins and ends with something alphanumeric",
			"1-----______A",
			true,
		},

		{
			"an invalid dash and underscore name that doesn't begin and end with something alphanumeric",
			"-----______",
			false,
		},
		{
			"an invalid dash and underscore name that doesn't end with something alphanumeric",
			"A-----______",
			false,
		},
		{
			"an invalid dash and underscore name that doesn't begin with something alphanumeric",
			"-----______1",
			false,
		},
		{
			"an invalid name with a dot",
			"this-Is-some-NA.ME_that_is_valid1234-5",
			false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g.Expect(CheckName(test.value)).To(BeIdenticalTo(test.expectedResult))
		})
	}
}
