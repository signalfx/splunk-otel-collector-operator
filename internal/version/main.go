// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package version contains the operator's version, as well as versions of underlying components.
package version

import (
	"fmt"
	"runtime"
)

var (
	version          string
	buildDate        string
	collectorVersion string
)

// Version holds this Operator's version as well as the version of some of the components it uses.
type Version struct {
	Operator  string `json:"splunk-otel-operator"`
	BuildDate string `json:"build-date"`
	Collector string `json:"splunk-otel-collector-version"`
	Go        string `json:"go-version"`
}

// Get returns the Version object with the relevant information.
func Get() Version {
	return Version{
		Operator:  version,
		BuildDate: buildDate,
		Collector: Collector(),
		Go:        runtime.Version(),
	}
}

func (v Version) String() string {
	return fmt.Sprintf(
		"Version(Operator='%v', BuildDate='%v', SplunkOtelCollector='%v', Go='%v')",
		v.Operator,
		v.BuildDate,
		v.Collector,
		v.Go,
	)
}

// Collector returns the default SplunkOtelAgent to use when no versions are specified via CLI or configuration.
func Collector() string {
	if len(collectorVersion) > 0 {
		// this should always be set, as it's specified during the build
		return collectorVersion
	}

	// fallback value, useful for tests
	return "0.0.0"
}
