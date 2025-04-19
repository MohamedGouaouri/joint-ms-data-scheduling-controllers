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

type EdgeNetworkDef struct {
	Kind       string `json:"kind,omitempty"`
	ApiVersion string `json:"apiVersion,omitempty"`
	Namespace  string `json:"namespace,omitempty"`
}
type Microservice struct {
	Name          string                   `json:"name,omitempty"`
	DeploymentRef string                   `json:"deploymentRef,omitempty"`
	Namespace     string                   `json:"namespace,omitempty"`
	VolumeClaim   string                   `json:"volumeClaim,omitempty"`
	Dependencies  []MicroserviceDependency `json:"dependencies,omitempty"`
}
type MicroserviceDependency struct {
	MicroserviceRef string `json:"microserviceRef,omitempty"`
	MinBandwidth    int    `json:"minBandwidth,omitempty"`
}

// MicroserviceApplicationSpec defines the desired state of MicroserviceApplication.
type MicroserviceApplicationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of MicroserviceApplication. Edit microserviceapplication_types.go to remove/update
	// Foo string `json:"foo,omitempty"`
	EdgeNetwork   EdgeNetworkDef `json:"edgeNetwork,omitempty"`
	Microservices []Microservice `json:"microservices,omitempty"`
}

// MicroserviceApplicationStatus defines the observed state of MicroserviceApplication.
type MicroserviceApplicationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// TODO: Add information about node ranks
	Ranks map[string]int `json:"ranks,omitempty"` // The ranks are calculated by the controller

}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// MicroserviceApplication is the Schema for the microserviceapplications API.
type MicroserviceApplication struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MicroserviceApplicationSpec   `json:"spec,omitempty"`
	Status MicroserviceApplicationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MicroserviceApplicationList contains a list of MicroserviceApplication.
type MicroserviceApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MicroserviceApplication `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MicroserviceApplication{}, &MicroserviceApplicationList{})
}

type Neighbour struct {
	Microservice Microservice
	Dependency   MicroserviceDependency
}

func (app *MicroserviceApplication) Neighbours(ms Microservice) []Neighbour {
	// Find microservice in the application
	// Return its dependencies
	var neighbours []Neighbour = make([]Neighbour, 0)
	for _, dependency := range ms.Dependencies {
		neighbourMs, found := app.FindMicroservice(dependency.MicroserviceRef)
		if found {
			neighbours = append(neighbours, Neighbour{
				Dependency:   dependency,
				Microservice: neighbourMs,
			})
		}

	}
	return neighbours
}

func (app *MicroserviceApplication) FindMicroservice(name string) (Microservice, bool) {
	for _, looper := range app.Spec.Microservices {
		if looper.Name == name {
			return looper, true
		}
	}
	return Microservice{}, false
}
