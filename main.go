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
	path   string
	mutex  sync.Mutex
	client *http.Client

	up *prometheus.Desc
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
	}
}

// Describe sends the descriptors of the metrics provided by this Collector to the provided channel.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.up
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
