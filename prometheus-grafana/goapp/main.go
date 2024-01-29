package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	PrometheusNamespace   = "game"
	PrometheusSubsystem   = "g1"
	PrometheusConstLabels = prometheus.Labels{"env": "dev"}

	totalRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   PrometheusNamespace,
		Subsystem:   PrometheusSubsystem,
		Name:        "api_requests_total",
		Help:        "The total number of requests.",
		ConstLabels: PrometheusConstLabels,
	}, []string{"router", "method", "code"})

	messageQueueLength = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   PrometheusNamespace,
		Subsystem:   PrometheusSubsystem,
		Name:        "message_queue_length",
		Help:        "Length of message queue.",
		ConstLabels: PrometheusConstLabels,
	}, []string{"name"})

	requestTimeHistogram = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace:   PrometheusNamespace,
		Subsystem:   PrometheusSubsystem,
		Name:        "request_time_histogram",
		Help:        "Histogram of request time consuming.",
		ConstLabels: PrometheusConstLabels,
		Buckets:     prometheus.DefBuckets,
	})

	requestTimeSummary = prometheus.NewSummary(prometheus.SummaryOpts{
		Namespace:   PrometheusNamespace,
		Subsystem:   PrometheusSubsystem,
		Name:        "request_time_summary",
		Help:        "Summary of request time consuming.",
		ConstLabels: PrometheusConstLabels,
		Objectives:  map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.95: 0.01, 0.99: 0.001},
	})
)

func main() {
	registry := prometheus.NewRegistry()

	registry.MustRegister(totalRequests, messageQueueLength, requestTimeHistogram, requestTimeSummary)

	go simulateTotalRequests()
	go simulateMessageQueueLength()
	go simulateRequestTimeHistogram()
	go simulateRequestTimeSummary()

	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	log.Fatalln(http.ListenAndServe("192.168.192.100:60000", nil))
}

func simulateTotalRequests() {
	for {
		totalRequests.WithLabelValues("/user/list", "GET", "200").Inc()

		code := "200"
		if rand.Intn(100) < 30 {
			code = "500"
		}
		totalRequests.WithLabelValues("/user/add", "POST", code).Inc()

		time.Sleep(1 * time.Second)
	}
}

func simulateMessageQueueLength() {
	email := make(chan struct{}, 1000)
	for i := 0; i < 100; i++ {
		email <- struct{}{}
	}
	go func() {
		for {
			email <- struct{}{}
			t := time.Duration(rand.Intn(1000))
			time.Sleep(t * time.Millisecond)
		}
	}()
	go func() {
		for {
			<-email
			t := time.Duration(rand.Intn(1000))
			time.Sleep(t * time.Millisecond)
		}
	}()

	for {
		messageQueueLength.WithLabelValues("email").Set(float64(len(email)))

		time.Sleep(500 * time.Millisecond)
	}
}

func simulateRequestTimeHistogram() {
	for {
		requestTimeHistogram.Observe(rand.Float64())
		time.Sleep(1 * time.Second)
	}
}

func simulateRequestTimeSummary() {
	for {
		requestTimeSummary.Observe(rand.Float64())
		time.Sleep(1 * time.Second)
	}
}
