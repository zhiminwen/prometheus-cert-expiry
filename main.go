package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dsnet/try"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	conf := ReadConfig("config.yaml")
	log.Printf("config: %+v", conf)

	certExpiryVector := RegisterMetrics()
	go checkCerts(conf, certExpiryVector)

	http.Handle("/metrics", promhttp.Handler())

	try.E(http.ListenAndServe(fmt.Sprintf(":%d", conf.Port), nil))
}

func RegisterMetrics() *prometheus.GaugeVec {
	gaugeVector := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "certificate",
			Subsystem: "checker",
			Name:      "expiry",
			Help:      "days left to the cert expiry",
		},
		[]string{
			"type",
			"name",
			"path",
			"address",
			"subject",
			"issuer",
		},
	)
	prometheus.MustRegister(gaugeVector)

	return gaugeVector
}

func checkCerts(conf Config, certExpiryVector *prometheus.GaugeVec) {
	for _, cert := range conf.Certs {
		duration, err := time.ParseDuration(cert.Interval)
		if err != nil {
			log.Printf("failed to parse duration: %s, ignore this cert check:%s", err, cert.Name)
			continue
		}
		go func(cert Cert, duration time.Duration) {
			for { //loop forever
				checkCert(cert, certExpiryVector)
				time.Sleep(duration)
			}
		}(cert, duration)
	}
}
