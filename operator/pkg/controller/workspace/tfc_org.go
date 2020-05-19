package workspace

import (
	"context"
	"fmt"
	"net/url"
	"os"

	tfc "github.com/hashicorp/go-tfe"
	appv1alpha1 "github.com/hashicorp/terraform-k8s/operator/pkg/apis/app/v1alpha1"
	"github.com/hashicorp/terraform-k8s/operator/version"
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

func createTerraformConfig(address string, tfConfig *cliconfig.Config) (*tfc.Config, error) {
	if len(address) == 0 {
		address = tfc.DefaultAddress
	}
	u, err := url.Parse(address)
    if u.Scheme == "" {
    return nil, fmt.Errorf("Invalid Terraform Cloud or Enterprise URL. Please specify a scheme (http:// or https://)")
    }
	if err != nil {
		return nil, fmt.Errorf("Not a valid Terraform Cloud or Enterprise URL, %v", err)
	}
	host := u.Host
	if host == "" {
		return nil, fmt.Errorf("Terraform Cloud or Enterprise URL hostname is ''. Invalid hostname")
	}

	if len(tfConfig.Credentials[host]) == 0 {
		return nil, fmt.Errorf("Define token for %s", host)
	}

	return &tfc.Config{
		Address: address,
		Token:   fmt.Sprintf("%v", tfConfig.Credentials[host]["token"]),
	}, nil
}

// GetClient creates the configuration for Terraform Cloud
func (t *TerraformCloudClient) GetClient(tfEndpoint string) error {
	tfConfig, diag := cliconfig.LoadConfig()
	if diag.Err() != nil {
		return diag.Err()
	}

	config, err := createTerraformConfig(tfEndpoint, tfConfig)
	if err != nil {
		return err
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

func getTerraformVersion() *string {
	tfVersion := os.Getenv("TF_VERSION")
	if tfVersion == "" {
		return &version.TerraformVersion
	}
	return &tfVersion
}

// CreateWorkspace creates a Terraform Cloud Workspace that auto-applies
func (t *TerraformCloudClient) CreateWorkspace(workspace string) (string, error) {
	options := tfc.WorkspaceCreateOptions{
		AutoApply:        &AutoApply,
		Name:             &workspace,
		TerraformVersion: getTerraformVersion(),
	}
	ws, err := t.Client.Workspaces.Create(context.TODO(), t.Organization, options)
	if err != nil {
		return "", err
	}

	return ws.ID, nil
}

// GetSSHKeyByNameOrID Lookup provided Key ID by name or ID, return ID.
func (t *TerraformCloudClient) GetSSHKeyByNameOrID(SSHKeyID string) (string, error) {
	sshKeys, err := t.Client.SSHKeys.List(context.TODO(), t.Organization, tfc.SSHKeyListOptions{})
	if err != nil {
		return "", err
	}
	for _, elem := range sshKeys.Items {
		if elem.ID == SSHKeyID || elem.Name == SSHKeyID {
			return elem.ID, nil
		}
	}
	log.Error(tfc.ErrResourceNotFound, "No SSHKey found for "+SSHKeyID)
	return "", tfc.ErrResourceNotFound
}

// AssignWorkspaceSSHKey to Terraform Cloud Workspace
func (t *TerraformCloudClient) AssignWorkspaceSSHKey(workspaceID string, SSHKeyID string) (string, error) {

	sshKey, err := t.GetSSHKeyByNameOrID(SSHKeyID)
	if err != nil {
		return "", err
	}
	sshOptions := tfc.WorkspaceAssignSSHKeyOptions{
		SSHKeyID: &sshKey,
	}
	ws, err := t.Client.Workspaces.AssignSSHKey(context.TODO(), workspaceID, sshOptions)
	if err != nil {
		return "", err
	}

	return ws.ID, nil
}

// UnassignWorkspaceSSHKey from Terraform Cloud Workspace
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
