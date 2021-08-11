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
	"context"
	"fmt"

	replicationv1alpha1 "github.com/csi-addons/volume-replication-operator/api/v1alpha1"

	"github.com/go-logr/logr"
)

const (
	volumeReplicationFinalizer = "replication.storage.openshift.io"
)

// addFinalizerToVR adds the VR finalizer on the VolumeReplication instance.
func (r *VolumeReplicationReconciler) addFinalizerToVR(logger logr.Logger, vr *replicationv1alpha1.VolumeReplication,
) error {
	if !contains(vr.ObjectMeta.Finalizers, volumeReplicationFinalizer) {
		logger.Info("adding finalizer to volumeReplication object", "Finalizer", volumeReplicationFinalizer)
		vr.ObjectMeta.Finalizers = append(vr.ObjectMeta.Finalizers, volumeReplicationFinalizer)
		if err := r.Client.Update(context.TODO(), vr); err != nil {
			return fmt.Errorf("failed to add finalizer (%s) to VolumeReplication resource"+
				" (%s/%s) %w",
				volumeReplicationFinalizer, vr.Namespace, vr.Name, err)
		}
	}

	return nil
}

// removeFinalizerFromVR removes the VR finalizer from the VolumeReplication instance.
func (r *VolumeReplicationReconciler) removeFinalizerFromVR(logger logr.Logger, vr *replicationv1alpha1.VolumeReplication) error {
	if contains(vr.ObjectMeta.Finalizers, volumeReplicationFinalizer) {
		logger.Info("removing finalizer from volumeReplication object", "Finalizer", volumeReplicationFinalizer)
		vr.ObjectMeta.Finalizers = remove(vr.ObjectMeta.Finalizers, volumeReplicationFinalizer)
		if err := r.Client.Update(context.TODO(), vr); err != nil {
			return fmt.Errorf("failed to remove finalizer (%s) from VolumeReplication resource"+
				" (%s/%s), %w",
				volumeReplicationFinalizer, vr.Namespace, vr.Name, err)
		}
	}

	return nil
}
