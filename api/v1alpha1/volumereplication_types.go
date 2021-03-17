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

// ReplicationState represents the replication operations to be performed on the volume
type ReplicationState string

const (
	// Primary ReplicationState enables mirroring and promotes the volume to primary
	Primary ReplicationState = "primary"

	// Secondary ReplicationState demotes the volume to secondary and resyncs the volume if out of sync
	Secondary ReplicationState = "secondary"

	// Resync option resyncs the volume
	Resync ReplicationState = "resync"
)

// State captures the latest state of the replication operation
type State string

const (
	// Replicating means the volume is mirroring or replicating
	Replicating State = "Replicating"

	// ReplicationFailure means the last replication operation failed
	ReplicationFailure State = "Failed"

	//Resyncing means that the volume is resyncing
	Resyncing State = "Resyncing"
)

// VolumeReplicationSpec defines the desired state of VolumeReplication
type VolumeReplicationSpec struct {
	// VolumeReplicationClass is the VolumeReplicationClass name for this VolumeReplication resource
	// +kubebuilder:validation:Required
	VolumeReplicationClass string `json:"volumeReplicationClass"`

	// ReplicationState represents the replication operation to be performed on the volume.
	// Supported operations are "primary", "secondary" and "resync"
	// +kubebuilder:validation:Required
	ReplicationState ReplicationState `json:"replicationState"`

	// DataSource represents the object associated with the volume
	// +kubebuilder:validation:Required
	DataSource corev1.TypedLocalObjectReference `json:"dataSource"`
}

// VolumeReplicationStatus defines the observed state of VolumeReplication
type VolumeReplicationStatus struct {
	State              State       `json:"state,omitempty"`
	Message            string      `json:"message,omitempty"`
	LastStartTime      metav1.Time `json:"lastStartTime,omitempty"`
	LastCompletionTime metav1.Time `json:"lastCompletionTime,omitempty"`
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
