package v1alpha1

import (
	tfc "github.com/hashicorp/go-tfe"
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

// VCS holds all the information needed to connect the workspace to a VCS repository
type VCS struct {
	// Token ID of the VCS Connection (OAuth Connection Token) to use
	// https://www.terraform.io/docs/cloud/vcs
	TokenID string `json:"token_id"`
	// A reference to your VCS repository in the format org/repo
	RepoIdentifier string `json:"repo_identifier"`
	// The repository branch to use
	// +optional
	Branch string `json:"branch"`
	// Whether submodules should be fetched when cloning the VCS repository (Defaults to false)
	IngressSubmodules bool `json:"ingress_submodules,omitempty"`
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
}

// Notification notification holds all the necessary information required to configure any workspace notification
type Notification struct {
	// Notification type. Can be one of email, generic, or slack
	Type tfc.NotificationDestinationType `json:"type"`
	// Control if the notification is enabled or not
	Enabled bool `json:"enabled"`
	// Name of the hook
	Name string `json:"name"`
	// URL of the hook
	// +optional
	URL string `json:"url"`
	// Token used to generate an HMAC on the verificatio0n request
	// +optional
	Token string `json:"token"`
	// When the web hook gets triggered. Acceptable values are run:created, run:planning, run:needs_attention, run:applying, run:completed, run:errored.
	// +optional
	Triggers []string `json:"triggers,omitempty"`
	// List of recipients' email addresses. Only applicable for TFE endpoints.
	// +optional
	Recipients []string `json:"recipients,omitempty"`
	// List of users to receive the notificaiton email.
	// +optional
	Users []string `json:"users,omitempty"`
}

// WorkspaceSpec defines the desired state of Workspace
// +k8s:openapi-gen=true
type WorkspaceSpec struct {
	// Terraform Cloud organization
	Organization string `json:"organization"`
	// Module source and version to use
	// +optional
	// +nullable
	Module *Module `json:"module"`
	// Details of the VCS repository we want to connect to the workspace
	// +optional
	// +nullable
	VCS *VCS `json:"vcs"`
	// Variables as inputs to module
	// +optional
	Variables []*Variable `json:"variables,omitempty"`
	// File path within operator pod to load workspace secrets
	SecretsMountPath string `json:"secretsMountPath"`
	// SSH Key ID. This key must already exist in the TF Cloud organization.  This can either be the user assigned name of the SSH Key, or the system assigned ID.
	// +optional
	SSHKeyID string `json:"sshKeyID,omitempty"`
	// Outputs denote outputs wanted
	// +optional
	Outputs []*OutputSpec `json:"outputs,omitempty"`
	// Terraform version used for this workspace. The default is `latest`.
	// +optional
	TerraformVersion string `json:"terraformVersion"`
	// Specifies the agent pool ID we wish to use.
	// +optional
	AgentPoolID string `json:"agentPoolID,omitempty"`
	// Notification configuration
	// +optional
	Notifications []*Notification `json:"notifications,omitempty"`
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
	// Configuration Version ID
	ConfigVersionID string `json:"configVersionID"`
	// Outputs from state file
	// +optional
	Outputs []*OutputStatus `json:"outputs,omitempty"`
}

// +kubebuilder:object:root=true

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

// +kubebuilder:object:root=true

// WorkspaceList contains a list of Workspace
type WorkspaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workspace `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Workspace{}, &WorkspaceList{})
}
