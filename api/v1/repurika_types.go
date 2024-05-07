/*
Copyright 2024.

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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RepurikaSpec defines the desired state of Repurika
type RepurikaSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=20
	// +kubebuilder:validation:ExclusiveMaximum=false
	// Size defines the number of pods deployed by the operator
	Size int32 `json:"size"`

	Selector *metav1.LabelSelector `json:"selector,omitempty"`

	Template corev1.PodTemplateSpec `json:"template,omitempty"`
}

// RepurikaStatus defines the observed state of Repurika
type RepurikaStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Pods []string `json:"pods,omitempty"`

	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Repurika is the Schema for the repurikas API
type Repurika struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RepurikaSpec   `json:"spec,omitempty"`
	Status RepurikaStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RepurikaList contains a list of Repurika
type RepurikaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Repurika `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Repurika{}, &RepurikaList{})
}
