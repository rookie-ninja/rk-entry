// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkentry

import (
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"path"
)

// ReqMetricsRK request metrics to struct from prometheus collector
// 1: RestPath - API path of restful service
// 2: RestMethod - API method of restful service
// 3: GrpcService - Grpc service
// 4: GrpcMethod - Grpc method
// 5: ElapsedNanoP50 - quantile of p50 with time elapsed
// 6: ElapsedNanoP90 - quantile of p90 with time elapsed
// 7: ElapsedNanoP99 - quantile of p99 with time elapsed
// 8: ElapsedNanoP999 - quantile of p999 with time elapsed
// 9: Count - total number of requests
// 10: ResCode - response code labels
type ReqMetricsRK struct {
	RestPath        string       `json:"restPath" yaml:"restPath"`
	RestMethod      string       `json:"restMethod" yaml:"restMethod"`
	GrpcService     string       `json:"grpcService" yaml:"grpcService"`
	GrpcMethod      string       `json:"grpcMethod" yaml:"grpcMethod"`
	ElapsedNanoP50  float64      `json:"elapsedNanoP50" yaml:"elapsedNanoP50"`
	ElapsedNanoP90  float64      `json:"elapsedNanoP90" yaml:"elapsedNanoP90"`
	ElapsedNanoP99  float64      `json:"elapsedNanoP99" yaml:"elapsedNanoP99"`
	ElapsedNanoP999 float64      `json:"elapsedNanoP999" yaml:"elapsedNanoP999"`
	Count           uint64       `json:"count" yaml:"count"`
	ResCode         []*ResCodeRK `json:"resCode" yaml:"resCode"`
}

// ResCodeRK defines labels and request count
type ResCodeRK struct {
	ResCode string `json:"resCode" yaml:"resCode"`
	Count   uint64 `json:"count" yaml:"count"`
}

// NewPromMetricsInfo parse metrics in prometheus client into rk style metrics for common service.
func NewPromMetricsInfo(sumCollector *prometheus.SummaryVec) []*ReqMetricsRK {
	res := make([]*ReqMetricsRK, 0)

	// Request total by path.
	// Why we need this? since prometheus will record each summary by label
	// and path will occur multiple times with other labels like res_code
	var reqMetricsMap = make(map[string]*ReqMetricsRK)

	// Get counters from interceptor
	// counters would be empty if metrics option was turned off
	// from config file
	channel := make(chan prometheus.Metric)

	// Collect metrics from prometheus client
	go func() {
		sumCollector.Collect(channel)
		close(channel)
	}()

	// Iterate metrics
	for element := range channel {
		// Write to family
		metricsPB := &dto.Metric{}
		if err := element.Write(metricsPB); err != nil {
			continue
		}

		// Iterate labels
		grpcService, grpcMethod, restPath, restMethod, resCode := getPathMethodAndResCode(metricsPB)
		key := path.Join(grpcService, grpcMethod, restMethod, restPath)
		// We got path, let's continue
		if len(key) > 0 {
			// contains the same path?
			if val, ok := reqMetricsMap[key]; ok {
				// 1: We meet request summary with same path add value to it
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

				// 2: Add total count
				val.Count += metricsPB.GetSummary().GetSampleCount()

				// 3: Add res code
				for i := range val.ResCode {
					if val.ResCode[i].ResCode != resCode {
						// add it if different
						val.ResCode = append(val.ResCode, &ResCodeRK{
							ResCode: resCode,
							Count:   metricsPB.GetSummary().GetSampleCount(),
						})
					} else {
						val.ResCode[i].Count += metricsPB.GetSummary().GetSampleCount()
					}
				}
			} else {
				rk := &ReqMetricsRK{
					GrpcService: grpcService,
					GrpcMethod:  grpcMethod,
					RestMethod:  restMethod,
					RestPath:    restPath,
					ResCode:     make([]*ResCodeRK, 0),
				}
				// 1: Record count of request
				rk.Count += metricsPB.GetSummary().GetSampleCount()

				// 2: Iterate summary and extract quantile for P50, P90 and P99
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

				// 3: Add res code
				rk.ResCode = append(rk.ResCode, &ResCodeRK{
					ResCode: resCode,
					Count:   metricsPB.GetSummary().GetSampleCount(),
				})

				reqMetricsMap[key] = rk
			}
		}
	}

	for _, v := range reqMetricsMap {
		res = append(res, v)
	}

	return res
}

// getPathMethodAndResCode parse out path and response code.
func getPathMethodAndResCode(metricsPB *dto.Metric) (grpcService, grpcMethod, restPath, restMethod, resCode string) {
	for i := range metricsPB.Label {
		switch metricsPB.Label[i].GetName() {
		case "grpcService":
			grpcService = metricsPB.Label[i].GetValue()
		case "restMethod":
			restMethod = metricsPB.Label[i].GetValue()
		case "restPath":
			restPath = metricsPB.Label[i].GetValue()
		case "grpcMethod":
			grpcMethod = metricsPB.Label[i].GetValue()
		case "resCode":
			resCode = metricsPB.Label[i].GetValue()
		}
	}

	return grpcService, grpcMethod, restPath, restMethod, resCode
}

// Calculate new quantile based on number.
// formula:
// newQuantile = oldQuantile*oldFraction + newQuantile*newFraction
func calcNewQuantile(oldCount, newCount uint64, oldQuantile, newQuantile float64) float64 {
	oldFraction := float64(oldCount) / float64(oldCount+newCount)
	newFraction := float64(newCount) / float64(oldCount+newCount)
	return oldQuantile*oldFraction + newQuantile*newFraction
}
