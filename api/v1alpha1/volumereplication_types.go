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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ImageState represents the replication operations to be performed on the image
type ImageState string

const (
	// Primary ImageState enables mirroring and promotes the image to primary
	Primary ImageState = "primary"

	// Secondary ImageState demotes the image to secondary and resyncs the image if out of sync
	Secondary ImageState = "secondary"

	// Resync option resyncs the image
	Resync ImageState = "resync"
)

// ReplicationState captures the latest state of the replication operation
type ReplicationState string

const (
	// Replicating means the image is mirroring or replicating
	Replicating ReplicationState = "Replicating"

	// ReplicationFailure means the last replication operation failed
	ReplicationFailure ReplicationState = "Failed"
)

// VolumeReplicationSpec defines the desired state of VolumeReplication
type VolumeReplicationSpec struct {
	// VolumeReplicationClass is the VolumeReplicationClass name for this VolumeReplication resource
	// +kubebuilder:validation:Required
	VolumeReplicationClass string `json:"volumeReplicationClass"`

	// ImageState represents the replication operation to be performed on the image.
	// Supported operations are "primary", "secondary" and "resync"
	// +kubebuilder:validation:Required
	ImageState ImageState `json:"imageState"`

	// DataSource represents the object associated with the image
	// +kubebuilder:validation:Required
	DataSource corev1.TypedLocalObjectReference `json:"dataSource"`
}

// VolumeReplicationStatus defines the observed state of VolumeReplication
type VolumeReplicationStatus struct {
	State              ReplicationState `json:"state,omitempty"`
	Message            string           `json:"message,omitempty"`
	LastStartTime      metav1.Time      `json:"lastStartTime,omitempty"`
	LastCompletionTime metav1.Time      `json:"lastCompletionTime,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// VolumeReplication is the Schema for the volumereplications API
type VolumeReplication struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VolumeReplicationSpec   `json:"spec,omitempty"`
	Status VolumeReplicationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// VolumeReplicationList contains a list of VolumeReplication
type VolumeReplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VolumeReplication `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VolumeReplication{}, &VolumeReplicationList{})
}
