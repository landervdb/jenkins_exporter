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
	"encoding/xml"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Build represents a particular build of a Job.
type Build struct {
	raw       buildXML
	Number    int
	Timestamp int
	Duration  int
	Result    string
	EnvVars   map[string]string
}

type buildXML struct {
	XMLName   xml.Name   `xml:"build"`
	Result    string     `xml:"result"`
	Timestamp int        `xml:"timestamp"`
	Duration  int        `xml:"duration"`
	Number    int        `xml:"number"`
	Actions   actionsXML `xml:"actions"`
}

type actionsXML struct {
	XMLName          xml.Name            `xml:"actions"`
	BuildEnvironment buildEnvironmentXML `xml:"org.jenkinsci.plugins.buildenvironment.actions.BuildEnvironmentBuildAction"`
}

type buildEnvironmentXML struct {
	XMLName     xml.Name       `xml:"org.jenkinsci.plugins.buildenvironment.actions.BuildEnvironmentBuildAction"`
	DataHolders dataHoldersXML `xml:"dataHolders"`
}

type dataHoldersXML struct {
	XMLName xml.Name   `xml:"dataHolders"`
	EnvVars envVarsXML `xml:"org.jenkinsci.plugins.buildenvironment.data.EnvVarsData"`
}

type envVarsXML struct {
	XMLName xml.Name       `xml:"org.jenkinsci.plugins.buildenvironment.data.EnvVarsData"`
	Data    envVarsDataXML `xml:"data"`
}

type envVarsDataXML struct {
	XMLName xml.Name    `xml:"data"`
	Entries []envVarXML `xml:"entry"`
}

type envVarXML struct {
	XMLName xml.Name `xml:"entry"`
	Values  []string `xml:"string"`
}

func (build *buildXML) getEnvVars() map[string]string {
	envVars := make(map[string]string)
	for _, e := range build.Actions.BuildEnvironment.DataHolders.EnvVars.Data.Entries {
		envVars[e.Values[0]] = e.Values[1]
	}
	return envVars
}

func newBuildFromXML(path string) (Build, error) {
	var build Build

	xmlFile, err := os.Open(path)
	if err != nil {
		return build, err
	}
	defer xmlFile.Close()

	byteValue, err := ioutil.ReadAll(xmlFile)
	if err != nil {
		return build, err
	}

	err = xml.Unmarshal(forceXMLVersion(byteValue), &build.raw)
	if err != nil {
		return build, err
	}

	build.EnvVars = build.raw.getEnvVars()
	buildNumber, err := strconv.ParseInt(build.EnvVars["BUILD_NUMBER"], 10, 64)
	if err != nil {
		return build, err
	}
	build.Number = int(buildNumber)
	build.Timestamp = build.raw.Timestamp
	build.Duration = build.raw.Duration
	build.Result = build.raw.Result

	return build, nil
}

func parseBuild(path string) (Build, error) {
	var build Build

	if _, err := os.Stat(path); err != nil {
		return build, err
	}

	xmlPath := filepath.Join(path, "build.xml")

	build, err := newBuildFromXML(xmlPath)
	if err != nil {
		return build, err
	}

	return build, nil
}

// forceXMLVersion is a dirty hack because the Go XML parser does not support XML 1.1 at this point, and newer version of Jenkins do output this.
func forceXMLVersion(buf []byte) []byte {
	str := string(buf)
	ret := strings.Replace(str, "<?xml version='1.1' encoding='UTF-8'?>", "<?xml version='1.0' encoding='UTF-8'?>", 1)
	return []byte(ret)
}
