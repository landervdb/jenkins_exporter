// Copyright 2019 Lander Van den Bulcke
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"net/http"
	"os"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
)

var (
	bind        = flag.String("metrics.bind", ":9506", "Address to expose the metrics on")
	path        = flag.String("metrics.path", "/metrics", "Path to expose the metrics on")
	jenkinsPath = flag.String("jenkins.path", "/var/lib/jenkins", "Path to the Jenkins folder")
	logLevel    = flag.String("log.level", "INFO", "The minimal log level to be displayed")
)

const (
	namespace = "jenkins"
)

// Collector is a Prometheus Collector that fetches and generates the Jenkins metrics.
type Collector struct {
	path                           string
	mutex                          sync.Mutex
	up                             *prometheus.Desc
	lastBuildNumber                *prometheus.GaugeVec
	lastBuildTimestamp             *prometheus.GaugeVec
	lastBuildDuration              *prometheus.GaugeVec
	lastSuccessfulBuildNumber      *prometheus.GaugeVec
	lastSuccessfulBuildTimestamp   *prometheus.GaugeVec
	lastSuccessfulBuildDuration    *prometheus.GaugeVec
	lastUnsuccessfulBuildNumber    *prometheus.GaugeVec
	lastUnsuccessfulBuildTimestamp *prometheus.GaugeVec
	lastUnsuccessfulBuildDuration  *prometheus.GaugeVec
	lastStableBuildNumber          *prometheus.GaugeVec
	lastStableBuildTimestamp       *prometheus.GaugeVec
	lastStableBuildDuration        *prometheus.GaugeVec
	lastUnstableBuildNumber        *prometheus.GaugeVec
	lastUnstableBuildTimestamp     *prometheus.GaugeVec
	lastUnstableBuildDuration      *prometheus.GaugeVec
	lastFailedBuildNumber          *prometheus.GaugeVec
	lastFailedBuildTimestamp       *prometheus.GaugeVec
	lastFailedBuildDuration        *prometheus.GaugeVec
}

// NewCollector creates an instance of Collector.
func NewCollector(path string) *Collector {
	return &Collector{
		path: path,
		up: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "up"),
			"Could the Jenkins folder be parsed",
			nil,
			nil),
		lastBuildNumber: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "last_build_number",
			Help:      "Build number of the last build",
		},
			[]string{"folder", "job"},
		),
		lastBuildTimestamp: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "last_build_timestamp",
			Help:      "Timestamp of the last build",
		},
			[]string{"folder", "job"},
		),
		lastBuildDuration: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "last_build_duration",
			Help:      "Duration of the last build",
		},
			[]string{"folder", "job"},
		),
		lastSuccessfulBuildNumber: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "last_successful_build_number",
			Help:      "Build number of the last successful build",
		},
			[]string{"folder", "job"},
		),
		lastSuccessfulBuildTimestamp: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "last_successful_build_timestamp",
			Help:      "Timestamp of the last successful build",
		},
			[]string{"folder", "job"},
		),
		lastSuccessfulBuildDuration: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "last_successful_build_duration",
			Help:      "Duration of the last successful build",
		},
			[]string{"folder", "job"},
		),
		lastUnsuccessfulBuildNumber: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "last_unsuccessful_build_number",
			Help:      "Build number of the last unsuccessful build",
		},
			[]string{"folder", "job"},
		),
		lastUnsuccessfulBuildTimestamp: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "last_unsuccessful_build_timestamp",
			Help:      "Timestamp of the last unsuccessful build",
		},
			[]string{"folder", "job"},
		),
		lastUnsuccessfulBuildDuration: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "last_unsuccessful_build_duration",
			Help:      "Duration of the last unsuccessful build",
		},
			[]string{"folder", "job"},
		),
		lastStableBuildNumber: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "last_stable_build_number",
			Help:      "Build number of the last stable build",
		},
			[]string{"folder", "job"},
		),
		lastStableBuildTimestamp: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "last_stable_build_timestamp",
			Help:      "Timestamp of the last stable build",
		},
			[]string{"folder", "job"},
		),
		lastStableBuildDuration: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "last_stable_build_duration",
			Help:      "Duration of the last stable build",
		},
			[]string{"folder", "job"},
		),
		lastUnstableBuildNumber: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "last_unstable_build_number",
			Help:      "Build number of the last unstable build",
		},
			[]string{"folder", "job"},
		),
		lastUnstableBuildTimestamp: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "last_unstable_build_timestamp",
			Help:      "Timestamp of the last unstable build",
		},
			[]string{"folder", "job"},
		),
		lastUnstableBuildDuration: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "last_unstable_build_duration",
			Help:      "Duration of the last unstable build",
		},
			[]string{"folder", "job"},
		),
		lastFailedBuildNumber: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "last_failed_build_number",
			Help:      "Build number of the last failed build",
		},
			[]string{"folder", "job"},
		),
		lastFailedBuildTimestamp: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "last_failed_build_timestamp",
			Help:      "Timestamp of the last failed build",
		},
			[]string{"folder", "job"},
		),
		lastFailedBuildDuration: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "last_failed_build_duration",
			Help:      "Duration of the last failed build",
		},
			[]string{"folder", "job"},
		),
	}
}

// Describe sends the descriptors of the metrics provided by this Collector to the provided channel.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.up
	c.lastBuildNumber.Describe(ch)
	c.lastBuildTimestamp.Describe(ch)
	c.lastBuildDuration.Describe(ch)
	c.lastSuccessfulBuildNumber.Describe(ch)
	c.lastSuccessfulBuildTimestamp.Describe(ch)
	c.lastSuccessfulBuildDuration.Describe(ch)
	c.lastUnsuccessfulBuildNumber.Describe(ch)
	c.lastUnsuccessfulBuildTimestamp.Describe(ch)
	c.lastUnsuccessfulBuildDuration.Describe(ch)
	c.lastStableBuildNumber.Describe(ch)
	c.lastStableBuildTimestamp.Describe(ch)
	c.lastStableBuildDuration.Describe(ch)
	c.lastUnstableBuildNumber.Describe(ch)
	c.lastUnstableBuildTimestamp.Describe(ch)
	c.lastUnstableBuildDuration.Describe(ch)
	c.lastFailedBuildNumber.Describe(ch)
	c.lastFailedBuildTimestamp.Describe(ch)
	c.lastFailedBuildDuration.Describe(ch)
}

// Collect actually collects all the metrics provided by this Collector and sends them to the provided channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var up float64 = 1

	_, err := os.Stat(c.path)
	if err != nil {
		up = 0
		log.Errorf("Failed to parse contents of %s", c.path)
	}

	ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, up)
}

func main() {

	flag.Parse()

	log.Base().SetLevel(*logLevel)
	log.Infoln("Starting Jenkins exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	collector := NewCollector(*jenkinsPath)
	prometheus.MustRegister(collector)
	prometheus.MustRegister(version.NewCollector("jenkins_exporter"))

	http.Handle(*path, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			 <head><title>Jenkins Exporter</title></head>
			 <body>
			 <h1>Jenkins Exporter</h1>
			 <p><a href='` + *path + `'>Metrics</a></p>
			 </body>
			 </html>`))
	})

	log.Infof("Listening on %s", *bind)
	if err := http.ListenAndServe(*bind, nil); err != nil {
		log.Fatal(err)
	}

}
