package main

import (
	"net/http"
	"time"

	"math/rand/v2"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func randRange(min, max int) int {
	return rand.IntN(max-min) + min
}

func main() {

	h := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:                            "my_histogram",
		NativeHistogramBucketFactor:     1.1,
		NativeHistogramMaxBucketNumber:  100,
		NativeHistogramMinResetDuration: 1 * time.Hour,
	}, []string{"id"})

	r := prometheus.NewRegistry()
	r.MustRegister(h)

	handler := promhttp.HandlerFor(
		r,
		promhttp.HandlerOpts{
			EnableOpenMetrics: false,
		})

	h.With(prometheus.Labels{"id": "first"}).Observe(10)
	h.With(prometheus.Labels{"id": "first"}).Observe(40)
	h.With(prometheus.Labels{"id": "first"}).Observe(80)
	h.With(prometheus.Labels{"id": "second"}).Observe(340)
	h.With(prometheus.Labels{"id": "second"}).Observe(640)
	h.With(prometheus.Labels{"id": "second"}).Observe(10040)

	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "my_gauge",
	})
	gauge.Set(512)
	r.MustRegister(gauge)

	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		quit := make(chan struct{})
		go func() {
			for {
				select {
				case <-ticker.C:
					h.With(prometheus.Labels{"id": "first"}).Observe(float64(randRange(10, 100)))
					h.With(prometheus.Labels{"id": "second"}).Observe(float64(randRange(300, 10000)))
				case <-quit:
					ticker.Stop()
					return
				}
			}
		}()
	}()

	http.Handle("/metrics", handler)
	http.ListenAndServe("127.0.0.1:2112", nil)
}
