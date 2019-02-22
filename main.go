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
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/landervdb/jenkins_exporter/jenkins"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
)

var (
	bind        = flag.String("metrics.bind", ":9506", "Address to expose the metrics on")
	path        = flag.String("metrics.path", "/metrics", "Path to expose the metrics on")
	ignoreList  = flag.String("jenkins.ignore", "", "Comma-separated list of folders to ignore")
	jenkinsPath = flag.String("jenkins.path", "/var/lib/jenkins", "Path to the Jenkins folder")
	envVars     = flag.String("jenkins.envvars", "", "Custom environment variables to parse into metrics. Format: ENVVAR1:metric_name;ENVVAR2:metric_name,...")
	logLevel    = flag.String("log.level", "INFO", "The minimal log level to be displayed")
)

const (
	namespace  = "jenkins"
	numParsers = 20
)

// Collector is a Prometheus Collector that fetches and generates the Jenkins metrics.
type Collector struct {
	opts               jenkins.JobPathOpts
	mutex              sync.Mutex
	up                 *prometheus.Desc
	collectDuration    *prometheus.Desc
	collectFailures    prometheus.Counter
	lastBuildNumber    *prometheus.GaugeVec
	lastBuildTimestamp *prometheus.GaugeVec
	lastBuildDuration  *prometheus.GaugeVec
	customGauges       map[string]*prometheus.GaugeVec
}

// NewCollector creates an instance of Collector.
func NewCollector(opts jenkins.JobPathOpts) *Collector {
	return &Collector{
		opts: opts,
		up: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "up"),
			"Whether the Jenkins path is a valid Jenkins tree",
			nil,
			nil),
		collectDuration: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "collect_duration_seconds"),
			"The time it took to collect the metrics in seconds",
			nil,
			nil),
		collectFailures: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "collect_failures",
				Help:      "The number of collection failures since the exporter was started",
			},
		),
		lastBuildNumber: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "last_build_number",
				Help:      "Build number of the last build",
			},
			[]string{"folder", "jenkins_job", "result"},
		),
		lastBuildTimestamp: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "last_build_timestamp_seconds",
				Help:      "Timestamp of the last build",
			},
			[]string{"folder", "jenkins_job", "result"},
		),
		lastBuildDuration: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "last_build_duration_seconds",
				Help:      "Duration of the last build",
			},
			[]string{"folder", "jenkins_job", "result"},
		),
	}
}

// Describe sends the descriptors of the metrics provided by this Collector to the provided channel.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.up
	ch <- c.collectDuration
	c.collectFailures.Describe(ch)
	c.lastBuildNumber.Describe(ch)
	c.lastBuildTimestamp.Describe(ch)
	c.lastBuildDuration.Describe(ch)

	for _, cg := range c.customGauges {
		cg.Describe(ch)
	}
}

// Collect actually collects all the metrics provided by this Collector and sends them to the provided channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	log.Info("Started collection")

	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime).Seconds()
		log.Infof("Collection completed in %f seconds", duration)
		ch <- prometheus.MustNewConstMetric(c.collectDuration, prometheus.GaugeValue, duration)
	}()

	jobPaths := make(chan jenkins.JobPath)
	go func() {
		err := jenkins.GetJobPaths(c.opts, jobPaths)
		if err != nil {
			log.Errorf("collecting job paths failed: %v", err)
			ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 0)
			c.collectFailures.Inc()
			return
		}
		ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 1)
	}()

	jobs := make(chan jenkins.Job)
	var wg sync.WaitGroup
	wg.Add(numParsers)
	for i := 0; i < numParsers; i++ {
		go func() {
			doParse(jobPaths, jobs)
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(jobs)
	}()

	for job := range jobs {
		if job.LastSuccessfulBuild.Number != 0 {
			c.lastBuildNumber.WithLabelValues(job.Folder, job.Name, "successful").Set(float64(job.LastSuccessfulBuild.Number))
			c.lastBuildTimestamp.WithLabelValues(job.Folder, job.Name, "successful").Set(float64(job.LastSuccessfulBuild.Timestamp) / 1000)
			c.lastBuildDuration.WithLabelValues(job.Folder, job.Name, "successful").Set(float64(job.LastSuccessfulBuild.Duration) / 1000)
			populateCustomGauges(job.Folder, job.Name, "successful", job.LastSuccessfulBuild, c.customGauges)
		}

		if job.LastUnsuccessfulBuild.Number != 0 {
			c.lastBuildNumber.WithLabelValues(job.Folder, job.Name, "unsuccessful").Set(float64(job.LastUnsuccessfulBuild.Number))
			c.lastBuildTimestamp.WithLabelValues(job.Folder, job.Name, "unsuccessful").Set(float64(job.LastUnsuccessfulBuild.Timestamp) / 1000)
			c.lastBuildDuration.WithLabelValues(job.Folder, job.Name, "unsuccessful").Set(float64(job.LastUnsuccessfulBuild.Duration) / 1000)
			populateCustomGauges(job.Folder, job.Name, "unsuccessful", job.LastUnsuccessfulBuild, c.customGauges)
		}

		if job.LastStableBuild.Number != 0 {
			c.lastBuildNumber.WithLabelValues(job.Folder, job.Name, "stable").Set(float64(job.LastStableBuild.Number))
			c.lastBuildTimestamp.WithLabelValues(job.Folder, job.Name, "stable").Set(float64(job.LastStableBuild.Timestamp) / 1000)
			c.lastBuildDuration.WithLabelValues(job.Folder, job.Name, "stable").Set(float64(job.LastStableBuild.Duration) / 1000)
			populateCustomGauges(job.Folder, job.Name, "stable", job.LastStableBuild, c.customGauges)
		}

		if job.LastUnstableBuild.Number != 0 {
			c.lastBuildNumber.WithLabelValues(job.Folder, job.Name, "unstable").Set(float64(job.LastUnstableBuild.Number))
			c.lastBuildTimestamp.WithLabelValues(job.Folder, job.Name, "unstable").Set(float64(job.LastUnstableBuild.Timestamp) / 1000)
			c.lastBuildDuration.WithLabelValues(job.Folder, job.Name, "unstable").Set(float64(job.LastUnstableBuild.Duration) / 1000)
			populateCustomGauges(job.Folder, job.Name, "unstable", job.LastUnstableBuild, c.customGauges)
		}

		if job.LastFailedBuild.Number != 0 {
			c.lastBuildNumber.WithLabelValues(job.Folder, job.Name, "failed").Set(float64(job.LastFailedBuild.Number))
			c.lastBuildTimestamp.WithLabelValues(job.Folder, job.Name, "failed").Set(float64(job.LastFailedBuild.Timestamp) / 1000)
			c.lastBuildDuration.WithLabelValues(job.Folder, job.Name, "failed").Set(float64(job.LastFailedBuild.Duration) / 1000)
			populateCustomGauges(job.Folder, job.Name, "failed", job.LastFailedBuild, c.customGauges)
		}

		log.Debugf("Parsed job %s in folder %s", job.Name, job.Folder)
	}

	c.collectFailures.Collect(ch)
	c.lastBuildNumber.Collect(ch)
	c.lastBuildDuration.Collect(ch)
	c.lastBuildTimestamp.Collect(ch)

	for _, cg := range c.customGauges {
		cg.Collect(ch)
	}
}

func doParse(jobPaths <-chan jenkins.JobPath, jobs chan<- jenkins.Job) {
	for jobPath := range jobPaths {
		job, err := jobPath.Parse()
		if err != nil {
			log.Debugf("Failed to parse %s: %v", jobPath, err)
			continue
		}
		jobs <- job
	}
}

func createCustomGauges(input string) (map[string]*prometheus.GaugeVec, error) {

	customGauges := make(map[string]*prometheus.GaugeVec)

	if len(input) == 0 {
		return customGauges, nil
	}

	pairs := strings.Split(input, ";")
	for _, pair := range pairs {
		els := strings.Split(pair, ":")
		if len(els) != 2 {
			return customGauges, fmt.Errorf("Custom metrics config format is invalid: %s", input)
		}
		envVar := els[0]
		metricName := els[1]
		match, err := regexp.MatchString("^[a-z_]*$", metricName)
		if err != nil || !match {
			return customGauges, fmt.Errorf("Provided invalid metric name: %s", metricName)
		}
		customGauges[envVar] = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      fmt.Sprintf("custom_last_%s", metricName),
				Help:      fmt.Sprintf("Custom metric generated from environment variable %s", envVar),
			},
			[]string{"folder", "jenkins_job", "result"},
		)
		log.Infof("Added custom metric custom_last_%s using %s", metricName, envVar)
	}

	return customGauges, nil
}

func populateCustomGauges(folder, job, result string, build jenkins.Build, customGauges map[string]*prometheus.GaugeVec) {
	for ev, cg := range customGauges {
		val, ok := build.EnvVars[ev]
		if !ok {
			continue
		}
		parsed, err := strconv.ParseFloat(val, 64)
		if err != nil {
			log.Debugf("Couldn't parse environment variable %s: %v", ev, err)
			continue
		}
		cg.WithLabelValues(folder, job, result).Set(parsed)
	}
}

func main() {

	flag.Parse()

	log.Base().SetLevel(*logLevel)
	log.Infoln("Starting Jenkins exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	opts := &jenkins.JobPathOpts{
		Root:       *jenkinsPath,
		IgnoreList: strings.Split(*ignoreList, ","),
	}

	collector := NewCollector(*opts)

	customMetrics, err := createCustomGauges(*envVars)
	if err != nil {
		log.Fatalf("Error parsing custom metrics config: %v", err)
	}
	collector.customGauges = customMetrics

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
