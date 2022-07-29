package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

func HttpCodeToString(code int) string {
	return fmt.Sprintf("%d", code)
}

func createCounterVec(counterName, helperName, namespace string, labelNames []string) (*prometheus.CounterVec, error) {
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:      counterName,
			Namespace: namespace,
			Help:      helperName,
		},
		labelNames,
	)
	err := prometheus.Register(counter)
	if err != nil {
		if e, ok := err.(prometheus.AlreadyRegisteredError); ok {
			counter = e.ExistingCollector.(*prometheus.CounterVec)
		} else {
			return nil, err
		}
	}
	return counter, nil
}

func createGaugeVec(gaugeName, helperName, namespace string, labelNames []string) (*prometheus.GaugeVec, error) {
	gauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      gaugeName,
			Namespace: namespace,
			Help:      helperName,
		},
		labelNames,
	)
	err := prometheus.Register(gauge)
	if err != nil {
		if e, ok := err.(prometheus.AlreadyRegisteredError); ok {
			gauge = e.ExistingCollector.(*prometheus.GaugeVec)
		} else {
			return nil, err
		}
	}
	return gauge, nil
}
