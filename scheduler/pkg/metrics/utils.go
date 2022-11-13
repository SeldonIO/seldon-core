/*
Copyright 2022 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
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
