package workspacehelper

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"

	tfc "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform/command/cliconfig"
	appv1alpha1 "github.com/snyk/terraform-k8s/api/v1alpha1"
	"github.com/snyk/terraform-k8s/version"
)

var (
	// AutoApply run to workspace
	AutoApply     = true
	AgentPageSize = 100
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
		Address:    address,
		Token:      fmt.Sprintf("%v", tfConfig.Credentials[host]["token"]),
		Headers:    http.Header{"User-Agent": []string{ua}},
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

func (t *TerraformCloudClient) SetTerraformVersion(workspace, terraformVersion string) error {
	wsUpdateOptions := tfc.WorkspaceUpdateOptions{
		TerraformVersion: &terraformVersion,
	}
	_, err := t.Client.Workspaces.Update(context.TODO(), t.Organization, workspace, wsUpdateOptions)
	if err != nil {
		return err
	}
	return nil
}

// getAgentPoolID uses AgentPoolName to lookup and return AgentPoolID
func getAgentPoolID(specTFCAgentPoolName string, agentPools []*tfc.AgentPool) (*tfc.AgentPool, error) {
	for _, agentPool := range agentPools {
		if specTFCAgentPoolName == agentPool.Name {
			return agentPool, nil
		}
	}
	return nil, fmt.Errorf("No valid agent pools exist with name %v", specTFCAgentPoolName)
}

func (t *TerraformCloudClient) listAgentPools() ([]*tfc.AgentPool, error) {
	options := tfc.AgentPoolListOptions{
		ListOptions: tfc.ListOptions{PageSize: AgentPageSize},
	}

	agentpools, err := t.Client.AgentPools.List(context.TODO(), t.Organization, options)
	if err != nil {
		return nil, fmt.Errorf("Problem fetching agent pools %s", err)
	}
	return agentpools.Items, nil
}

func (t *TerraformCloudClient) updateAgentPoolID(instance *appv1alpha1.Workspace, workspace *tfc.Workspace) error {
	var agentPoolID string

	// When agent pool name provided, look it up otherwise set to id in spec
	if instance.Spec.AgentPoolName != "" {
		agentPools, err := t.listAgentPools()
		agentPool, err := getAgentPoolID(instance.Spec.AgentPoolName, agentPools)
		if err != nil {
			return err
		}
		agentPoolID = agentPool.ID
	} else if instance.Spec.AgentPoolID != "" {
		agentPoolID = instance.Spec.AgentPoolID
	}

	if agentPoolID == workspace.AgentPoolID {
		return nil
	}

	updateOptions := tfc.WorkspaceUpdateOptions{
		AgentPoolID: &agentPoolID,
	}

	if agentPoolID != "" {
		updateOptions.ExecutionMode = tfc.String("agent")
	} else {
		updateOptions.ExecutionMode = tfc.String("")
	}

	_, err := t.Client.Workspaces.UpdateByID(context.TODO(), workspace.ID, updateOptions)
	if err != nil {
		return err
	}
	return nil
}

// CheckWorkspace looks for a remote tfc workspace
func (t *TerraformCloudClient) CheckWorkspace(workspace string, instance *appv1alpha1.Workspace) (*tfc.Workspace, error) {
	ws, err := t.Client.Workspaces.Read(context.TODO(), t.Organization, workspace)
	if err != nil && err == tfc.ErrResourceNotFound {
		id, wsErr := t.CreateWorkspace(workspace, instance)
		if wsErr != nil {
			return nil, wsErr
		} else {
			ws = &tfc.Workspace{ID: id}
			err = nil
		}
		ws = &tfc.Workspace{ID: id}
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

	if instance.Spec.TerraformVersion != ws.TerraformVersion {
		err = t.SetTerraformVersion(workspace, instance.Spec.TerraformVersion)
		if err != nil {
			return nil, err
		}
	}

	if instance.Spec.AgentPoolID != ws.AgentPoolID {
		err := t.updateAgentPoolID(instance, ws)
		if err != nil {
			return nil, fmt.Errorf("error while updating Agent Pool ID settings for workspace %q: %s", ws.Name, err)
		}
	}

	return ws, err
}

// CreateWorkspace creates a Terraform Cloud Workspace that auto-applies
func (t *TerraformCloudClient) CreateWorkspace(workspace string, instance *appv1alpha1.Workspace) (string, error) {
	var tfVersion string
	if instance.Spec.TerraformVersion == "" {
		tfVersion = "latest"
	} else {
		tfVersion = instance.Spec.TerraformVersion
	}

	options := tfc.WorkspaceCreateOptions{
		AutoApply:        &AutoApply,
		Name:             &workspace,
		TerraformVersion: &tfVersion,
	}

	if instance.Spec.VCS != nil {
		options.VCSRepo = &tfc.VCSRepoOptions{
			Branch:            &instance.Spec.VCS.Branch,
			Identifier:        &instance.Spec.VCS.RepoIdentifier,
			IngressSubmodules: &instance.Spec.VCS.IngressSubmodules,
			OAuthTokenID:      &instance.Spec.VCS.TokenID,
		}
	}

	if instance.Spec.AgentPoolID != "" {
		options.AgentPoolID = &instance.Spec.AgentPoolID
		options.ExecutionMode = tfc.String("agent")
	} else if instance.Spec.AgentPoolName != "" {
		agentPools, err := t.listAgentPools()
		if err != nil {
			return "", err
		}
		agentPool, err := getAgentPoolID(instance.Spec.AgentPoolName, agentPools)
		if err != nil {
			return "", err
		}
		options.AgentPoolID = &agentPool.ID
		options.ExecutionMode = tfc.String("agent")
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

func compareNotifications(wsNotification *tfc.NotificationConfiguration, specNotification *appv1alpha1.Notification) bool {
	if wsNotification.Name != specNotification.Name {
		return false
	}
	log.Info(fmt.Sprintf("%#v vs %#v", wsNotification, specNotification))

	if wsNotification.Token != specNotification.Token ||
		wsNotification.URL != specNotification.URL ||
		wsNotification.DestinationType != specNotification.Type ||
		wsNotification.Enabled != specNotification.Enabled {
		return false
	}
	if !reflect.DeepEqual(wsNotification.EmailAddresses, specNotification.Recipients) {
		return false
	}

	var wsUserIDs []string
	for _, user := range wsNotification.EmailUsers {
		wsUserIDs = append(wsUserIDs, user.ID)
	}

	if !reflect.DeepEqual(wsNotification.Triggers, specNotification.Triggers) {
		return false
	}

	return reflect.DeepEqual(wsUserIDs, specNotification.Users)
}

// Check Notification checks, and if necessary creates, the workspace's notifications
func (t *TerraformCloudClient) CheckNotifications(instance *appv1alpha1.Workspace) error {
	workspaceID := instance.Status.WorkspaceID

	notifications, err := t.Client.NotificationConfigurations.List(context.TODO(), workspaceID,
		tfc.NotificationConfigurationListOptions{})
	if err != nil {
		return err
	}

	if len(notifications.Items) == 0 && len(instance.Spec.Notifications) == 0 {
		return nil
	}

	var toRemove []*tfc.NotificationConfiguration

	// Find notifications not defined in spec
	for _, wsNotification := range notifications.Items {
		found := false
		for _, specNotification := range instance.Spec.Notifications {
			if compareNotifications(wsNotification, specNotification) {
				found = true
				break
			}
		}
		if !found {
			toRemove = append(toRemove, wsNotification)
		}
	}

	// Remove extra notifications
	for _, notification := range toRemove {
		err := t.Client.NotificationConfigurations.Delete(context.TODO(), notification.ID)
		if err != nil {
			return err
		}
	}

	// refresh notifications list
	notifications, err = t.Client.NotificationConfigurations.List(context.TODO(), workspaceID,
		tfc.NotificationConfigurationListOptions{})
	if err != nil {
		return err
	}
	var toAdd []*appv1alpha1.Notification
	// Find notifications missing from workspace
	for _, specNotification := range instance.Spec.Notifications {
		found := false
		for _, wsNotification := range notifications.Items {
			if wsNotification.Name == specNotification.Name {
				found = true
				break
			}
		}
		if !found {
			toAdd = append(toAdd, specNotification)
		}
	}

	// Add missing notifications
	for _, notification := range toAdd {
		createOpts := tfc.NotificationConfigurationCreateOptions{
			Name:            &notification.Name,
			DestinationType: &notification.Type,
			Enabled:         &notification.Enabled,
			Token:           &notification.Token,
			URL:             &notification.URL,
			EmailAddresses:  notification.Recipients,
			Triggers:        notification.Triggers,
		}
		for _, user := range notification.Users {
			createOpts.EmailUsers = append(createOpts.EmailUsers, &tfc.User{ID: user})
		}

		_, err := t.Client.NotificationConfigurations.Create(context.TODO(), workspaceID, createOpts)
		if err != nil {
			return err
		}
	}

	return nil
}
