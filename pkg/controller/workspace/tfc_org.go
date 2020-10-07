package workspace

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	tfc "github.com/hashicorp/go-tfe"
	appv1alpha1 "github.com/hashicorp/terraform-k8s/pkg/apis/app/v1alpha1"
	"github.com/hashicorp/terraform-k8s/version"
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
	if err != nil {
		return nil, fmt.Errorf("Not a valid Terraform Cloud or Enterprise URL, %v", err)
	}
	if u.Scheme == "" {
		return nil, fmt.Errorf("Invalid Terraform Cloud or Enterprise URL. Please specify a scheme (http:// or https://)")
	}
	host := u.Host
	if host == "" {
		return nil, fmt.Errorf("Terraform Cloud or Enterprise URL hostname is ''. Invalid hostname")
	}

	if len(tfConfig.Credentials[host]) == 0 {
		return nil, fmt.Errorf("Define token for %s", host)
	}

	httpClient := tfc.DefaultConfig().HTTPClient
	transport := httpClient.Transport.(*http.Transport)
	if transport.TLSClientConfig == nil {
		transport.TLSClientConfig = &tls.Config{}
	}
	skipTLS := os.Getenv("TF_INSECURE")
	if skipTLS != "" && strings.ToLower(skipTLS) != "false" {
		transport.TLSClientConfig.InsecureSkipVerify = true
	}

	ua := fmt.Sprintf("terraform-k8s/%s", version.Version)
	return &tfc.Config{
		Address: address,
		Token:   fmt.Sprintf("%v", tfConfig.Credentials[host]["token"]),
		Headers: http.Header{"User-Agent": []string{ua}},
		HTTPClient: httpClient,
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
		return err
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
func (t *TerraformCloudClient) CheckWorkspace(workspace string, instance *appv1alpha1.Workspace) (*tfc.Workspace, error) {
	ws, err := t.Client.Workspaces.Read(context.TODO(), t.Organization, workspace)
	if err != nil && err == tfc.ErrResourceNotFound {
		id, err := t.CreateWorkspace(workspace, instance)
		if err != nil {
			return nil, err
		} else {
			ws = &tfc.Workspace{ID: id}
		}
	} else if err != nil {
		return nil, err
	}

	if instance.Spec.SSHKeyID != "" {
		_, err = t.AssignWorkspaceSSHKey(ws.ID, instance.Spec.SSHKeyID)
		if err != nil {
			return nil, fmt.Errorf("Error while assigning ssh key to workspace: %s", err)
		}
	} else if ws.SSHKey != nil {
		_, err = t.UnassignWorkspaceSSHKey(ws.ID)
		if err != nil {
			return nil, fmt.Errorf("Error while unassigning ssh key to workspace: %s", err)
		}
	}

	return ws, err
}

func getTerraformVersion() *string {
	tfVersion := os.Getenv("TF_VERSION")
	if tfVersion == "" {
		return &version.TerraformVersion
	}
	return &tfVersion
}

// CreateWorkspace creates a Terraform Cloud Workspace that auto-applies
func (t *TerraformCloudClient) CreateWorkspace(workspace string, instance *appv1alpha1.Workspace) (string, error) {
	options := tfc.WorkspaceCreateOptions{
		AutoApply:        &AutoApply,
		Name:             &workspace,
		TerraformVersion: getTerraformVersion(),
	}

	if instance.Spec.VCS != nil {
		options.VCSRepo = &tfc.VCSRepoOptions{
			Branch:            &instance.Spec.VCS.Branch,
			Identifier:        &instance.Spec.VCS.RepoIdentifier,
			IngressSubmodules: &instance.Spec.VCS.IngressSubmodules,
			OAuthTokenID:      &instance.Spec.VCS.TokenID,
		}
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
