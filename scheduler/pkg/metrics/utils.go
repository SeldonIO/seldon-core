/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

func HttpCodeToString(code int) string {
	return fmt.Sprintf("%d", code)
}

func createCounterVec(counterName string, helperName string, labelNames []string) (*prometheus.CounterVec, error) {
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: counterName,
			Help: helperName,
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

func createGaugeVec(gaugeName string, helperName string, labelNames []string) (*prometheus.GaugeVec, error) {
	gauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: gaugeName,
			Help: helperName,
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
