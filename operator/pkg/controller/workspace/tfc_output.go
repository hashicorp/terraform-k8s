package workspace

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-k8s/operator/pkg/apis/app/v1alpha1"
)

// GetStateVersionDownloadURL retrieves download URL for state file
func (t *TerraformCloudClient) GetStateVersionDownloadURL(workspace *v1alpha1.Workspace) error {
	workspaceID := workspace.Status.WorkspaceID
	runID := workspace.Status.RunID

	stateVersion, err := t.Client.StateVersions.Current(context.TODO(), workspaceID)
	if err != nil {
		return fmt.Errorf("could not get current state version, WorkspaceID, %s, Error, %v", workspaceID, err)
	}
	if stateVersion.Run.ID != runID {
		return fmt.Errorf("current state does not match runID, StateVersionRunID, %s, RunID, %s", stateVersion.Run.ID, runID)
	}

	workspace.Status.StateDownloadURL = stateVersion.DownloadURL

	return nil
}

// GetOutputsFromState gets list of outputs from state file
func (t *TerraformCloudClient) GetOutputsFromState(workspace *v1alpha1.Workspace) error {
	if workspace.Status.StateDownloadURL == "" {
		return fmt.Errorf("could not download blank state")
	}
	_, err := t.Client.StateVersions.Download(context.TODO(), workspace.Status.StateDownloadURL)
	if err != nil {
		return fmt.Errorf("could not download state, Error, %v", err)
	}
	return nil
}
