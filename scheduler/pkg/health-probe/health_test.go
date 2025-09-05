/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package health_probe

import (
	"errors"
	"testing"

	g "github.com/onsi/gomega"
)

func TestProbeType_Valid(t *testing.T) {
	gomega := g.NewWithT(t)

	t.Run("valid probe types", func(t *testing.T) {
		gomega.Expect(ProbeStartUp.Valid()).To(g.BeTrue())
		gomega.Expect(ProbeReadiness.Valid()).To(g.BeTrue())
		gomega.Expect(ProbeLiveness.Valid()).To(g.BeTrue())
	})

	t.Run("invalid probe types", func(t *testing.T) {
		invalidProbe := ProbeType(999)
		gomega.Expect(invalidProbe.Valid()).To(g.BeFalse())

		zeroProbe := ProbeType(0)
		gomega.Expect(zeroProbe.Valid()).To(g.BeFalse())

		combinedProbe := ProbeStartUp | ProbeReadiness
		gomega.Expect(combinedProbe.Valid()).To(g.BeFalse())
	})
}

func TestManager_AddCheck(t *testing.T) {
	gomega := g.NewWithT(t)

	t.Run("successfully add check with single probe type", func(t *testing.T) {
		manager := NewManager()
		callback := func() error { return nil }

		gomega.Expect(func() {
			manager.AddCheck(callback, ProbeReadiness)
		}).ToNot(g.Panic())
	})

	t.Run("successfully add check with multiple probe types", func(t *testing.T) {
		manager := NewManager()
		callback := func() error { return nil }

		gomega.Expect(func() {
			manager.AddCheck(callback, ProbeReadiness, ProbeLiveness)
		}).ToNot(g.Panic())
	})

	t.Run("successfully add check with all probe types", func(t *testing.T) {
		manager := NewManager()
		callback := func() error { return nil }

		gomega.Expect(func() {
			manager.AddCheck(callback, ProbeStartUp, ProbeReadiness, ProbeLiveness)
		}).ToNot(g.Panic())
	})

	t.Run("panic on nil callback", func(t *testing.T) {
		manager := NewManager()

		gomega.Expect(func() {
			manager.AddCheck(nil, ProbeReadiness)
		}).To(g.PanicWith("nil callback"))
	})

	t.Run("panic on invalid probe type", func(t *testing.T) {
		manager := NewManager()
		callback := func() error { return nil }
		invalidProbe := ProbeType(999)

		gomega.Expect(func() {
			manager.AddCheck(callback, invalidProbe)
		}).To(g.PanicWith("Invalid probe type 999"))
	})

	t.Run("panic on mix of valid and invalid probe types", func(t *testing.T) {
		manager := NewManager()
		callback := func() error { return nil }
		invalidProbe := ProbeType(999)

		gomega.Expect(func() {
			manager.AddCheck(callback, ProbeReadiness, invalidProbe)
		}).To(g.PanicWith("Invalid probe type 999"))
	})
}

func TestManager_CheckReadiness(t *testing.T) {
	gomega := g.NewWithT(t)

	t.Run("returns nil when no checks are added", func(t *testing.T) {
		manager := NewManager()
		err := manager.CheckReadiness()
		gomega.Expect(err).To(g.BeNil())
	})

	t.Run("returns nil when no readiness checks are added", func(t *testing.T) {
		manager := NewManager()
		manager.AddCheck(func() error { return nil }, ProbeLiveness)

		err := manager.CheckReadiness()
		gomega.Expect(err).To(g.BeNil())
	})

	t.Run("returns nil when readiness check succeeds", func(t *testing.T) {
		manager := NewManager()
		manager.AddCheck(func() error { return nil }, ProbeReadiness)

		err := manager.CheckReadiness()
		gomega.Expect(err).To(g.BeNil())
	})

	t.Run("returns error when readiness check fails", func(t *testing.T) {
		manager := NewManager()
		expectedError := errors.New("service unavailable")
		manager.AddCheck(func() error { return expectedError }, ProbeReadiness)

		err := manager.CheckReadiness()
		gomega.Expect(err).To(g.HaveOccurred())
		gomega.Expect(err.Error()).To(g.ContainSubstring("failed probe: "))
	})

	t.Run("runs multiple readiness checks successfully", func(t *testing.T) {
		manager := NewManager()
		manager.AddCheck(func() error { return nil }, ProbeReadiness)
		manager.AddCheck(func() error { return nil }, ProbeReadiness)

		err := manager.CheckReadiness()
		gomega.Expect(err).To(g.BeNil())
	})

	t.Run("fails on first failing readiness check", func(t *testing.T) {
		manager := NewManager()
		manager.AddCheck(func() error { return errors.New("first error") }, ProbeReadiness)
		manager.AddCheck(func() error { return errors.New("second error") }, ProbeReadiness)

		err := manager.CheckReadiness()
		gomega.Expect(err).To(g.HaveOccurred())
		gomega.Expect(err.Error()).To(g.ContainSubstring("failed probe: "))
	})

	t.Run("runs only readiness checks when multiple probe types exist", func(t *testing.T) {
		manager := NewManager()
		readinessCalled := false
		livenessCalled := false

		manager.AddCheck(func() error {
			readinessCalled = true
			return nil
		}, ProbeReadiness)

		manager.AddCheck(func() error {
			livenessCalled = true
			return nil
		}, ProbeLiveness)

		err := manager.CheckReadiness()
		gomega.Expect(err).To(g.BeNil())
		gomega.Expect(readinessCalled).To(g.BeTrue())
		gomega.Expect(livenessCalled).To(g.BeFalse())
	})
}

func TestManager_CheckLiveness(t *testing.T) {
	gomega := g.NewWithT(t)

	t.Run("returns nil when no checks are added", func(t *testing.T) {
		manager := NewManager()
		err := manager.CheckLiveness()
		gomega.Expect(err).To(g.BeNil())
	})

	t.Run("returns nil when no liveness checks are added", func(t *testing.T) {
		manager := NewManager()
		manager.AddCheck(func() error { return nil }, ProbeReadiness)

		err := manager.CheckLiveness()
		gomega.Expect(err).To(g.BeNil())
	})

	t.Run("returns nil when liveness check succeeds", func(t *testing.T) {
		manager := NewManager()
		manager.AddCheck(func() error { return nil }, ProbeLiveness)

		err := manager.CheckLiveness()
		gomega.Expect(err).To(g.BeNil())
	})

	t.Run("returns error when liveness check fails", func(t *testing.T) {
		manager := NewManager()
		expectedError := errors.New("service dead")
		manager.AddCheck(func() error { return expectedError }, ProbeLiveness)

		err := manager.CheckLiveness()
		gomega.Expect(err).To(g.HaveOccurred())
		gomega.Expect(err.Error()).To(g.ContainSubstring("failed probe: "))
	})
}

func TestManager_CheckStartup(t *testing.T) {
	gomega := g.NewWithT(t)

	t.Run("returns nil when no checks are added", func(t *testing.T) {
		manager := NewManager()
		err := manager.CheckStartup()
		gomega.Expect(err).To(g.BeNil())
	})

	t.Run("returns nil when no startup checks are added", func(t *testing.T) {
		manager := NewManager()
		manager.AddCheck(func() error { return nil }, ProbeReadiness)

		err := manager.CheckStartup()
		gomega.Expect(err).To(g.BeNil())
	})

	t.Run("returns nil when startup check succeeds", func(t *testing.T) {
		manager := NewManager()
		manager.AddCheck(func() error { return nil }, ProbeStartUp)

		err := manager.CheckStartup()
		gomega.Expect(err).To(g.BeNil())
	})

	t.Run("returns error when startup check fails", func(t *testing.T) {
		manager := NewManager()
		expectedError := errors.New("startup failed")
		manager.AddCheck(func() error { return expectedError }, ProbeStartUp)

		err := manager.CheckStartup()
		gomega.Expect(err).To(g.HaveOccurred())
		gomega.Expect(err.Error()).To(g.ContainSubstring("failed probe: "))
	})
}

func TestManager_MultipleProbeTypes(t *testing.T) {
	gomega := g.NewWithT(t)

	t.Run("service with multiple probe types responds to all checks", func(t *testing.T) {
		manager := NewManager()
		callCount := 0
		manager.AddCheck(func() error {
			callCount++
			return nil
		}, ProbeReadiness, ProbeLiveness, ProbeStartUp)

		err := manager.CheckReadiness()
		gomega.Expect(err).To(g.BeNil())
		gomega.Expect(callCount).To(g.Equal(1))

		err = manager.CheckLiveness()
		gomega.Expect(err).To(g.BeNil())
		gomega.Expect(callCount).To(g.Equal(2))

		err = manager.CheckStartup()
		gomega.Expect(err).To(g.BeNil())
		gomega.Expect(callCount).To(g.Equal(3))
	})

	t.Run("service with partial probe types responds only to matching checks", func(t *testing.T) {
		manager := NewManager()
		callCount := 0
		manager.AddCheck(func() error {
			callCount++
			return nil
		}, ProbeReadiness, ProbeLiveness)

		err := manager.CheckReadiness()
		gomega.Expect(err).To(g.BeNil())
		gomega.Expect(callCount).To(g.Equal(1))

		err = manager.CheckLiveness()
		gomega.Expect(err).To(g.BeNil())
		gomega.Expect(callCount).To(g.Equal(2))

		err = manager.CheckStartup()
		gomega.Expect(err).To(g.BeNil())
		gomega.Expect(callCount).To(g.Equal(2))
	})
}
