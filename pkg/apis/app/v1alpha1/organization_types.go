package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type Module struct {
	Source  string `json:"source"`
	Version string `json:"version"`
}

type Variable struct {
	// Variable name
	Key string `json:"key"`
	// Variable value
	// +optional
	Value string `json:"value"`
	// VolumeMount if secret
	// +optional
	VolumeMount corev1.VolumeMount `json:"volumeMount"`
}

// OrganizationSpec defines the desired state of Organization
// +k8s:openapi-gen=true
type OrganizationSpec struct {
	// Module source and version to use
	Module *Module `json:"module"`
	// Variables as inputs to module
	// +listType=set
	// +optional
	Variables []*Variable `json:"variables"`
	// Volumes for sensitive variables
	// +listType=set
	// +optional
	Volumes []*corev1.Volume `json:"volumes"`
}

// OrganizationStatus defines the observed state of Organization
// +k8s:openapi-gen=true
type OrganizationStatus struct {
	// Workspace ID
	WorkspaceID string `json:"workspaceID"`
	// Run ID
	RunID string `json:"runID"`
	// Configuration hash
	ConfigHash string `json:"configHash"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Organization is the Schema for the organizations API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=organizations,scope=Namespaced
type Organization struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OrganizationSpec   `json:"spec,omitempty"`
	Status OrganizationStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OrganizationList contains a list of Organization
type OrganizationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Organization `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Organization{}, &OrganizationList{})
}
