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

package v1alpha1

import (
	"github.com/stretchr/testify/assert"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"testing"
)

func TestDefaultSpecNoNils(t *testing.T) {
	var a = Agent{}
	a.Default()
	assert.NotNil(t, a.Spec.Realm)
	assert.NotNil(t, a.Spec.ClusterName)
	assert.NotNil(t, a.Spec.Agent)
	assert.NotNil(t, a.Spec.ClusterReceiver)
	assert.NotNil(t, a.Spec.Gateway)
	assert.NotNil(t, a.Spec.Instrumentation)
}

func TestDefaultResourceLimits(t *testing.T) {
	type testCase struct {
		in      string
		outSpec resource.Quantity
		outEnv  []v1.EnvVar
	}
	var a = Agent{}
	a.Default()
	testCases := []testCase{
		{in: defaultAgentCPU,
			outSpec: a.Spec.Agent.Resources.Limits[v1.ResourceCPU]},
		{in: defaultAgentMemory,
			outSpec: a.Spec.Agent.Resources.Limits[v1.ResourceMemory],
			outEnv: a.Spec.Agent.Env},
		{in: defaultClusterReceiverCPU,
			outSpec: a.Spec.ClusterReceiver.Resources.Limits[v1.ResourceCPU]},
		{in: defaultClusterReceiverMemory,
			outSpec: a.Spec.ClusterReceiver.Resources.Limits[v1.ResourceMemory],
			outEnv: a.Spec.ClusterReceiver.Env},
		{in: defaultGatewayCPU,
			outSpec: a.Spec.Gateway.Resources.Limits[v1.ResourceCPU]},
		{in: defaultGatewayMemory,
			outSpec: a.Spec.Gateway.Resources.Limits[v1.ResourceMemory],
			outEnv: a.Spec.Gateway.Env},
	}
	for _, c := range testCases {
		assert.Equal(t, resource.MustParse(c.in), c.outSpec)
		if c.outEnv != nil {
			found := false
			for _, i := range c.outEnv {
				if i.Name == "SPLUNK_MEMORY_TOTAL_MIB" {
					found = true
					assert.Equal(t, i.Value,
						getMemSizeInMiB(resource.MustParse(c.in)))
				}
			}
			assert.True(t, found,
				"SPLUNK_MEMORY_TOTAL_MIB is absent from the env variables")
		}
	}
}

func TestGetMemSizeInMiB(t *testing.T) {
	type testCase struct {
		in  string
		out string
	}
	testCases := []testCase{
		{in: "1048575", out: "0"},
		{in: "1048576", out: "1"},
		{in: "1023Ki", out: "0"},
		{in: "1024Ki", out: "1"},
		{in: "1Mi", out: "1"},
		{in: "1Gi", out: "1024"},
		{in: "1Ti", out: "1048576"},
		{in: "1Pi", out: "1073741824"},
		{in: "1Ei", out: "1099511627776"},
		{in: "1048k", out: "0"},
		{in: "1049k", out: "1"},
		{in: "1M", out: "0"},
		{in: "2M", out: "1"},
		{in: "1G", out: "953"},
		{in: "1T", out: "953674"},
		{in: "1P", out: "953674316"},
		{in: "1E", out: "953674316406"},
	}
	for _, c := range testCases {
		assert.Equal(t, getMemSizeInMiB(resource.MustParse(c.in)), c.out)
	}
}
