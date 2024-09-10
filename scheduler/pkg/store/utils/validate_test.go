/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

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
			"this-name-is-valid1234",
			true,
		},
		{
			"a valid numerical name",
			"111111111111111111",
			true,
		},
		{
			"a valid name that begins and ends with something alphanumeric",
			"a-a-a",
			true,
		},
		{
			"an invalid name that doesn't begin and end with something alphanumeric",
			"--",
			false,
		},
		{
			"an invalid name that doesn't end with something alphanumeric",
			"1--",
			false,
		},
		{
			"an invalid name that doesn't begin with something alphanumeric",
			"--a",
			false,
		},
		{
			"an invalid name with an uppercase letter",
			"this-name-is-not-Valid1234",
			false,
		},
		{
			"an invalid name with a dot",
			"not.valid",
			false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g.Expect(CheckName(test.value)).To(BeIdenticalTo(test.expectedResult))
		})
	}
}
