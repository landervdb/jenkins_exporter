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
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// JobPath represents a path to a job on the filesystem.
type JobPath string

// Parse loads the configuration for the job from disk and marshals it into a Job object.
func (jp JobPath) Parse() (Job, error) {
	job := Job{
		path: jp,
	}
	err := job.fetch()
	return job, err
}

// GetJobPaths recursively searches a given folder for jobs and puts the JobPaths associated with the discovered jobs on the resultChan channel.
func GetJobPaths(path string, resultChan chan<- JobPath) error {
	defer close(resultChan)

	err := parseJobFolder(path, resultChan)
	if err != nil {
		log.Println(err)
	}

	return nil
}

func parseJobFolder(path string, resultChan chan<- JobPath) error {

	parseChildJobs(path, resultChan)

	buildsPath := filepath.Join(path, "builds")

	_, err := os.Stat(buildsPath)
	if err == nil {
		resultChan <- JobPath(path)
	}

	return nil
}

func parseChildJobs(path string, resultChan chan<- JobPath) error {
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

		err = parseJobFolder(filepath.Join(path, "jobs", jobDir.Name()), resultChan)
		if err != nil {
			continue
		}
	}

	return nil
}
