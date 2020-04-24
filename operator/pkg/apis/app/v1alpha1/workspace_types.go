package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// Module references a Terraform module
type Module struct {
	// Any remote module source (version control, registry)
	Source string `json:"source"`
	// Module version for registry modules
	// +optional
	Version string `json:"version"`
}

// OutputSpec specifies which values need to be output
type OutputSpec struct {
	// Output name
	// +optional
	Key string `json:"key"`
	// Attribute name in module
	// +optional
	ModuleOutputName string `json:"moduleOutputName"`
}

// OutputStatus outputs the values of Terraform output
type OutputStatus struct {
	// Attribute name in module
	// +optional
	Key string `json:"key"`
	// Value
	// +optional
	Value string `json:"value"`
}

// LastAppliedVariableValues stores the last applied value (or hash for sensitive) for each variable
type LastAppliedVariableValues map[string]string

// Variable denotes an input to the module
type Variable struct {
	// Variable name
	Key string `json:"key"`
	// Variable value
	// +optional
	Value string `json:"value"`
	// Source for the variable's value. Cannot be used if value is not empty.
	// +optional
	ValueFrom *corev1.EnvVarSource `json:"valueFrom,omitempty"`
	// String input should be an HCL-formatted variable
	// +optional
	HCL bool `json:"hcl"`
	// Variable is a secret and should be retrieved from file
	Sensitive bool `json:"sensitive"`
	// EnvironmentVariable denotes if this variable should be created as environment variable
	EnvironmentVariable bool `json:"environmentVariable"`
	// Always update in TFC is true
	// +optional
	AlwaysUpdate bool `json:"alwaysUpdate"`
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
	Variables []*Variable `json:"variables,omitempty"`
	// File path within operator pod to load workspace secrets
	SecretsMountPath string `json:"secretsMountPath"`
	// Outputs denote outputs wanted
	// +listType=set
	// +optional
	Outputs []*OutputSpec `json:"outputs,omitempty"`
}

// WorkspaceStatus defines the observed state of Workspace
// +k8s:openapi-gen=true
type WorkspaceStatus struct {
	// Run Status gets the run status
	RunStatus string `json:"runStatus"`
	// Workspace ID
	WorkspaceID string `json:"workspaceID"`
	// Run ID
	RunID string `json:"runID"`
	// Outputs from state file
	// +listType=set
	// +optional
	Outputs []*OutputStatus `json:"outputs,omitempty"`
	// Track Variable Value Changes
	// +optional
	LastAppliedVariableValues LastAppliedVariableValues `json:"lastAppliedVariableValues"`
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
