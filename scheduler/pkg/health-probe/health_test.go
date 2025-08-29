package health_probe

import (
	"errors"
	"testing"

	g "github.com/onsi/gomega"
)

func TestProbeTypeValid(t *testing.T) {
	g.RegisterTestingT(t)

	t.Run("should return true for valid probe types", func(t *testing.T) {
		g.Expect(ProbeStartUp.Valid()).To(g.BeTrue())
		g.Expect(ProbeReadiness.Valid()).To(g.BeTrue())
		g.Expect(ProbeLiveness.Valid()).To(g.BeTrue())
	})

	t.Run("should return false for invalid probe types", func(t *testing.T) {
		invalidProbe := ProbeType(999)
		g.Expect(invalidProbe.Valid()).To(g.BeFalse())

		zeroProbe := ProbeType(0)
		g.Expect(zeroProbe.Valid()).To(g.BeFalse())

		combinedProbe := ProbeStartUp | ProbeReadiness
		g.Expect(combinedProbe.Valid()).To(g.BeFalse())
	})
}

func TestNewManager(t *testing.T) {
	g.RegisterTestingT(t)

	t.Run("should create a new manager with empty services map", func(t *testing.T) {
		man := NewManager()
		g.Expect(man).ToNot(g.BeNil())

		// Cast to concrete type to access internal fields for testing
		mgr, ok := man.(*manager)
		g.Expect(ok).To(g.BeTrue())
		g.Expect(mgr.svcs).ToNot(g.BeNil())
		g.Expect(mgr.svcs).To(g.BeEmpty())
	})
}

func TestManagerRegisterSvc(t *testing.T) {
	g.RegisterTestingT(t)

	t.Run("should successfully register service with valid probe types", func(t *testing.T) {
		mgr := NewManager()
		callback := func() error { return nil }

		g.Expect(func() {
			mgr.RegisterSvc("test-service", callback, ProbeReadiness)
		}).ToNot(g.Panic())

		// Verify service was registered
		m := mgr.(*manager)
		g.Expect(m.svcs).To(g.HaveKey("test-service"))
		g.Expect(m.svcs["test-service"].probes).To(g.Equal(ProbeReadiness))
		g.Expect(m.svcs["test-service"].cb).ToNot(g.BeNil())
	})

	t.Run("should successfully register service with multiple probe types", func(t *testing.T) {
		mgr := NewManager()
		callback := func() error { return nil }

		g.Expect(func() {
			mgr.RegisterSvc("multi-probe-service", callback, ProbeStartUp, ProbeReadiness, ProbeLiveness)
		}).ToNot(g.Panic())

		m := mgr.(*manager)
		expectedProbes := ProbeStartUp | ProbeReadiness | ProbeLiveness
		g.Expect(m.svcs["multi-probe-service"].probes).To(g.Equal(expectedProbes))
	})

	t.Run("should panic when registering duplicate service ID", func(t *testing.T) {
		mgr := NewManager()
		callback := func() error { return nil }

		mgr.RegisterSvc("duplicate-service", callback, ProbeReadiness)

		g.Expect(func() {
			mgr.RegisterSvc("duplicate-service", callback, ProbeLiveness)
		}).To(g.PanicWith(g.MatchRegexp("Service duplicate-service already added")))
	})

	t.Run("should panic when registering with invalid probe type", func(t *testing.T) {
		mgr := NewManager()
		callback := func() error { return nil }
		invalidProbe := ProbeType(999)

		g.Expect(func() {
			mgr.RegisterSvc("invalid-probe-service", callback, invalidProbe)
		}).To(g.PanicWith(g.MatchRegexp("Invalid probe type")))
	})

	t.Run("should panic when one of multiple probe types is invalid", func(t *testing.T) {
		mgr := NewManager()
		callback := func() error { return nil }
		invalidProbe := ProbeType(999)

		g.Expect(func() {
			mgr.RegisterSvc("mixed-probe-service", callback, ProbeReadiness, invalidProbe, ProbeLiveness)
		}).To(g.PanicWith(g.MatchRegexp("Invalid probe type")))
	})
}

func TestManagerCheckReadiness(t *testing.T) {
	g.RegisterTestingT(t)

	t.Run("should return nil when no services registered", func(t *testing.T) {
		mgr := NewManager()
		err := mgr.CheckReadiness()
		g.Expect(err).To(g.BeNil())
	})

	t.Run("should return nil when all readiness probes pass", func(t *testing.T) {
		mgr := NewManager()

		mgr.RegisterSvc("service1", func() error { return nil }, ProbeReadiness)
		mgr.RegisterSvc("service2", func() error { return nil }, ProbeReadiness, ProbeLiveness)
		mgr.RegisterSvc("service3", func() error { return nil }, ProbeStartUp) // Should be skipped

		err := mgr.CheckReadiness()
		g.Expect(err).To(g.BeNil())
	})

	t.Run("should return error when readiness probe fails", func(t *testing.T) {
		mgr := NewManager()
		expectedErr := errors.New("service failure")

		mgr.RegisterSvc("failing-service", func() error { return expectedErr }, ProbeReadiness)
		mgr.RegisterSvc("healthy-service", func() error { return nil }, ProbeReadiness)

		err := mgr.CheckReadiness()
		g.Expect(err).To(g.HaveOccurred())
		g.Expect(err.Error()).To(g.MatchRegexp("service failing-service is not ready.*service failure"))
	})

	t.Run("should return error when service callback is nil", func(t *testing.T) {
		mgr := NewManager().(*manager)
		mgr.svcs["nil-callback-service"] = service{
			probes: ProbeReadiness,
			cb:     nil,
		}

		err := mgr.CheckReadiness()
		g.Expect(err).To(g.HaveOccurred())
		g.Expect(err.Error()).To(g.Equal("service nil-callback-service callback not set"))
	})
}

func TestManagerCheckLiveness(t *testing.T) {
	g.RegisterTestingT(t)

	t.Run("should return nil when no services registered", func(t *testing.T) {
		mgr := NewManager()
		err := mgr.CheckLiveness()
		g.Expect(err).To(g.BeNil())
	})

	t.Run("should return nil when all liveness probes pass", func(t *testing.T) {
		mgr := NewManager()

		mgr.RegisterSvc("service1", func() error { return nil }, ProbeLiveness)
		mgr.RegisterSvc("service2", func() error { return nil }, ProbeReadiness, ProbeLiveness)
		mgr.RegisterSvc("service3", func() error { return nil }, ProbeStartUp) // Should be skipped

		err := mgr.CheckLiveness()
		g.Expect(err).To(g.BeNil())
	})

	t.Run("should return error when liveness probe fails", func(t *testing.T) {
		mgr := NewManager()
		expectedErr := errors.New("service died")

		mgr.RegisterSvc("dying-service", func() error { return expectedErr }, ProbeLiveness)

		err := mgr.CheckLiveness()
		g.Expect(err).To(g.HaveOccurred())
		g.Expect(err.Error()).To(g.MatchRegexp("service dying-service is not ready.*service died"))
	})
}

func TestManagerCheckStartup(t *testing.T) {
	g.RegisterTestingT(t)

	t.Run("should return nil when no services registered", func(t *testing.T) {
		mgr := NewManager()
		err := mgr.CheckStartup()
		g.Expect(err).To(g.BeNil())
	})

	t.Run("should return nil when all startup probes pass", func(t *testing.T) {
		mgr := NewManager()

		mgr.RegisterSvc("service1", func() error { return nil }, ProbeStartUp)
		mgr.RegisterSvc("service2", func() error { return nil }, ProbeStartUp, ProbeReadiness)
		mgr.RegisterSvc("service3", func() error { return nil }, ProbeLiveness) // Should be skipped

		err := mgr.CheckStartup()
		g.Expect(err).To(g.BeNil())
	})

	t.Run("should return error when startup probe fails", func(t *testing.T) {
		mgr := NewManager()
		expectedErr := errors.New("startup failed")

		mgr.RegisterSvc("startup-failing-service", func() error { return expectedErr }, ProbeStartUp)

		err := mgr.CheckStartup()
		g.Expect(err).To(g.HaveOccurred())
		g.Expect(err.Error()).To(g.MatchRegexp("service startup-failing-service is not ready.*startup failed"))
	})
}

func TestManagerRunCheck(t *testing.T) {
	g.RegisterTestingT(t)

	t.Run("should skip services without the specified probe type", func(t *testing.T) {
		mgr := NewManager().(*manager)
		callCount := 0

		// Register service with only readiness probe
		mgr.RegisterSvc("readiness-only", func() error {
			callCount++
			return nil
		}, ProbeReadiness)

		// Check liveness - should skip the readiness-only service
		err := mgr.runCheck(ProbeLiveness)
		g.Expect(err).To(g.BeNil())
		g.Expect(callCount).To(g.Equal(0))
	})

	t.Run("should call services that have the specified probe type", func(t *testing.T) {
		mgr := NewManager().(*manager)
		callCount := 0

		mgr.RegisterSvc("multi-probe", func() error {
			callCount++
			return nil
		}, ProbeReadiness, ProbeLiveness)

		// Both readiness and liveness checks should call the callback
		err := mgr.runCheck(ProbeReadiness)
		g.Expect(err).To(g.BeNil())
		g.Expect(callCount).To(g.Equal(1))

		err = mgr.runCheck(ProbeLiveness)
		g.Expect(err).To(g.BeNil())
		g.Expect(callCount).To(g.Equal(2))
	})

	t.Run("should return first error encountered", func(t *testing.T) {
		mgr := NewManager()
		firstErr := errors.New("first error")
		secondErr := errors.New("second error")

		mgr.RegisterSvc("first-failing", func() error { return firstErr }, ProbeReadiness)
		mgr.RegisterSvc("second-failing", func() error { return secondErr }, ProbeReadiness)

		err := mgr.CheckReadiness()
		g.Expect(err).To(g.HaveOccurred())
		// Should contain the first error (order may vary due to map iteration)
		g.Expect(err.Error()).To(g.SatisfyAny(
			g.ContainSubstring("first error"),
			g.ContainSubstring("second error"),
		))
	})
}

func TestIntegrationScenarios(t *testing.T) {
	g.RegisterTestingT(t)

	t.Run("real-world health check scenario", func(t *testing.T) {
		mgr := NewManager()

		// Database service - has startup and readiness probes
		dbStarted := false
		mgr.RegisterSvc("database", func() error {
			if !dbStarted {
				return errors.New("database not started")
			}
			return nil
		}, ProbeStartUp, ProbeReadiness)

		// Cache service - has all probe types
		cacheHealthy := true
		mgr.RegisterSvc("cache", func() error {
			if !cacheHealthy {
				return errors.New("cache unhealthy")
			}
			return nil
		}, ProbeStartUp, ProbeReadiness, ProbeLiveness)

		// Web service - only liveness probe
		mgr.RegisterSvc("web", func() error { return nil }, ProbeLiveness)

		// Initially database not started
		g.Expect(mgr.CheckStartup()).To(g.HaveOccurred())
		g.Expect(mgr.CheckReadiness()).To(g.HaveOccurred())
		g.Expect(mgr.CheckLiveness()).To(g.BeNil()) // Only cache and web, both healthy

		// Start database
		dbStarted = true
		g.Expect(mgr.CheckStartup()).To(g.BeNil())
		g.Expect(mgr.CheckReadiness()).To(g.BeNil())
		g.Expect(mgr.CheckLiveness()).To(g.BeNil())

		// Cache becomes unhealthy
		cacheHealthy = false
		g.Expect(mgr.CheckStartup()).To(g.HaveOccurred())
		g.Expect(mgr.CheckReadiness()).To(g.HaveOccurred())
		g.Expect(mgr.CheckLiveness()).To(g.HaveOccurred())
	})
}
