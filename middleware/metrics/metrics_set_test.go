// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkmidmetrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"strings"
	"testing"
)

const (
	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	counter     = "counter"
	gauge       = "gauge"
	histogram   = "histogram"
	summary     = "summary"
	label       = "label"
	value       = "value"
)

var labelMap = map[string]string{label: value}

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func TestNewMetricsSet_WithEmptyNamespace(t *testing.T) {
	set := NewMetricsSet("", "sub_sys", prometheus.DefaultRegisterer)
	assert.NotNil(t, set, "metrics set should not be nil")

	assert.Equal(t, namespaceDefault, set.GetNamespace())
	assert.Equal(t, "sub_sys", set.GetSubSystem())
	assert.Equal(t, prometheus.DefaultRegisterer, set.GetRegisterer())
}

func TestNewMetricsSet_WithEmptySubSystem(t *testing.T) {
	set := NewMetricsSet("ns", "", prometheus.DefaultRegisterer)
	assert.NotNil(t, set, "metrics set should not be nil")

	assert.Equal(t, "ns", set.GetNamespace())
	assert.Equal(t, subSystemDefault, set.GetSubSystem())
	assert.Equal(t, prometheus.DefaultRegisterer, set.GetRegisterer())
}

func TestNewMetricsSet_WithNilRegisterer(t *testing.T) {
	set := NewMetricsSet("ns", "sub_sys", nil)
	assert.NotNil(t, set, "metrics set should not be nil")

	assert.Equal(t, "ns", set.GetNamespace())
	assert.Equal(t, "sub_sys", set.GetSubSystem())
	assert.Equal(t, prometheus.DefaultRegisterer, set.GetRegisterer())
}

func TestNewMetricsSet_HappyCase(t *testing.T) {
	registerer := prometheus.NewRegistry()
	set := NewMetricsSet("ns", "sub_sys", registerer)
	assert.NotNil(t, set, "metrics set should not be nil")

	assert.Equal(t, "ns", set.GetNamespace())
	assert.Equal(t, "sub_sys", set.GetSubSystem())
	assert.Equal(t, registerer, set.GetRegisterer())
}

func TestMetricsSet_GetNamespace_WithEmptyNamespace(t *testing.T) {
	set := NewMetricsSet("", "sub_sys", prometheus.NewRegistry())
	assert.Equal(t, namespaceDefault, set.GetNamespace())
}

func TestMetricsSet_GetNamespace_HappyCase(t *testing.T) {
	set := NewMetricsSet("ns", "sub_sys", prometheus.NewRegistry())
	assert.Equal(t, "ns", set.GetNamespace())
}

func TestMetricsSet_GetNamespace_WithNilRegisterer(t *testing.T) {
	set := NewMetricsSet("ns", "sub_sys", nil)
	assert.Equal(t, prometheus.DefaultRegisterer, set.GetRegisterer())
}

func TestMetricsSet_GetRegisterer_HappyCase(t *testing.T) {
	registerer := prometheus.NewRegistry()
	set := NewMetricsSet("ns", "sub_sys", registerer)
	assert.Equal(t, registerer, set.GetRegisterer())
}

func TestMetricsSet_GetSubSystem_WithEmptySubSystem(t *testing.T) {
	set := NewMetricsSet("ns", "", prometheus.NewRegistry())
	assert.Equal(t, subSystemDefault, set.GetSubSystem())
}

func TestMetricsSet_GetSubSystem_HappyCase(t *testing.T) {
	set := NewMetricsSet("ns", "sub_sys", prometheus.NewRegistry())
	assert.Equal(t, "sub_sys", set.GetSubSystem())
}

// register and unregister counter
func TestMetricsSet_RegisterCounter_WithEmptyName(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	err := set.RegisterCounter("")
	assert.NotNil(t, err)
	assert.Empty(t, set.ListCounters())
}

func TestMetricsSet_RegisterCounter_WithExceedNameLength(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	err := set.RegisterCounter(randStringBytes(257))
	assert.NotNil(t, err)
	assert.Empty(t, set.ListCounters())
}

func TestMetricsSet_RegisterCounter_WithDuplicate(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	err := set.RegisterCounter(counter)
	defer set.UnRegisterCounter(counter)

	assert.Nil(t, err)
	assert.NotEmpty(t, set.ListCounters())

	assert.NotNil(t, set.RegisterCounter(counter))
}

func TestMetricsSet_RegisterCounter_HappyCase(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	err := set.RegisterCounter(counter, label)
	defer set.UnRegisterCounter(counter)

	assert.Nil(t, err)
	assert.NotEmpty(t, set.ListCounters())
	assert.NotNil(t, set.GetCounter(counter))
}

func TestMetricsSet_UnRegisterCounter_WithNonExistKey(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	set.UnRegisterCounter(counter)
	assert.Empty(t, set.ListCounters())
}

func TestMetricsSet_UnRegisterCounter_HappyCase(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	// register counter
	err := set.RegisterCounter(counter, label)
	assert.Nil(t, err)
	assert.NotEmpty(t, set.ListCounters())
	assert.NotNil(t, set.GetCounter(counter))

	// unregister counter
	set.UnRegisterCounter(counter)
	assert.Empty(t, set.ListCounters())
	assert.Nil(t, set.GetCounter(counter))
}

// register and unregister gauge
func TestMetricsSet_RegisterGauge_WithEmptyName(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	err := set.RegisterGauge("", label)
	assert.NotNil(t, err)
	assert.Empty(t, set.ListGauges())
}

func TestMetricsSet_RegisterGauge_WithExceedNameLength(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	err := set.RegisterGauge(randStringBytes(257))
	assert.NotNil(t, err)
	assert.Empty(t, set.ListGauges())
}

func TestMetricsSet_RegisterGauge_WithDuplicate(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	err := set.RegisterGauge(gauge)
	defer set.UnRegisterGauge(gauge)

	assert.Nil(t, err)
	assert.NotEmpty(t, set.ListGauges())

	assert.NotNil(t, set.RegisterGauge(gauge))
}

func TestMetricsSet_RegisterGauge_HappyCase(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	err := set.RegisterGauge(gauge, label)
	defer set.UnRegisterGauge(gauge)

	assert.Nil(t, err)
	assert.NotEmpty(t, set.ListGauges())
	assert.NotNil(t, set.GetGauge(gauge))
}

func TestMetricsSet_UnRegisterGauge_WithNonExistKey(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	set.UnRegisterGauge(gauge)
	assert.Empty(t, set.ListGauges())
}

func TestMetricsSet_UnRegisterGauge_HappyCase(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	// register gauge
	err := set.RegisterGauge(gauge, label)
	assert.Nil(t, err)
	assert.NotEmpty(t, set.ListGauges())
	assert.NotNil(t, set.GetGauge(gauge))

	// unregister gauge
	set.UnRegisterGauge(gauge)
	assert.Empty(t, set.ListGauges())
	assert.Nil(t, set.GetGauge(gauge))
}

// register and unregister histogram
func TestMetricsSet_RegisterHistogram_WithEmptyName(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	err := set.RegisterHistogram("", []float64{}, label)
	assert.NotNil(t, err)
	assert.Empty(t, set.ListHistograms())
}

func TestMetricsSet_RegisterHistogram_WithExceedNameLength(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	err := set.RegisterHistogram(randStringBytes(257), []float64{}, label)
	assert.NotNil(t, err)
	assert.Empty(t, set.ListHistograms())
}

func TestMetricsSet_RegisterHistogram_WithNilBucket(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	err := set.RegisterHistogram(histogram, nil, label)
	defer set.UnRegisterHistogram(histogram)

	assert.Nil(t, err)
	assert.NotEmpty(t, set.ListHistograms())
}

func TestMetricsSet_RegisterHistogram_WithDuplicate(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	err := set.RegisterHistogram(histogram, []float64{}, label)
	defer set.UnRegisterHistogram(histogram)

	assert.Nil(t, err)
	assert.NotEmpty(t, set.ListHistograms())

	assert.NotNil(t, set.RegisterHistogram(histogram, []float64{}, label))
}

func TestMetricsSet_RegisterHistogram_HappyCase(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	err := set.RegisterHistogram(histogram, []float64{}, label)
	defer set.UnRegisterHistogram(histogram)

	assert.Nil(t, err)
	assert.NotEmpty(t, set.ListHistograms())
	assert.NotNil(t, set.GetHistogram(histogram))
}

func TestMetricsSet_UnRegisterHistogram_WithNonExistKey(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	set.UnRegisterHistogram(histogram)
	assert.Empty(t, set.ListHistograms())
}

func TestMetricsSet_UnRegisterHistogram_HappyCase(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	// register histogram
	err := set.RegisterHistogram(histogram, []float64{}, label)
	assert.Nil(t, err)
	assert.NotEmpty(t, set.ListHistograms())
	assert.NotNil(t, set.GetHistogram(histogram))

	// unregister histogram
	set.UnRegisterHistogram(histogram)
	assert.Empty(t, set.ListHistograms())
	assert.Nil(t, set.GetHistogram(histogram))
}

// register and unregister summary
func TestMetricsSet_RegisterSummary_WithEmptyName(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	err := set.RegisterSummary("", map[float64]float64{}, label)
	assert.NotNil(t, err)
	assert.Empty(t, set.ListSummaries())
}

func TestMetricsSet_RegisterSummary_WithExceedNameLength(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	err := set.RegisterSummary(randStringBytes(257), map[float64]float64{}, label)
	assert.NotNil(t, err)
	assert.Empty(t, set.ListSummaries())
}

func TestMetricsSet_RegisterSummary_WithNilObjective(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	err := set.RegisterSummary(summary, nil, label)
	defer set.UnRegisterSummary(summary)

	assert.Nil(t, err)
	assert.NotEmpty(t, set.ListSummaries())
}

func TestMetricsSet_RegisterSummary_WithDuplicate(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	err := set.RegisterSummary(summary, map[float64]float64{}, label)
	defer set.UnRegisterSummary(summary)

	assert.Nil(t, err)
	assert.NotEmpty(t, set.ListSummaries())

	assert.NotNil(t, set.RegisterSummary(summary, map[float64]float64{}, label))
}

func TestMetricsSet_RegisterSummary_HappyCase(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	err := set.RegisterSummary(summary, map[float64]float64{}, label)
	defer set.UnRegisterSummary(summary)

	assert.Nil(t, err)
	assert.NotEmpty(t, set.ListSummaries())
	assert.NotNil(t, set.GetSummary(summary))
}

func TestMetricsSet_UnRegisterSummary_WithNonExistKey(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	set.UnRegisterSummary(summary)
	assert.Empty(t, set.ListSummaries())
}

func TestMetricsSet_UnRegisterSummary_HappyCase(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	// register summary
	err := set.RegisterSummary(summary, map[float64]float64{}, label)
	assert.Nil(t, err)
	assert.NotEmpty(t, set.ListSummaries())
	assert.NotNil(t, set.GetSummary(summary))

	// unregister summary
	set.UnRegisterSummary(summary)
	assert.Empty(t, set.ListSummaries())
	assert.Nil(t, set.GetSummary(summary))
}

// get counter
func TestMetricsSet_GetCounter_ExpectNil(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	assert.Nil(t, set.GetCounter(counter))
}

func TestMetricsSet_GetCounter_HappyCase(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	assert.Nil(t, set.RegisterCounter(counter, label))
	defer set.UnRegisterCounter(counter)
	assert.NotNil(t, set.GetCounter(counter))
}

// get gauge
func TestMetricsSet_GetGauge_ExpectNil(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	assert.Nil(t, set.GetGauge(gauge))
}

func TestMetricsSet_GetGauge_HappyCase(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	assert.Nil(t, set.RegisterGauge(gauge, label))
	defer set.UnRegisterGauge(gauge)
	assert.NotNil(t, set.GetGauge(gauge))
}

// get histogram
func TestMetricsSet_GetHistogram_ExpectNil(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	assert.Nil(t, set.GetHistogram(histogram))
}

func TestMetricsSet_GetHistogram_HappyCase(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	assert.Nil(t, set.RegisterHistogram(histogram, []float64{}, label))
	defer set.UnRegisterHistogram(histogram)
	assert.NotNil(t, set.GetHistogram(histogram))
}

// get summary
func TestMetricsSet_GetSummary_ExpectNil(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	assert.Nil(t, set.GetSummary(summary))
}

func TestMetricsSet_GetSummary_HappyCase(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	assert.Nil(t, set.RegisterSummary(summary, map[float64]float64{}, label))
	defer set.UnRegisterSummary(summary)
	assert.NotNil(t, set.GetSummary(summary))
}

// list counters
func TestMetricsSet_ListCounters_ExpectEmpty(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	assert.Empty(t, set.ListCounters())
}

func TestMetricsSet_ListCounters_HappyCase(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	assert.Nil(t, set.RegisterCounter(counter, label))
	defer set.UnRegisterCounter(counter)
	assert.Len(t, set.ListCounters(), 1)
}

// list gauges
func TestMetricsSet_ListGauges_ExpectEmpty(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	assert.Empty(t, set.ListGauges())
}

func TestMetricsSet_ListGauges_HappyCase(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	assert.Nil(t, set.RegisterGauge(gauge, label))
	defer set.UnRegisterGauge(gauge)
	assert.Len(t, set.ListGauges(), 1)
}

// list histograms
func TestMetricsSet_ListHistograms_ExpectEmpty(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	assert.Empty(t, set.ListHistograms())
}

func TestMetricsSet_ListHistograms_HappyCase(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	assert.Nil(t, set.RegisterHistogram(histogram, []float64{}, label))
	defer set.UnRegisterHistogram(histogram)
	assert.Len(t, set.ListHistograms(), 1)
}

// list summaries
func TestMetricsSet_ListSummaries_ExpectEmpty(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	assert.Empty(t, set.ListSummaries())
}

func TestMetricsSet_ListSummaries_HappyCase(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	assert.Nil(t, set.RegisterSummary(summary, map[float64]float64{}, label))
	defer set.UnRegisterSummary(summary)
	assert.Len(t, set.ListSummaries(), 1)
}

// get counter with values
func TestMetricsSet_GetCounterWithValues_ExpectNil(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	assert.Nil(t, set.GetCounterWithValues(counter))
}

func TestMetricsSet_GetCounterWithValues_HappyCase(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	assert.Nil(t, set.RegisterCounter(counter, label))
	defer set.UnRegisterCounter(counter)
	assert.NotNil(t, set.GetCounterWithValues(counter, value))
}

// get gauge with values
func TestMetricsSet_GetGaugeWithValues_ExpectNil(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	assert.Nil(t, set.GetGaugeWithValues(gauge))
}

func TestMetricsSet_GetGaugeWithValues_HappyCase(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	assert.Nil(t, set.RegisterGauge(gauge, label))
	defer set.UnRegisterGauge(gauge)
	assert.NotNil(t, set.GetGaugeWithValues(gauge, value))
}

// get histogram with values
func TestMetricsSet_GetHistogramWithValues_ExpectNil(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	assert.Nil(t, set.GetHistogramWithValues(histogram))
}

func TestMetricsSet_GetHistogramWithValues_HappyCase(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	assert.Nil(t, set.RegisterHistogram(histogram, []float64{}, label))
	defer set.UnRegisterHistogram(histogram)
	assert.NotNil(t, set.GetHistogramWithValues(histogram, value))
}

// get summary with values
func TestMetricsSet_GetSummaryWithValues_ExpectNil(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	assert.Nil(t, set.GetSummaryWithValues(summary))
}

func TestMetricsSet_GetSummaryWithValues_HappyCase(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	assert.Nil(t, set.RegisterSummary(summary, map[float64]float64{}, label))
	defer set.UnRegisterSummary(summary)
	assert.NotNil(t, set.GetSummaryWithValues(summary, value))
}

// get counter with labels
func TestMetricsSet_GetCounterWithLabels_ExpectNil(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	assert.Nil(t, set.GetCounterWithLabels(counter, labelMap))
}

func TestMetricsSet_GetCounterWithLabels_HappyCase(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	assert.Nil(t, set.RegisterCounter(counter, label))
	defer set.UnRegisterCounter(counter)
	assert.NotNil(t, set.GetCounterWithLabels(counter, labelMap))
}

// get gauge with values
func TestMetricsSet_GetGaugeWithLabels_ExpectNil(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	assert.Nil(t, set.GetGaugeWithLabels(gauge, labelMap))
}

func TestMetricsSet_GetGaugeWithLabels_HappyCase(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	assert.Nil(t, set.RegisterGauge(gauge, label))
	defer set.UnRegisterGauge(gauge)
	assert.NotNil(t, set.GetGaugeWithLabels(gauge, labelMap))
}

// get histogram with values
func TestMetricsSet_GetHistogramWithLabels_ExpectNil(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	assert.Nil(t, set.GetHistogramWithLabels(histogram, labelMap))
}

func TestMetricsSet_GetHistogramWithLabels_HappyCase(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	assert.Nil(t, set.RegisterHistogram(histogram, []float64{}, label))
	defer set.UnRegisterHistogram(histogram)
	assert.NotNil(t, set.GetHistogramWithLabels(histogram, labelMap))
}

// get summary with values
func TestMetricsSet_GetSummaryWithLabels_ExpectNil(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	assert.Nil(t, set.GetSummaryWithLabels(summary, labelMap))
}

func TestMetricsSet_GetSummaryWithLabels_HappyCase(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	assert.Nil(t, set.RegisterSummary(summary, map[float64]float64{}, label))
	defer set.UnRegisterSummary(summary)
	assert.NotNil(t, set.GetSummaryWithLabels(summary, labelMap))
}

func TestMetricsSet_getKey_HappyCase(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	key := set.getKey(counter)
	tokens := strings.Split(key, separator)
	assert.Len(t, tokens, 3)
	assert.Equal(t, set.namespace, tokens[0])
	assert.Equal(t, set.subSystem, tokens[1])
	assert.Equal(t, counter, tokens[2])
}

func TestMetricsSet_containsKey_ExpectFalse(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	assert.False(t, set.containsKey(counter))
}

func TestMetricsSet_containsKey_ExpectTrue(t *testing.T) {
	set := NewMetricsSet("", "", prometheus.NewRegistry())
	assert.Nil(t, set.RegisterCounter(counter, label))
	assert.True(t, set.containsKey(set.getKey(counter)))
}

func TestMetricsSet_validateName_CheckTrimSpace(t *testing.T) {
	assert.Nil(t, validateName(counter+" "))
}

func TestMetricsSet_validateName_WithEmptyString(t *testing.T) {
	assert.NotNil(t, validateName(""))
}

func TestMetricsSet_validateName_WithExceedString(t *testing.T) {
	assert.NotNil(t, validateName(randStringBytes(257)))
}

func TestMetricsSet_validateName_HappyCase(t *testing.T) {
	assert.Nil(t, validateName(counter))
}
