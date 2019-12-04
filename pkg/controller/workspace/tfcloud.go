package workspace

import (
	"context"
	"fmt"

	tfc "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-k8s/pkg/apis/app/v1alpha1"
	"github.com/hashicorp/terraform/command/cliconfig"
)

const (
	PageSize = 500
)

var (
	ErrResourceNotFound = tfc.ErrResourceNotFound
	TerraformVariable   = tfc.CategoryTerraform
	EnvironmentVariable = tfc.CategoryEnv
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

func (t *TerraformCloudClient) listVariables() (*tfc.VariableList, error) {
	options := tfc.VariableListOptions{
		ListOptions:  tfc.ListOptions{PageSize: PageSize},
		Organization: &t.Organization,
		Workspace:    &t.Workspace,
	}
	return t.Client.Variables.List(context.TODO(), options)
}

// DeleteAllVariables removes all variables from the workspace for re-creation
func (t *TerraformCloudClient) DeleteAllVariables() error {
	variables, err := t.listVariables()
	if err != nil {
		return err
	}
	for _, variable := range variables.Items {
		err := t.Client.Variables.Delete(context.TODO(), variable.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateTerraformVariables creates Terraform variables for Terraform Cloud
func (t *TerraformCloudClient) CreateTerraformVariables(variables []*v1alpha1.Variable) error {
	for _, variable := range variables {
		options := tfc.VariableCreateOptions{
			Key:       &variable.Key,
			Value:     &variable.Value,
			Category:  &TerraformVariable,
			Sensitive: &variable.Sensitive,
			Workspace: &tfc.Workspace{Name: t.Workspace},
		}
		_, err := t.Client.Variables.Create(context.TODO(), options)
		if err != nil {
			return err
		}
	}
	return nil
}
