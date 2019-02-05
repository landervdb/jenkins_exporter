# Jenkins Exporter for Prometheus

_This is a work in progress, and things might still change drastically._

This exporter exports Prometheus metrics retrieved from Jenkins by parsing the XML files in the Jenkins jobs folder.

## Building and running

Prerequisites:

 - Go 1.11+
 - Jenkins Build Environment plugin installed

This exporter makes use of the new Go modules system, so you need at least Go 1.11. You can clone this repository to any path and then just run `go build`. 

Alternatively, you can still put the repository in your `$GOPATH`, and compile the exporter using

```bash
$ GO111MODULES=on go build
```

To run the exporter, it needs access to the file system folder where Jenkins stores its configuration (by default, this is `/var/lib/jenkins`). Additionally, the _Build Environment_ plugin for Jenkins is required, since this allows us to parse environment variables that were set during the builds in question on Jenkins.

```bash
$ jenkins_exporter -h
Usage of jenkins_exporter:
  -jenkins.envvars string
    	Custom environment variables to parse into metrics. Format: ENVVAR1:metric_name;ENVVAR2:metric_name,...
  -jenkins.path string
    	Path to the Jenkins folder (default "/var/lib/jenkins")
  -log.level string
    	The minimal log level to be displayed (default "INFO")
  -metrics.bind string
    	Address to expose the metrics on (default ":9506")
  -metrics.path string
    	Path to expose the metrics on (default "/metrics")
```

## Exported metrics

```
# HELP jenkins_collect_duration_seconds The time it took to collect the metrics in seconds
# TYPE jenkins_collect_duration_seconds gauge
jenkins_collect_duration_seconds 0.013396465
# HELP jenkins_collect_failures The number of collection failures since the exporter was started
# TYPE jenkins_collect_failures counter
jenkins_collect_failures 0
# HELP jenkins_custom_last_checkout_build_number Custom metric generated from environment variable CHECKOUT_BUILD_NUMBER
# TYPE jenkins_custom_last_checkout_build_number gauge
jenkins_custom_last_checkout_build_number{folder="{folder}",jenkins_job="{job}",result="{result}"} 4
# HELP jenkins_exporter_build_info A metric with a constant '1' value labeled by version, revision, branch, and goversion from which jenkins_exporter was built.
# TYPE jenkins_exporter_build_info gauge
jenkins_exporter_build_info{branch="",goversion="go1.11.5",revision="",version=""} 1
# HELP jenkins_last_build_duration_seconds Duration of the last build
# TYPE jenkins_last_build_duration_seconds gauge
jenkins_last_build_duration_seconds{folder="{folder}",jenkins_job="{job}",result="{result}"} 0.332
# HELP jenkins_last_build_number Build number of the last build
# TYPE jenkins_last_build_number gauge
jenkins_last_build_number{folder="{folder}",jenkins_job="{job}",result="{result}"} 10
# HELP jenkins_last_build_timestamp_seconds Timestamp of the last build
# TYPE jenkins_last_build_timestamp_seconds gauge
jenkins_last_build_timestamp_seconds{folder="{folder}",jenkins_job="{job}",result="{result}"} 1.549030450633e+09
# HELP jenkins_up Whether the Jenkins path is a valid Jenkins tree
# TYPE jenkins_up gauge
jenkins_up 1
```

## Custom metrics

By using the `-jenkins.envvars` command line flag, you can add custom metrics. These are parsed from the environment variable (set during the build of the Jenkins job) you define. Environment variables with a non-numerical value will be ignored. The following syntax is expected: 

```
ENVVARNAME:metric_name[;ENVARBANE2:second_metric_name[...]]
```

## Author

This exporter was created by [landervdb](https://github.com/landervdb).
