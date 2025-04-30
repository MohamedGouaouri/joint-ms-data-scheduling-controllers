/*
Copyright 2025.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type Allocation struct {
	Node string `json:"node,omitempty" protobuf:"bytes,1,opt,name=node"`
	Size int    `json:"size,omitempty" protobuf:"bytes,2,opt,name=size"`
}

// VolumeAllocationSpec defines the desired state of VolumeAllocation.
type VolumeAllocationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// Note: Storage class must be set to our custom sotrage class for proper storage allocation
	StorageClassName      string `json:"storageClassName,omitempty" protobuf:"bytes,1,opt,name=storageClassName"`
	Microservice          string `json:"microservice,omitempty" protobuf:"bytes,2,opt,name=microservice"`
	MicroservicePlacement string `json:"microservicePlacement,omitempty" protobuf:"bytes,3,opt,name=microservicePlacement"`
	VolumeSize            string `json:"volumeSize,omitempty" protobuf:"bytes,4,opt,name=volumeSize"`
	EdgeNetworkTopology   string `json:"edgeNetworkTopology,omitempty" protobuf:"bytes,5,opt,name=edgeNetworkTopology"`
}

// VolumeAllocationStatus defines the observed state of VolumeAllocation.
type VolumeAllocationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// VolumeAllocation is the Schema for the volumeallocations API.
type VolumeAllocation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   VolumeAllocationSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status VolumeAllocationStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// +kubebuilder:object:root=true

// VolumeAllocationList contains a list of VolumeAllocation.
type VolumeAllocationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VolumeAllocation `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VolumeAllocation{}, &VolumeAllocationList{})
}
