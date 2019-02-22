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

package jenkins

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// JobPath represents a path to a job on the filesystem.
type JobPath string

type JobPathOpts struct {
	Root       string
	IgnoreList []string
}

// Parse loads the configuration for the job from disk and marshals it into a Job object.
func (jp JobPath) Parse() (Job, error) {
	job := Job{
		path: jp,
	}
	err := job.fetch()
	return job, err
}

// GetJobPaths recursively searches a given folder for jobs and puts the JobPaths associated with the discovered jobs on the resultChan channel.
func GetJobPaths(opts JobPathOpts, resultChan chan<- JobPath) error {
	defer close(resultChan)

	err := parseJobFolder(opts.Root, opts, resultChan)
	if err != nil {
		return err
	}

	return nil
}

func contains(needle string, haystack []string) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}

func parseJobFolder(path string, opts JobPathOpts, resultChan chan<- JobPath) error {

	if contains(filepath.Base(path), opts.IgnoreList) {
		return nil
	}

	childErr := parseChildJobs(path, opts, resultChan)
	buildErr := parseBuildPath(path, resultChan)

	if childErr != nil && buildErr != nil {
		// Check if config.xml file exists, if so it's an empty Jenkins folder, which we don't care about
		_, err := os.Stat(filepath.Join(path, "config.xml"))
		if err != nil {
			return fmt.Errorf("parsing paths failed: %v, %v", childErr, buildErr)
		}
	}

	return nil
}

func parseChildJobs(path string, opts JobPathOpts, resultChan chan<- JobPath) error {
	jobsPath := filepath.Join(path, "jobs")

	_, err := os.Stat(jobsPath)
	if err != nil {
		return err
	}

	jobDirs, err := ioutil.ReadDir(jobsPath)
	if err != nil {
		return err
	}

	for _, jobDir := range jobDirs {
		if !jobDir.IsDir() {
			continue
		}

		err = parseJobFolder(filepath.Join(path, "jobs", jobDir.Name()), opts, resultChan)
		if err != nil {
			return err
		}
	}

	return nil
}

func parseBuildPath(path string, resultChan chan<- JobPath) error {
	buildsPath := filepath.Join(path, "builds")

	_, err := os.Stat(buildsPath)
	if err != nil {
		return err
	}

	resultChan <- JobPath(path)

	return nil
}
