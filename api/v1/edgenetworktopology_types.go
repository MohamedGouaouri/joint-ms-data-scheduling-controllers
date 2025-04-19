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

type EdgeNode struct {
	Name  string         `json:"name,omitempty"`
	IP    string         `json:"ip,omitempty"`
	Links []EdgeNodeLink `json:"links,omitempty"`
}
type EdgeNodeLink struct {
	EdgeNodeRef       string `json:"edgeNodeRef,omitempty"`
	BandwidthCapacity int    `json:"bandwidthCapacity,omitempty"`
	InterfaceName     string `json:"interfaceName,omitempty"`
	InterfaceIP       string `json:"interfaceIp,omitempty"`
}

// EdgeNetworkTopologySpec defines the desired state of EdgeNetworkTopology.
type EdgeNetworkTopologySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Edges []EdgeNode `json:"edges,omitempty"`
}

// EdgeNetworkTopologyStatus defines the observed state of EdgeNetworkTopology.
type EdgeNetworkTopologyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// EdgeNetworkTopology is the Schema for the edgenetworktopologies API.
type EdgeNetworkTopology struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EdgeNetworkTopologySpec   `json:"spec,omitempty"`
	Status EdgeNetworkTopologyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// EdgeNetworkTopologyList contains a list of EdgeNetworkTopology.
type EdgeNetworkTopologyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EdgeNetworkTopology `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EdgeNetworkTopology{}, &EdgeNetworkTopologyList{})
}
