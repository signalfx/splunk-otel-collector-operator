// Copyright The OpenTelemetry Authors
// Copyright Splunk Inc
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

package autodetect

// Distro holds the auto-detected distro type.
type Distro int

const (
	// Unknown is used when the current distro can't be determined.
	UnknownDistro Distro = iota

	// OpenShift represents a distro of type OpenShift.
	OpenShiftDistro Distro = iota

	// Kubernetes represents a distro of type Kubernetes.
	KubernetesDistro
)

func (d Distro) String() string {
	return [...]string{"Unknown", "OpenShift", "Kubernetes"}[d]
}
