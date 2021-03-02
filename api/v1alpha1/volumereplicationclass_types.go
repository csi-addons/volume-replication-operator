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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// VolumeReplicationClassSpec specifies parameters that a underlying storage system uses when creating a volume replica. A specific VolumeReplicationClass is used by specifying its name in a VolumeReplicationRequest object.
type VolumeReplicationClassSpec struct {
	// driver is the name of the storage driver that handles this VolumeReplicationClass.
	// Required
	Driver string `json:"driver" protobuf:"bytes,2,opt,name=driver"`

	// parameters is a key-value map with storage driver specific parameters for creating the volume replicas.
	// These values are opaque to Kubernetes.
	// +optional
	Parameters map[string]string `json:"parameters,omitempty" protobuf:"bytes,3,rep,name=parameters"`
}

// VolumeReplicationClassStatus defines the observed state of VolumeReplicationClass
type VolumeReplicationClassStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// VolumeReplicationClass specifies parameters that a underlying storage system
// uses when creating a volume replica. A specific VolumeReplicationClass is
// used by specifying its name in a VolumeReplicationRequest object.
// VolumeReplicationClasses are non-namespaced
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolume:name="Driver",type=string,JSONPath=`.driver`
// +kubebuilder:printcolume:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type VolumeReplicationClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VolumeReplicationClassSpec   `json:"spec,omitempty"`
	Status VolumeReplicationClassStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// VolumeReplicationClassList contains a list of VolumeReplicationClass
type VolumeReplicationClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VolumeReplicationClass `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VolumeReplicationClass{}, &VolumeReplicationClassList{})
}
