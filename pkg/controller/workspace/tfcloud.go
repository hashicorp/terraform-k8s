package workspace

import (
	"context"
	"fmt"

	tfc "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform/command/cliconfig"
)

const (
	PageSize = 500
)

var (
	ErrResourceNotFound = tfc.ErrResourceNotFound
)

type TerraformCloudClient struct {
	Client       *tfc.Client
	Organization string
	Workspace    string
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
func (t *TerraformCloudClient) CheckWorkspace() error {
	_, err := t.Client.Workspaces.Read(context.TODO(), t.Organization, t.Workspace)
	return err
}

func (t *TerraformCloudClient) listVariables() (*tfc.VariableList, error) {
	options := tfc.VariableListOptions{
		ListOptions:  tfc.ListOptions{PageSize: PageSize},
		Organization: &t.Organization,
		Workspace:    &t.Workspace,
	}
	return t.Client.Variables.List(context.TODO(), options)
}

// CreateWorkspace creates a Terraform Cloud Workspace that auto-applies
func (t *TerraformCloudClient) CreateWorkspace() error {
	autoApply := true
	options := tfc.WorkspaceCreateOptions{
		AutoApply: &autoApply,
		Name:      &t.Workspace,
	}
	_, err := t.Client.Workspaces.Create(context.TODO(), t.Organization, options)
	if err != nil {
		return err
	}
	return nil
}

// DeleteWorkspace removes the workspace from Terraform Cloud
func (t *TerraformCloudClient) DeleteWorkspace() error {
	err := t.Client.Workspaces.Delete(context.TODO(), t.Organization, t.Workspace)
	if err != nil {
		return err
	}
	return nil
}
