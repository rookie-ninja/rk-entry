// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package rkentry

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	opts = prometheus.SummaryOpts{
		Namespace:  "namespace",
		Subsystem:  "subSystem",
		Name:       "name",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001, 0.999: 0.0001},
	}

	labelKeys   = []string{"restMethod", "resCode"}
	summaryVec  = prometheus.NewSummaryVec(opts, labelKeys)
	observer, _ = summaryVec.GetMetricWithLabelValues("GET", "200")
)

func TestNewPromMetricsInfo_HappyCase(t *testing.T) {
	observer.Observe(100)
	observer.Observe(200)
	metrics := NewPromMetricsInfo(summaryVec)
	assert.Len(t, metrics, 1)

	metric := metrics[0]

	assert.Equal(t, uint64(2), metric.Count)
	assert.Equal(t, "GET", metric.RestMethod)
	assert.Len(t, metric.ResCode, 1)
	assert.Equal(t, "200", metric.ResCode[0].ResCode)
	assert.Equal(t, uint64(2), metric.ResCode[0].Count)
}
