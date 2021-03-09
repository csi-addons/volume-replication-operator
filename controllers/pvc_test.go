/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"testing"

	replicationv1alpha1 "github.com/kube-storage/volume-replication-operator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	mockPVName       = "test-pv"
	mockPVCName      = "test-pvc"
	mockNamespace    = "test-ns"
	mockVolumeHandle = "test-volume-handle"
)

var mockVolumeReplicationObj = &replicationv1alpha1.VolumeReplication{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "volume-replication",
		Namespace: mockNamespace,
	},
	Spec: replicationv1alpha1.VolumeReplicationSpec{
		DataSource: corev1.TypedLocalObjectReference{
			Name: mockPVCName,
		},
	},
}

var mockPersistentVolume = &corev1.PersistentVolume{
	ObjectMeta: metav1.ObjectMeta{
		Name: mockPVName,
	},
	Spec: corev1.PersistentVolumeSpec{
		PersistentVolumeSource: corev1.PersistentVolumeSource{
			CSI: &corev1.CSIPersistentVolumeSource{
				VolumeHandle: mockVolumeHandle,
			},
		},
	},
}

var mockPersistentVolumeClaim = &corev1.PersistentVolumeClaim{
	ObjectMeta: metav1.ObjectMeta{
		Name:      mockPVCName,
		Namespace: mockNamespace,
	},
	Spec: corev1.PersistentVolumeClaimSpec{
		VolumeName: mockPVName,
	},
	Status: corev1.PersistentVolumeClaimStatus{
		Phase: corev1.ClaimBound,
	},
}

func createFakeScheme(t *testing.T) *runtime.Scheme {
	scheme, err := replicationv1alpha1.SchemeBuilder.Build()
	if err != nil {
		assert.Fail(t, "unable to build scheme")
	}
	err = corev1.AddToScheme(scheme)
	if err != nil {
		assert.Fail(t, "failed to add corev1 scheme")
	}
	err = replicationv1alpha1.AddToScheme(scheme)
	if err != nil {
		assert.Fail(t, "failed to add replicationv1alpha1 scheme")
	}
	return scheme
}

func createFakeVolumeReplicationReconciler(t *testing.T, obj ...runtime.Object) VolumeReplicationReconciler {
	scheme := createFakeScheme(t)
	client := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(obj...).Build()

	return VolumeReplicationReconciler{
		Client: client,
		Scheme: scheme,
		Log:    logf.Log.WithName("controller_volumereplication_test"),
	}
}
func TestGetVolumeHandle(t *testing.T) {
	testcases := []struct {
		name                 string
		pv                   *corev1.PersistentVolume
		pvc                  *corev1.PersistentVolumeClaim
		expectedVolumeHandle string
		errorExpected        bool
	}{
		{
			name:                 "case 1: volume handle available",
			pv:                   mockPersistentVolume,
			pvc:                  mockPersistentVolumeClaim,
			expectedVolumeHandle: mockVolumeHandle,
			errorExpected:        false,
		},
		{
			name: "case 2: pvc name in VolumeReplication CR not found",
			pv:   mockPersistentVolume,
			pvc: &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pvc-name",
					Namespace: mockNamespace,
				},
			},
			expectedVolumeHandle: mockVolumeHandle,
			errorExpected:        true,
		},
		{
			name: "case 3: pvc not bound",
			pv:   mockPersistentVolume,
			pvc: &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      mockPVCName,
					Namespace: mockNamespace,
				},
			},
			expectedVolumeHandle: mockVolumeHandle,
			errorExpected:        true,
		},
	}

	for _, tc := range testcases {
		volumeReplication := &replicationv1alpha1.VolumeReplication{}
		mockVolumeReplicationObj.DeepCopyInto(volumeReplication)

		testPV := &corev1.PersistentVolume{}
		tc.pv.DeepCopyInto(testPV)

		testPVC := &corev1.PersistentVolumeClaim{}
		tc.pvc.DeepCopyInto(testPVC)

		namespacedName := types.NamespacedName{Name: mockPVCName,
			Namespace: volumeReplication.Namespace}

		reconciler := createFakeVolumeReplicationReconciler(t, testPV, testPVC, volumeReplication)
		resultPVC, resultPV, err := reconciler.getPVCDataSource(namespacedName)
		if tc.errorExpected {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.NotEqual(t, nil, resultPVC)
			assert.NotEqual(t, nil, resultPV)
			assert.Equal(t, tc.expectedVolumeHandle, resultPV.Spec.CSI.VolumeHandle)
		}
	}
}
