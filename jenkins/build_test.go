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

import "testing"

var (
	testbuild = "testdata/jobs/rootjob/builds/lastSuccessfulBuild"
)

func TestBuildParse(t *testing.T) {
	var build Build

	build, err := parseBuild(testbuild)
	if err != nil {
		t.Error(err)
	}

	if build.Number != 1 {
		t.Errorf("build.Number is %d, expected %d", build.Number, 1)
	}

	if build.Result != "SUCCESS" {
		t.Errorf("build.Result is %s, expected %s", build.Result, "SUCCESS")
	}

	if build.Timestamp != 1548791914210 {
		t.Errorf("build.Timestamp is %d, expected %d", build.Timestamp, 1548791914210)
	}

	if build.Duration != 49 {
		t.Errorf("build.Duration is %d, expected %d", build.Duration, 49)
	}

	if build.EnvVars["BUILD_DISPLAY_NAME"] != "#1" {
		t.Errorf("build.EnvVars['BUILD_DISPLAY_NAME'] is %s, expected %s", build.EnvVars["BUILD_DISPLAY_NAME"], "#1")
	}
}
