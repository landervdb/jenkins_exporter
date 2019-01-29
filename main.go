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
	"log"

	"github.com/landervdb/jenkins_exporter/jenkins"
)

func main() {

	jobPaths := make(chan jenkins.JobPath)
	go jenkins.GetJobPaths("/tmp/jenkins", jobPaths)

	for jobPath := range jobPaths {
		job, err := jobPath.Parse()
		if err != nil {
			continue
		}
		log.Printf("result: %s number: %d job: %s folder: %s", job.LastBuild.Result, job.LastBuild.Number, job.Name, job.Folder)
	}

}
