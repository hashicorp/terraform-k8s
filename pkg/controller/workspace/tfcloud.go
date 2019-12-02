package workspace

import (
	"context"
	"fmt"

	tfc "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform/command/cliconfig"
)

var (
	ErrResourceNotFound = tfc.ErrResourceNotFound
)

type TerraformCloudClient struct {
	Client *tfc.Client
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

func (t *TerraformCloudClient) CreateWorkspace(organization string, name string) error {
	autoApply := true
	options := tfc.WorkspaceCreateOptions{
		AutoApply: &autoApply,
		Name:      &name,
	}
	_, err := t.Client.Workspaces.Create(context.TODO(), organization, options)
	if err != nil {
		return err
	}
	return nil
}
