package workspace

import (
	"context"
	"fmt"

	tfc "github.com/hashicorp/go-tfe"
	appv1alpha1 "github.com/hashicorp/terraform-k8s/operator/pkg/apis/app/v1alpha1"
	"github.com/hashicorp/terraform/command/cliconfig"
)

var (
	// AutoApply run to workspace
	AutoApply = true
)

// TerraformCloudClient has a TFC Client and organization
type TerraformCloudClient struct {
	Client           *tfc.Client
	Organization     string
	SecretsMountPath string
}

// GetClient creates the configuration for Terraform Cloud
func (t *TerraformCloudClient) GetClient() error {
	tfConfig, diag := cliconfig.LoadConfig()
	if diag.Err() != nil {
		return diag.Err()
	}

	config := &tfc.Config{
		Token: fmt.Sprintf("%v", tfConfig.Credentials["app.terraform.io"]["token"]),
	}

	client, err := tfc.NewClient(config)
	if err != nil {
		return diag.Err()
	}
	t.Client = client
	return nil
}

// CheckOrganization looks for an organization
func (t *TerraformCloudClient) CheckOrganization() error {
	_, err := t.Client.Organizations.Read(context.TODO(), t.Organization)
	return err
}

// CheckWorkspace looks for a workspace
func (t *TerraformCloudClient) CheckWorkspace(workspace string, instance *appv1alpha1.Workspace) (string, error) {
	ws, err := t.Client.Workspaces.Read(context.TODO(), t.Organization, workspace)
	if err != nil && err == tfc.ErrResourceNotFound {
		id, err := t.CreateWorkspace(workspace)
		if err != nil {
			return "", err
		}
		return id, nil
	} else if err != nil {
		return "", err
	}

	if instance.Spec.SSHKeyID != "" {
		t.AssignWorkspaceSSHKey(ws.ID, instance.Spec.SSHKeyID)
	} else if ws.SSHKey != nil {
		t.UnassignWorkspaceSSHKey(ws.ID)
	}

	return ws.ID, err
}

// CreateWorkspace creates a Terraform Cloud Workspace that auto-applies
func (t *TerraformCloudClient) CreateWorkspace(workspace string) (string, error) {
	options := tfc.WorkspaceCreateOptions{
		AutoApply: &AutoApply,
		Name:      &workspace,
	}
	ws, err := t.Client.Workspaces.Create(context.TODO(), t.Organization, options)
	if err != nil {
		return "", err
	}

	return ws.ID, nil
}

// Assign SSHKey to Terraform Cloud Workspace
func (t *TerraformCloudClient) AssignWorkspaceSSHKey(workspaceID string, SSHKeyID string) (string, error) {
	sshOptions := tfc.WorkspaceAssignSSHKeyOptions{
		SSHKeyID: &SSHKeyID,
	}
	ws, err := t.Client.Workspaces.AssignSSHKey(context.TODO(), workspaceID, sshOptions)
	if err != nil {
		return "", err
	}

	return ws.ID, nil
}

// Unassign SSHKey from Terraform Cloud Workspace
func (t *TerraformCloudClient) UnassignWorkspaceSSHKey(workspaceID string) (string, error) {
	ws, err := t.Client.Workspaces.UnassignSSHKey(context.TODO(), workspaceID)
	if err != nil {
		return "", err
	}

	return ws.ID, nil
}

// CheckWorkspacebyID checks a workspace by ID
func (t *TerraformCloudClient) CheckWorkspacebyID(workspaceID string) error {
	_, err := t.Client.Workspaces.ReadByID(context.TODO(), workspaceID)
	return err
}

// DeleteWorkspace removes the workspace from Terraform Cloud
func (t *TerraformCloudClient) DeleteWorkspace(workspaceID string) error {
	err := t.Client.Workspaces.DeleteByID(context.TODO(), workspaceID)
	if err != nil {
		return err
	}
	return nil
}
