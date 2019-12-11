package v1alpha1

import (
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
	// Variable is a secret and should be retrieved from file
	Sensitive bool `json:"sensitive"`
	// EnvironmentVariable denotes if this variable should be created as environment variable
	EnvironmentVariable bool `json:"environmentVariable"`
}

// WorkspaceSpec defines the desired state of Workspace
// +k8s:openapi-gen=true
type WorkspaceSpec struct {
	// Terraform Cloud organization
	Organization string `json:"organization"`
	// Module source and version to use
	Module *Module `json:"module"`
	// Variables as inputs to module
	// +listType=set
	// +optional
	Variables []*Variable `json:"variables"`
	// File path within operator pod to load workspace secrets
	SecretsMountPath string `json:"secretsMountPath"`
}

// WorkspaceStatus defines the observed state of Workspace
// +k8s:openapi-gen=true
type WorkspaceStatus struct {
	// Workspace ID
	WorkspaceID string `json:"workspaceID"`
	// Run ID
	RunID string `json:"runID"`
	// Configuration hash
	ConfigHash string `json:"configHash"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Workspace is the Schema for the workspaces API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=workspaces,scope=Namespaced
type Workspace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkspaceSpec   `json:"spec,omitempty"`
	Status WorkspaceStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkspaceList contains a list of Workspace
type WorkspaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workspace `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Workspace{}, &WorkspaceList{})
}
