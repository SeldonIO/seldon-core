/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package modelscaling

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestModelLagsSimple(t *testing.T) {
	g := NewGomegaWithT(t)
	t.Logf("Start!")

	lag := &lagStats{
		lag: 0,
	}

	lag.Enter("")
	g.Expect(lag.Get()).To(Equal(uint32(1)))
	lag.Exit("")
	g.Expect(lag.Get()).To(Equal(uint32(0)))
	lag.Enter("")
	lag.Enter("")
	g.Expect(lag.Get()).To(Equal(uint32(2)))
	lag.Reset()
	g.Expect(lag.Get()).To(Equal(uint32(0)))

	t.Logf("Done!")
}
