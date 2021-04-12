// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkentry

import (
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// Request metrics to struct from prometheus collector
// 1: Path - API path
// 2: ElapsedNanoP50 - quantile of p50 with time elapsed
// 3: ElapsedNanoP90 - quantile of p90 with time elapsed
// 4: ElapsedNanoP99 - quantile of p99 with time elapsed
// 5: ElapsedNanoP999 - quantile of p999 with time elapsed
// 6: Count - total number of requests
// 7: ResCode - response code labels
type ReqMetricsRK struct {
	Path            string       `json:"path"`
	ElapsedNanoP50  float64      `json:"elapsed_nano_p50"`
	ElapsedNanoP90  float64      `json:"elapsed_nano_p90"`
	ElapsedNanoP99  float64      `json:"elapsed_nano_p99"`
	ElapsedNanoP999 float64      `json:"elapsed_nano_p999"`
	Count           uint64       `json:"count"`
	ResCode         []*ResCodeRK `json:"res_code"`
}

// Labels and request count
type ResCodeRK struct {
	ResCode string `json:"res_code"`
	Count   uint64 `json:"count"`
}

// Parse metrics in prometheus client into rk style metrics for common servic.
func NewPromMetricsInfo(sumCollector *prometheus.SummaryVec) []*ReqMetricsRK {
	res := make([]*ReqMetricsRK, 0)

	// request total by path
	// why we need this? since prometheus will record each summary by label
	// and path will occur multiple times with other labels like res_code
	var reqMetricsMap = make(map[string]*ReqMetricsRK)

	// get counters from rk_gin_log interceptor
	// counters would be empty if metrics option was turned off
	// from config file
	channel := make(chan prometheus.Metric)

	// collect metrics from prometheus client
	go func() {
		sumCollector.Collect(channel)
		close(channel)
	}()

	// iterate metrics
	for element := range channel {
		// write to family
		metricsPB := &dto.Metric{}
		if err := element.Write(metricsPB); err != nil {
			continue
		}

		// iterate labels
		path, code := getPathAndResCode(metricsPB)
		// we got path, let's continue
		if len(path) > 0 {
			// contains the same path?
			if val, ok := reqMetricsMap[path]; ok {
				// 1: we meet request summary with same path add value to it
				for j := range metricsPB.GetSummary().GetQuantile() {
					switch quantile := metricsPB.GetSummary().Quantile[j].GetQuantile(); quantile {
					case 0.5:
						val.ElapsedNanoP50 = calcNewQuantile(
							val.Count,
							metricsPB.GetSummary().GetSampleCount(),
							val.ElapsedNanoP50,
							metricsPB.GetSummary().Quantile[j].GetValue())
					case 0.9:
						val.ElapsedNanoP90 = calcNewQuantile(
							val.Count,
							metricsPB.GetSummary().GetSampleCount(),
							val.ElapsedNanoP90,
							metricsPB.GetSummary().Quantile[j].GetValue())
					case 0.99:
						val.ElapsedNanoP99 = calcNewQuantile(
							val.Count,
							metricsPB.GetSummary().GetSampleCount(),
							val.ElapsedNanoP99,
							metricsPB.GetSummary().Quantile[j].GetValue())
					case 0.999:
						val.ElapsedNanoP999 = calcNewQuantile(
							val.Count,
							metricsPB.GetSummary().GetSampleCount(),
							val.ElapsedNanoP999,
							metricsPB.GetSummary().Quantile[j].GetValue())
					default:
						// do nothing
					}
				}

				// 2: add total count
				val.Count += metricsPB.GetSummary().GetSampleCount()

				// 3: add res code
				for i := range val.ResCode {
					if val.ResCode[i].ResCode != code {
						// add it if different
						val.ResCode = append(val.ResCode, &ResCodeRK{
							ResCode: code,
							Count:   metricsPB.GetSummary().GetSampleCount(),
						})
					} else {
						val.ResCode[i].Count += metricsPB.GetSummary().GetSampleCount()
					}
				}
			} else {
				rk := &ReqMetricsRK{
					Path:    path,
					ResCode: make([]*ResCodeRK, 0),
				}
				// 1: record count of request
				rk.Count += metricsPB.GetSummary().GetSampleCount()

				// 2: iterate summary and extract quantile for P50, P90 and P99
				for j := range metricsPB.Summary.Quantile {
					switch quantile := metricsPB.GetSummary().Quantile[j].GetQuantile(); quantile {
					case 0.5:
						rk.ElapsedNanoP50 = metricsPB.GetSummary().GetQuantile()[j].GetValue()
					case 0.9:
						rk.ElapsedNanoP90 = metricsPB.GetSummary().GetQuantile()[j].GetValue()
					case 0.99:
						rk.ElapsedNanoP99 = metricsPB.GetSummary().GetQuantile()[j].GetValue()
					case 0.999:
						rk.ElapsedNanoP999 = metricsPB.GetSummary().GetQuantile()[j].GetValue()
					default:
						// do nothing
					}
				}

				// 3: add res code
				rk.ResCode = append(rk.ResCode, &ResCodeRK{
					ResCode: code,
					Count:   metricsPB.GetSummary().GetSampleCount(),
				})

				reqMetricsMap[path] = rk
			}
		}
	}

	for _, v := range reqMetricsMap {
		res = append(res, v)
	}

	return res
}

// parse out path and response code
func getPathAndResCode(metricsPB *dto.Metric) (string, string) {
	var path = ""
	var resCode = ""
	for i := range metricsPB.Label {
		if metricsPB.Label[i].GetName() == "path" {
			path = metricsPB.Label[i].GetValue()
		}

		if metricsPB.Label[i].GetName() == "res_code" {
			resCode = metricsPB.Label[i].GetValue()
		}
	}

	return path, resCode
}

// calculate new quantile based on number
// formula:
// newQuantile = oldQuantile*oldFraction + newQuantile*newFraction
func calcNewQuantile(oldCount, newCount uint64, oldQuantile, newQuantile float64) float64 {
	oldFraction := float64(oldCount) / float64(oldCount+newCount)
	newFraction := float64(newCount) / float64(oldCount+newCount)
	return oldQuantile*oldFraction + newQuantile*newFraction
}
