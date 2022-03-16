// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkmidprom has a couple of utility functions to start prometheus and pushgateway client locally.
package rkmidprom

import (
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"strings"
	"sync"
)

const (
	maxKeyLength     = 256
	separator        = "::"
	namespaceDefault = "rk"
	subSystemDefault = "svc"
)

// SummaryObjectives will track quantile of P50, P90, P99, P9999 by default.
var SummaryObjectives = map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001, 0.999: 0.0001}

// MetricsSet is a collections of counter, gauge, summary, histogram and link to certain registerer.
// User need to provide own prometheus.Registerer.
//
// 1: namespace:  the namespace of prometheus metrics
// 2: sysSystem:  the subSystem of prometheus metrics
// 3: keys:       a map stores all keys
// 4: counters:   map of counters
// 5: gauges:     map of gauges
// 6: summaries:  map of summaries
// 7: histograms: map of histograms
// 8: lock:       lock for thread safety
// 9: registerer  prometheus.Registerer
type MetricsSet struct {
	namespace  string
	subSystem  string
	keys       map[string]bool
	counters   map[string]*prometheus.CounterVec
	gauges     map[string]*prometheus.GaugeVec
	summaries  map[string]*prometheus.SummaryVec
	histograms map[string]*prometheus.HistogramVec
	lock       sync.Mutex
	registerer prometheus.Registerer
}

// NewMetricsSet creates metrics set with namespace, subSystem and registerer.
//
// If no registerer was provided, then prometheus.DefaultRegisterer would be used.
//
// Important!
// namespace, subSystem, labels should match prometheus regex as bellow
// ^[a-zA-Z_:][a-zA-Z0-9_:]*$
// If provided name is not valid, then default ones would be assigned
func NewMetricsSet(namespace, subSystem string, registerer prometheus.Registerer) *MetricsSet {
	if !model.IsValidMetricName(model.LabelValue(namespace)) {
		namespace = namespaceDefault
	}

	if !model.IsValidMetricName(model.LabelValue(subSystem)) {
		subSystem = subSystemDefault
	}

	metrics := MetricsSet{
		namespace:  namespace,
		subSystem:  subSystem,
		keys:       make(map[string]bool),
		counters:   make(map[string]*prometheus.CounterVec),
		gauges:     make(map[string]*prometheus.GaugeVec),
		summaries:  make(map[string]*prometheus.SummaryVec),
		histograms: make(map[string]*prometheus.HistogramVec),
		lock:       sync.Mutex{},
		registerer: registerer,
	}

	if metrics.registerer == nil {
		metrics.registerer = prometheus.DefaultRegisterer
	}

	return &metrics
}

// GetNamespace returns namespace
func (set *MetricsSet) GetNamespace() string {
	return set.namespace
}

// GetSubSystem returns subsystem
func (set *MetricsSet) GetSubSystem() string {
	return set.subSystem
}

//GetRegisterer returns registerer
func (set *MetricsSet) GetRegisterer() prometheus.Registerer {
	return set.registerer
}

// RegisterCounter is thread safe
// Register a counter with namespace and subsystem in MetricsSet
func (set *MetricsSet) RegisterCounter(name string, labelKeys ...string) error {
	set.lock.Lock()
	defer set.lock.Unlock()

	if err := validateName(name); err != nil {
		return err
	}

	// check existence
	key := set.getKey(name)
	if set.containsKey(key) {
		return errors.New(fmt.Sprintf("duplicate counter name:%s", name))
	}

	// create a new one with default options
	opts := prometheus.CounterOpts{
		Namespace: set.namespace,
		Subsystem: set.subSystem,
		Name:      name,
		Help:      fmt.Sprintf("counter for name:%s and labels:%s", name, labelKeys),
	}

	// panic if labels are not matching
	counterVec := prometheus.NewCounterVec(opts, labelKeys)

	err := set.registerer.Register(counterVec)

	if err == nil {
		set.counters[key] = counterVec
		set.keys[key] = true
	}

	return err
}

// UnRegisterCounter is thread safe
// Unregister metrics, error would be thrown only when invalid name was provided
func (set *MetricsSet) UnRegisterCounter(name string) {
	set.lock.Lock()
	defer set.lock.Unlock()

	key := set.getKey(name)

	// check existence
	if set.containsKey(key) {
		prometheus.Unregister(set.counters[key])

		delete(set.counters, key)
		delete(set.keys, key)
	}
}

// RegisterGauge thread safe
// Register a gauge with namespace and subsystem in MetricsSet
func (set *MetricsSet) RegisterGauge(name string, labelKeys ...string) error {
	set.lock.Lock()
	defer set.lock.Unlock()

	if err := validateName(name); err != nil {
		return err
	}

	// check existence
	key := set.getKey(name)
	if set.containsKey(key) {
		return errors.New(fmt.Sprintf("duplicate gauge name:%s", name))
	}

	// create a new one with default options
	opts := prometheus.GaugeOpts{
		Namespace: set.namespace,
		Subsystem: set.subSystem,
		Name:      name,
		Help:      fmt.Sprintf("Gauge for name:%s and labels:%s", name, labelKeys),
	}

	// panic if labels are not matching
	gaugeVec := prometheus.NewGaugeVec(opts, labelKeys)

	err := set.registerer.Register(gaugeVec)

	if err == nil {
		set.gauges[key] = gaugeVec
		set.keys[key] = true
	}

	return err
}

// UnRegisterGauge thread safe
// Unregister metrics, error would be thrown only when invalid name was provided
func (set *MetricsSet) UnRegisterGauge(name string) {
	set.lock.Lock()
	defer set.lock.Unlock()

	key := set.getKey(name)

	// check existence
	if set.containsKey(key) {
		set.registerer.Unregister(set.gauges[key])

		delete(set.gauges, key)
		delete(set.keys, key)
	}
}

// RegisterHistogram thread safe
// Register a histogram with namespace, subsystem and objectives in MetricsSet
// If bucket is nil, then empty bucket would be applied
func (set *MetricsSet) RegisterHistogram(name string, bucket []float64, labelKeys ...string) error {
	set.lock.Lock()
	defer set.lock.Unlock()

	if err := validateName(name); err != nil {
		return err
	}

	// check existence
	key := set.getKey(name)
	if set.containsKey(key) {
		return errors.New(fmt.Sprintf("duplicate histogram name:%s", name))
	}

	if bucket == nil {
		bucket = make([]float64, 0)
	}

	// create a new one with default options
	opts := prometheus.HistogramOpts{
		Namespace: set.namespace,
		Subsystem: set.subSystem,
		Name:      name,
		Buckets:   bucket,
		Help:      fmt.Sprintf("Histogram for name:%s and labels:%s", name, labelKeys),
	}

	// It will panic if labels are not matching
	hisVec := prometheus.NewHistogramVec(opts, labelKeys)

	err := set.registerer.Register(hisVec)

	if err == nil {
		set.histograms[key] = hisVec
		set.keys[key] = true
	}

	return err
}

// UnRegisterHistogram thread safe
// Unregister metrics, error would be thrown only when invalid name was provided
func (set *MetricsSet) UnRegisterHistogram(name string) {
	set.lock.Lock()
	defer set.lock.Unlock()

	key := set.getKey(name)

	// check existence
	if set.containsKey(key) {
		hisVec := set.histograms[key]

		set.registerer.Unregister(*hisVec)

		delete(set.histograms, key)
		delete(set.keys, key)
	}
}

// RegisterSummary thread safe
// Register a summary with namespace, subsystem and objectives in MetricsSet
// If objectives is nil, then default SummaryObjectives would be applied
func (set *MetricsSet) RegisterSummary(name string, objectives map[float64]float64, labelKeys ...string) error {
	set.lock.Lock()
	defer set.lock.Unlock()

	if err := validateName(name); err != nil {
		return err
	}

	// check existence
	key := set.getKey(name)
	if set.containsKey(key) {
		return errors.New(fmt.Sprintf("duplicate summary name:%s", name))
	}

	if objectives == nil {
		objectives = SummaryObjectives
	}

	// create a new one with default options
	opts := prometheus.SummaryOpts{
		Namespace:  set.namespace,
		Subsystem:  set.subSystem,
		Name:       name,
		Objectives: objectives,
		Help:       fmt.Sprintf("Summary for name:%s and labels:%s", name, labelKeys),
	}

	// panic if labels are not matching
	summaryVec := prometheus.NewSummaryVec(opts, labelKeys)

	err := set.registerer.Register(summaryVec)

	if err == nil {
		set.summaries[key] = summaryVec
		set.keys[key] = true
	}

	return err
}

// UnRegisterSummary thread safe
// Unregister metrics, error would be thrown only when invalid name was provided
func (set *MetricsSet) UnRegisterSummary(name string) {
	set.lock.Lock()
	defer set.lock.Unlock()

	key := set.getKey(name)

	// check existence
	if set.containsKey(key) {
		summaryVec := set.summaries[key]

		set.registerer.Unregister(*summaryVec)
		//prometheus.Unregister(*summaryVec)

		delete(set.summaries, key)
		delete(set.keys, key)
	}
}

// GetCounter is thread safe
func (set *MetricsSet) GetCounter(name string) *prometheus.CounterVec {
	set.lock.Lock()
	defer set.lock.Unlock()

	return set.counters[set.getKey(name)]
}

// GetGauge is thread safe
func (set *MetricsSet) GetGauge(name string) *prometheus.GaugeVec {
	set.lock.Lock()
	defer set.lock.Unlock()

	return set.gauges[set.getKey(name)]
}

// GetHistogram is thread safe
func (set *MetricsSet) GetHistogram(name string) *prometheus.HistogramVec {
	set.lock.Lock()
	defer set.lock.Unlock()

	return set.histograms[set.getKey(name)]
}

// GetSummary is thread safe
func (set *MetricsSet) GetSummary(name string) *prometheus.SummaryVec {
	set.lock.Lock()
	defer set.lock.Unlock()

	return set.summaries[set.getKey(name)]
}

// ListCounters is thread safe
func (set *MetricsSet) ListCounters() []*prometheus.CounterVec {
	set.lock.Lock()
	defer set.lock.Unlock()

	res := make([]*prometheus.CounterVec, 0)
	for _, v := range set.counters {
		res = append(res, v)
	}
	return res
}

// ListGauges is thread safe
func (set *MetricsSet) ListGauges() []*prometheus.GaugeVec {
	set.lock.Lock()
	defer set.lock.Unlock()

	res := make([]*prometheus.GaugeVec, 0)
	for _, v := range set.gauges {
		res = append(res, v)
	}
	return res
}

// ListHistograms is thread safe
func (set *MetricsSet) ListHistograms() []*prometheus.HistogramVec {
	set.lock.Lock()
	defer set.lock.Unlock()

	res := make([]*prometheus.HistogramVec, 0)
	for _, v := range set.histograms {
		res = append(res, v)
	}
	return res
}

// ListSummaries is thread safe
func (set *MetricsSet) ListSummaries() []*prometheus.SummaryVec {
	set.lock.Lock()
	defer set.lock.Unlock()

	res := make([]*prometheus.SummaryVec, 0)
	for _, v := range set.summaries {
		res = append(res, v)
	}
	return res
}

// GetCounterWithValues is thread safe
//
// Get counter with values matched with labels
// Users should always be sure about the number of labels.
func (set *MetricsSet) GetCounterWithValues(name string, values ...string) prometheus.Counter {
	set.lock.Lock()
	defer set.lock.Unlock()

	key := set.getKey(name)

	if set.containsKey(key) {
		counterVec := set.counters[key]
		// ignore err
		counter, _ := counterVec.GetMetricWithLabelValues(values...)
		return counter
	}

	return nil
}

// GetCounterWithLabels is thread safe
//
// Get counter with values matched with labels
// Users should always be sure about the number of labels.
func (set *MetricsSet) GetCounterWithLabels(name string, labels prometheus.Labels) prometheus.Counter {
	set.lock.Lock()
	defer set.lock.Unlock()

	key := set.getKey(name)

	if set.containsKey(key) {
		counterVec := set.counters[key]
		// ignore error
		counter, _ := counterVec.GetMetricWith(labels)

		return counter
	}

	return nil
}

// GetGaugeWithValues is thread safe
//
// Get gauge with values matched with labels
// Users should always be sure about the number of labels.
func (set *MetricsSet) GetGaugeWithValues(name string, values ...string) prometheus.Gauge {
	set.lock.Lock()
	defer set.lock.Unlock()

	key := set.getKey(name)

	if set.containsKey(key) {
		gaugeVec := set.gauges[key]
		// ignore error
		gauge, _ := gaugeVec.GetMetricWithLabelValues(values...)

		return gauge
	}

	return nil
}

// GetGaugeWithLabels is thread safe
// Get gauge with values matched with labels
// Users should always be sure about the number of labels.
func (set *MetricsSet) GetGaugeWithLabels(name string, labels prometheus.Labels) prometheus.Gauge {
	set.lock.Lock()
	defer set.lock.Unlock()

	key := set.getKey(name)

	if set.containsKey(key) {
		gaugeVec := set.gauges[key]
		// ignore error
		gauge, _ := gaugeVec.GetMetricWith(labels)

		return gauge
	}

	return nil
}

// GetSummaryWithValues is thread safe
//
// Get summary with values matched with labels
// Users should always be sure about the number of labels.
func (set *MetricsSet) GetSummaryWithValues(name string, values ...string) prometheus.Observer {
	set.lock.Lock()
	defer set.lock.Unlock()

	key := set.getKey(name)

	if set.containsKey(key) {
		summaryVec := set.summaries[key]
		// ignore error
		observer, _ := summaryVec.GetMetricWithLabelValues(values...)

		return observer
	}

	return nil
}

// GetSummaryWithLabels is thread safe
//
// Get summary with values matched with labels
// Users should always be sure about the number of labels.
func (set *MetricsSet) GetSummaryWithLabels(name string, labels prometheus.Labels) prometheus.Observer {
	set.lock.Lock()
	defer set.lock.Unlock()

	key := set.getKey(name)

	if set.containsKey(key) {
		summaryVec := set.summaries[key]
		// ignore error
		observer, _ := summaryVec.GetMetricWith(labels)

		return observer
	}

	return nil
}

// GetHistogramWithValues is thread safe
//
// Get histogram with values matched with labels
// Users should always be sure about the number of labels.
func (set *MetricsSet) GetHistogramWithValues(name string, values ...string) prometheus.Observer {
	set.lock.Lock()
	defer set.lock.Unlock()

	key := set.getKey(name)

	if set.containsKey(key) {
		hisVec := set.histograms[key]
		// ignore error
		observer, _ := hisVec.GetMetricWithLabelValues(values...)

		return observer
	}

	return nil
}

// GetHistogramWithLabels is thread safe
//
// Get histogram with values matched with labels
// Users should always be sure about the number of labels.
func (set *MetricsSet) GetHistogramWithLabels(name string, labels prometheus.Labels) prometheus.Observer {
	set.lock.Lock()
	defer set.lock.Unlock()

	key := set.getKey(name)

	if set.containsKey(key) {
		hisVec := set.histograms[key]
		// ignore error
		observer, _ := hisVec.GetMetricWith(labels)

		return observer
	}

	return nil
}

// Construct key with format of namespace::subSystem::name
func (set *MetricsSet) getKey(name string) string {
	key := strings.Join([]string{
		set.namespace,
		set.subSystem,
		name}, separator)

	return key
}

// Check existence
func (set *MetricsSet) containsKey(key string) bool {
	_, contains := set.keys[key]

	return contains
}

// Validate input name
func validateName(name string) error {
	name = strings.TrimSpace(name)

	if len(name) < 1 {
		return errors.New("empty name")
	}

	if len(name) > maxKeyLength {
		return errors.New(fmt.Sprintf("exceed max name length:%d", maxKeyLength))
	}

	return nil
}
