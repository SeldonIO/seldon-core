/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

// This package is responsible to synchronise starting up the different components of the "scheduler".
// In particular, it is responsible for making sure that the time between the scheduler starts and while
// the different model servers connect that the data plane (inferences) is not affected.
package synchroniser

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

func TestSimpleSynchroniser(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name    string
		timeout time.Duration
		signal  bool
	}

	tests := []test{
		{
			name:    "Simple",
			timeout: 100 * time.Millisecond,
			signal:  true,
		},
		{
			name:    "Longer timeout",
			timeout: 500 * time.Millisecond,
			signal:  true,
		},
		{
			name:    "Small timer",
			timeout: 1 * time.Millisecond,
			signal:  true,
		},
		{
			name:    "No signal",
			timeout: 100 * time.Millisecond,
			signal:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := NewSimpleSynchroniser(test.timeout)
			startTime := time.Now()
			g.Expect(s.IsReady()).To(BeFalse())
			if test.signal {
				s.Signals(1)
			}
			s.WaitReady()
			elapsed := time.Since(startTime)
			g.Expect(s.IsReady()).To(BeTrue())
			g.Expect(elapsed).To(BeNumerically(">", test.timeout))

			// this should have no effect
			s.Signals(100000)
			g.Expect(s.IsReady()).To(BeTrue())
			// make sure we are graceful after this point
			s.Signals(10)
			g.Expect(s.IsReady()).To(BeTrue())
			s.WaitReady()
			g.Expect(s.IsReady()).To(BeTrue())
		})
	}
}
