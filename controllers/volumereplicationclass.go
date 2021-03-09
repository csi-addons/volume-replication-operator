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

	replicationv1alpha1 "github.com/kube-storage/volume-replication-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func (r VolumeReplicationReconciler) getVolumeReplicaCLass(vrcName string) (*replicationv1alpha1.VolumeReplicationClass, error) {
	vrcObj := &replicationv1alpha1.VolumeReplicationClass{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: vrcName}, vrcObj)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Error(err, "VolumeReplicationClass not found", "VolumeReplicationClass", vrcName)
		} else {
			r.Log.Error(err, "Got an unexpected error while fetching VolumeReplicationClass", "VolumeReplicationClass", vrcName)
		}
		return nil, err
	}

	return vrcObj, nil
}
