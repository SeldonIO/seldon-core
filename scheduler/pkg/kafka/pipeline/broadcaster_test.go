/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package pipeline

import (
	"testing"

	gm "github.com/onsi/gomega"
)

func TestSubscribe(t *testing.T) {
	t.Parallel()

	gm.RegisterFailHandler(func(message string, callerSkip ...int) {
		t.Helper()
		t.Fatal(message)
	})

	b := NewBroadcaster()

	ch1 := b.Subscribe()
	gm.Expect(ch1).ToNot(gm.BeNil())

	ch2 := b.Subscribe()
	ch3 := b.Subscribe()

	gm.Expect(ch2).ToNot(gm.BeNil())
	gm.Expect(ch3).ToNot(gm.BeNil())

	gm.Expect(ch1).To(gm.BeIdenticalTo(ch2))
	gm.Expect(ch1).To(gm.BeIdenticalTo(ch3))
	gm.Expect(ch2).To(gm.BeIdenticalTo(ch3))
}

func TestBroadcast(t *testing.T) {
	t.Parallel()

	gm.RegisterFailHandler(func(message string, callerSkip ...int) {
		t.Helper()
		t.Fatal(message)
	})

	b := NewBroadcaster()
	b.Broadcast()

	ch1 := b.Subscribe()
	ch2 := b.Subscribe()
	ch3 := b.Subscribe()

	b.Broadcast()

	channels := []<-chan struct{}{ch1, ch2, ch3}
	for i, ch := range channels {
		gm.Expect(ch).Should(gm.BeClosed(), "channel %d did not receive broadcast", i+1)
	}
}

func TestMultipleBroadcasts(t *testing.T) {
	t.Parallel()

	gm.RegisterFailHandler(func(message string, callerSkip ...int) {
		t.Helper()
		t.Fatal(message)
	})

	b := NewBroadcaster()

	ch1 := b.Subscribe()
	b.Broadcast()

	gm.Expect(ch1).Should(gm.BeClosed())
	b.Broadcast()

	ch2 := b.Subscribe()
	ch3 := b.Subscribe()

	gm.Expect(ch2).Should(gm.Not(gm.BeClosed()))
	gm.Expect(ch3).Should(gm.Not(gm.BeClosed()))

	b.Broadcast()

	gm.Expect(ch2).Should(gm.BeClosed())
	gm.Expect(ch3).Should(gm.BeClosed())
}
