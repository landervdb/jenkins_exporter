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
	"testing"
)

var (
	paths = []string{
		"testdata/jobs/rootjob",
		"testdata/jobs/folder/jobs/failedjob",
		"testdata/jobs/folder/jobs/folderjob",
		"testdata/jobs/folder/jobs/jobwithoutbuilds",
	}
)

func TestGetJobPaths(t *testing.T) {
	resultChan := make(chan JobPath)

	go GetJobPaths("testdata", resultChan)

	i := 0

	for path := range resultChan {
		if !checkPath(string(path)) {
			t.Errorf("Path '%s' should not be included in results!", path)
		}
		i++
	}

	if i != len(paths) {
		t.Error("Not all paths were present in results!")
	}
}

func checkPath(path string) bool {
	for _, p := range paths {
		if p == path {
			return true
		}
	}
	return false
}
