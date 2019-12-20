package workspace

import (
	"context"
	"fmt"

	tfc "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-k8s/operator/pkg/apis/app/v1alpha1"
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

// UpsertWorkspace looks for a workspace
func (r *ReconcileWorkspace) UpsertWorkspace(found *v1alpha1.Workspace, workspace string) error {
	var workspaceID string
	ws, err := r.tfclient.ReadWorkspace(workspace)
	if err != nil && err == tfc.ErrResourceNotFound {
		r.reqLogger.Info("Creating new workspace", "Name", workspace)
		workspaceID, err = r.tfclient.CreateWorkspace(workspace)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		workspaceID = ws.ID
	}

	if found.Status.WorkspaceID != workspaceID {
		r.reqLogger.Info("Updating workspace ID", "WorkspaceID", workspaceID)
		found.Status.WorkspaceID = workspaceID
		if err := r.client.Update(context.TODO(), found); err != nil {
			r.reqLogger.Error(err, "Failed to update workspace", "Namespace", found.Namespace, "Name", found.Name)
			return err
		}
		return nil
	}
	return nil
}

func (t *TerraformCloudClient) ReadWorkspace(workspace string) (*tfc.Workspace, error) {
	return t.Client.Workspaces.Read(context.TODO(), t.Organization, workspace)
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
