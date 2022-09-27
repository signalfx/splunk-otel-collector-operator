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

package otel

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	k8sreconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/signalfx/splunk-otel-collector-operator/apis/otel/v1alpha1"
	"github.com/signalfx/splunk-otel-collector-operator/internal/collector/reconcile"
)

var logger = logf.Log.WithName("unit-tests")

func TestNewObjectsOnReconciliation(t *testing.T) {
	// prepare
	nsn := types.NamespacedName{Name: "my-instance", Namespace: "default"}
	reconciler := NewReconciler(logger, k8sClient, testScheme, nil)
	created := &v1alpha1.Agent{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nsn.Name,
			Namespace: nsn.Namespace,
		},
		Spec: v1alpha1.AgentSpec{},
	}
	err := k8sClient.Create(context.Background(), created)
	require.NoError(t, err)

	// test
	req := k8sreconcile.Request{
		NamespacedName: nsn,
	}
	_, err = reconciler.Reconcile(context.Background(), req)

	// verify
	require.NoError(t, err)

	// the base query for the underlying objects
	opts := []client.ListOption{
		client.InNamespace(nsn.Namespace),
		client.MatchingLabels(map[string]string{
			"app.kubernetes.io/instance":   fmt.Sprintf("%s.%s", nsn.Namespace, nsn.Name),
			"app.kubernetes.io/managed-by": "splunk-otel-collector-operator",
		}),
	}

	// verify that we have at least one object for each of the types we create
	// whether we have the right ones is up to the specific tests for each type
	{
		list := &corev1.ConfigMapList{}
		err = k8sClient.List(context.Background(), list, opts...)
		assert.NoError(t, err)
		assert.NotEmpty(t, list.Items)
	}
	{
		list := &corev1.ServiceAccountList{}
		err = k8sClient.List(context.Background(), list, opts...)
		assert.NoError(t, err)
		assert.NotEmpty(t, list.Items)
	}
	// TODO(splunk): forcibly disable this test until we add gateway support
	//{
	//	list := &corev1.ServiceList{}
	//	err = k8sClient.List(context.Background(), list, opts...)
	//	assert.NoError(t, err)
	//	assert.NotEmpty(t, list.Items)
	//}
	{
		list := &appsv1.DeploymentList{}
		err = k8sClient.List(context.Background(), list, opts...)
		assert.NoError(t, err)
		assert.NotEmpty(t, list.Items)
	}
	{
		list := &appsv1.DaemonSetList{}
		err = k8sClient.List(context.Background(), list, opts...)
		assert.NoError(t, err)
		assert.NotEmpty(t, list.Items)
	}

	// cleanup
	require.NoError(t, k8sClient.Delete(context.Background(), created))

}

func TestContinueOnRecoverableFailure(t *testing.T) {
	// prepare
	taskCalled := false
	reconciler := NewReconciler(logger, nil, nil, nil)
	reconciler.tasks = []Task{
		{
			Name: "should-fail",
			Do: func(context.Context, reconcile.Params) error {
				return errors.New("should fail")
			},
			BailOnError: false,
		},
		{
			Name: "should-be-called",
			Do: func(context.Context, reconcile.Params) error {
				taskCalled = true
				return nil
			},
		},
	}

	// test
	err := reconciler.RunTasks(context.Background(), reconcile.Params{})

	// verify
	assert.NoError(t, err)
	assert.True(t, taskCalled)
}

func TestBreakOnUnrecoverableError(t *testing.T) {
	// prepare
	taskCalled := false
	expectedErr := errors.New("should fail")
	nsn := types.NamespacedName{Name: "my-instance", Namespace: "default"}
	reconciler := NewReconciler(logger, k8sClient, scheme.Scheme, nil)
	reconciler.tasks = []Task{
		{
			Name: "should-fail",
			Do: func(context.Context, reconcile.Params) error {
				taskCalled = true
				return expectedErr
			},
			BailOnError: true,
		},
		{
			Name: "should-not-be-called",
			Do: func(context.Context, reconcile.Params) error {
				assert.Fail(t, "should not have been called")
				return nil
			},
		},
	}

	created := &v1alpha1.Agent{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nsn.Name,
			Namespace: nsn.Namespace,
		},
	}
	err := k8sClient.Create(context.Background(), created)
	require.NoError(t, err)

	// test
	req := k8sreconcile.Request{
		NamespacedName: nsn,
	}
	_, err = reconciler.Reconcile(context.Background(), req)

	// verify
	assert.Equal(t, expectedErr, err)
	assert.True(t, taskCalled)

	// cleanup
	assert.NoError(t, k8sClient.Delete(context.Background(), created))
}

func TestSkipWhenInstanceDoesNotExist(t *testing.T) {
	// prepare
	nsn := types.NamespacedName{Name: "non-existing-my-instance", Namespace: "default"}
	reconciler := NewReconciler(logger, k8sClient, scheme.Scheme, nil)
	reconciler.tasks = []Task{
		{
			Name: "should-not-be-called",
			Do: func(context.Context, reconcile.Params) error {
				assert.Fail(t, "should not have been called")
				return nil
			},
		},
	}

	// test
	req := k8sreconcile.Request{
		NamespacedName: nsn,
	}
	_, err := reconciler.Reconcile(context.Background(), req)

	// verify
	assert.NoError(t, err)
}
