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
	"os"
	"path/filepath"
	"strings"
)

// Job represents a Jenkins job and contains its latest builds.
type Job struct {
	path                  JobPath
	Name                  string
	Folder                string
	LastBuild             Build
	LastSuccessfulBuild   Build
	LastUnsuccessfulBuild Build
	LastStableBuild       Build
	LastUnstableBuild     Build
	LastFailedBuild       Build
}

func (job *Job) fetch() error {
	buildsPath := filepath.Join(string(job.path), "builds")
	buildsDir, err := os.Stat(buildsPath)
	if err != nil {
		return err
	}

	if !buildsDir.IsDir() {
		return fmt.Errorf("%s is not a directory", buildsPath)
	}

	job.LastSuccessfulBuild, _ = parseBuild(filepath.Join(buildsPath, "lastSuccessfulBuild"))
	job.LastUnsuccessfulBuild, _ = parseBuild(filepath.Join(buildsPath, "lastUnsuccessfulBuild"))
	job.LastStableBuild, _ = parseBuild(filepath.Join(buildsPath, "lastStableBuild"))
	job.LastUnstableBuild, _ = parseBuild(filepath.Join(buildsPath, "lastUnstableBuild"))
	job.LastFailedBuild, _ = parseBuild(filepath.Join(buildsPath, "lastFailedBuild"))

	job.LastBuild, err = job.selectLastBuild()
	if err != nil {
		return err
	}

	tokens := strings.Split(job.LastBuild.EnvVars["JOB_NAME"], "/")
	job.Name = tokens[len(tokens)-1]
	if len(tokens) == 1 {
		job.Folder = "/"
	} else {
		job.Folder = strings.Join(tokens[:len(tokens)-1], "/")
	}

	return nil
}

func (job *Job) selectLastBuild() (Build, error) {
	var lastBuild Build
	var max = 0

	if job.LastSuccessfulBuild.Number > max {
		lastBuild = job.LastSuccessfulBuild
		max = job.LastSuccessfulBuild.Number
	}

	if job.LastStableBuild.Number > max {
		lastBuild = job.LastStableBuild
		max = job.LastStableBuild.Number
	}

	if job.LastUnsuccessfulBuild.Number > max {
		lastBuild = job.LastUnsuccessfulBuild
		max = job.LastUnsuccessfulBuild.Number
	}

	if job.LastUnstableBuild.Number > max {
		lastBuild = job.LastUnstableBuild
		max = job.LastUnstableBuild.Number
	}

	if job.LastFailedBuild.Number > max {
		lastBuild = job.LastFailedBuild
		max = job.LastFailedBuild.Number
	}

	if max == 0 {
		return lastBuild, fmt.Errorf("no builds found at %s", job.path)
	}

	return lastBuild, nil
}
