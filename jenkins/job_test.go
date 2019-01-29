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
	rootfolderJob    = "testdata/jobs/rootjob"
	successfulJob    = "testdata/jobs/folder/jobs/folderjob"
	failedJob        = "testdata/jobs/folder/jobs/failedjob"
	jobWithoutBuilds = "testdata/jobs/folder/jobs/jobwithoutbuilds"
	nonExistentJob   = "testdata/jobs/foobar"
	folder           = "testdata/jobs/folder"
)

func TestJobFetch(t *testing.T) {
	job := Job{
		path: JobPath(successfulJob),
	}

	err := job.fetch()
	if err != nil {
		t.Error(err)
	}

	if job.Name != "folderjob" {
		t.Errorf("job.Name is %s, expected %s", job.Name, "folderjob")
	}

	if job.Folder != "folder" {
		t.Errorf("job.Folder is %s, expected %s", job.Folder, "folder")
	}
}

func TestSuccessfulJob(t *testing.T) {
	job := Job{
		path: JobPath(successfulJob),
	}

	err := job.fetch()
	if err != nil {
		t.Error(err)
	}

	if job.LastBuild.Result != "SUCCESS" {
		t.Errorf("job.LastBuild.Result is %s, expected %s", job.LastBuild.Result, "SUCCESS")
	}

	if job.LastBuild.Number != job.LastSuccessfulBuild.Number {
		t.Error("job.LastBuild.Number != job.LastSuccessfulBuild.Number")
	}
}

func TestFailedJob(t *testing.T) {
	job := Job{
		path: JobPath(failedJob),
	}

	err := job.fetch()
	if err != nil {
		t.Error(err)
	}

	if job.LastBuild.Result != "FAILURE" {
		t.Errorf("job.LastBuild.Result is %s, expected %s", job.LastBuild.Result, "FAILURE")
	}

	if job.LastBuild.Number != job.LastFailedBuild.Number {
		t.Error("job.LastBuild.Number != job.LastFailedBuild.Number")
	}
}

func TestJobWithoutBuilds(t *testing.T) {
	job := Job{
		path: JobPath(jobWithoutBuilds),
	}

	err := job.fetch()
	if err == nil {
		t.Error("job without builds should return an error when parsed")
	}
}

func TestNonExistentJob(t *testing.T) {
	job := Job{
		path: JobPath(nonExistentJob),
	}

	err := job.fetch()
	if err == nil {
		t.Error("non existent job should return an error when parsed")
	}
}

func TestRootFolderJob(t *testing.T) {
	job := Job{
		path: JobPath(rootfolderJob),
	}

	err := job.fetch()
	if err != nil {
		t.Error(err)
	}

	if job.Folder != "/" {
		t.Errorf("job.Folder is %s, expected %s", job.Folder, "/")
	}
}

func TestFolderNotAJob(t *testing.T) {
	job := Job{
		path: JobPath(folder),
	}

	err := job.fetch()
	if err == nil {
		t.Error("trying to parse a folder as a job should return an error")
	}
}
