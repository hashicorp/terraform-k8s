package workspace

import (
	"bytes"
	"context"
	"fmt"

	"github.com/hashicorp/terraform-k8s/operator/pkg/apis/app/v1alpha1"
	"github.com/hashicorp/terraform/states/statefile"
)

// GetStateVersionDownloadURL retrieves download URL for state file
func (t *TerraformCloudClient) GetStateVersionDownloadURL(workspaceID string, runID string) (string, error) {

	stateVersion, err := t.Client.StateVersions.Current(context.TODO(), workspaceID)
	if err != nil {
		return "", fmt.Errorf("could not get current state version, WorkspaceID, %s, Error, %v", workspaceID, err)
	}
	if stateVersion.Run.ID != runID {
		return "", fmt.Errorf("current state does not match runID, StateVersionRunID, %s, RunID, %s", stateVersion.Run.ID, runID)
	}

	return stateVersion.DownloadURL, nil
}

// GetOutputsFromState gets list of outputs from state file
func (t *TerraformCloudClient) GetOutputsFromState(stateDownloadURL string) ([]*v1alpha1.OutputStatus, error) {
	if stateDownloadURL == "" {
		return nil, fmt.Errorf("could not download blank state")
	}
	data, err := t.Client.StateVersions.Download(context.TODO(), stateDownloadURL)
	if err != nil {
		return nil, fmt.Errorf("could not download state, Error, %v", err)
	}
	reader := bytes.NewReader(data)
	file, err := statefile.Read(reader)
	outputValues := file.State.Modules[""].OutputValues
	outputs := []*v1alpha1.OutputStatus{}
	for key, value := range outputValues {
		if !value.Sensitive {
			outputs = append(outputs, &v1alpha1.OutputStatus{Key: key, Value: value.Value.AsString()})
		}
	}
	return outputs, nil
}
