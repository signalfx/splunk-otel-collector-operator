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

package webhooks

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestCreateResourceMap(t *testing.T) {
	cases := []struct {
		containerName string
		podName       string
		nodeName      string
		uid           string
		namespace     string
		env           []corev1.EnvVar
		attrs         map[string]string
		idx           int
	}{
		{
			containerName: "test-container1",
			namespace:     "test-namespace",
			podName:       "test-pod",
			nodeName:      "test-node",
			uid:           "12345",
			attrs: map[string]string{
				"k8s.container.name": "test-container1",
				"k8s.namespace.name": "test-namespace",
				"k8s.node.name":      "test-node",
				"k8s.pod.name":       "test-pod",
				"k8s.pod.uid":        "12345",
			},
			idx: -1,
		},
		{
			containerName: "test-container2",
			env: []corev1.EnvVar{
				{
					Name:  "env2",
					Value: "value2",
				},
				{
					Name:  envOTELResourceAttrs,
					Value: "k1=v1,k2=v2,k3=v3=v4",
				},
			},
			attrs: map[string]string{
				"k1":                 "v1",
				"k2":                 "v2",
				"k8s.container.name": "test-container2",
			},
			idx: 1,
		},
	}

	h := &handler{
		logger: logr.DiscardLogger{},
	}

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			pod := corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: c.podName,
					UID:  types.UID(c.uid),
				},
				Spec: corev1.PodSpec{
					NodeName: c.nodeName,
					Containers: []corev1.Container{{
						Name: c.containerName,
						Env:  c.env,
					}},
				},
			}
			ns := corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: c.namespace}}
			attrs, idx := h.createResourceMap(context.Background(), ns, pod)
			assert.Equal(t, attrs, c.attrs)
			assert.Equal(t, idx, c.idx)
		})
	}
}
