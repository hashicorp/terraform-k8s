package workspace

import (
	"bytes"
	"context"
	"fmt"

	"github.com/hashicorp/terraform-k8s/operator/pkg/apis/app/v1alpha1"
	"github.com/hashicorp/terraform/states/statefile"
)

// GetStateVersionDownloadURL retrieves download URL for state file
func (t *TerraformCloudClient) GetStateVersionDownloadURL(workspaceID string) (string, error) {
	stateVersion, err := t.Client.StateVersions.Current(context.TODO(), workspaceID)
	if err != nil {
		return "", fmt.Errorf("could not get current state version, WorkspaceID, %s, Error, %v", workspaceID, err)
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

// CheckOutputs retrieves outputs for a run.
func (t *TerraformCloudClient) CheckOutputs(workspaceID string, runID string) ([]*v1alpha1.OutputStatus, error) {
	outputs := []*v1alpha1.OutputStatus{}
	if runID == "" {
		return outputs, nil
	}
	stateDownloadURL, err := t.GetStateVersionDownloadURL(workspaceID)
	if err != nil {
		return outputs, err
	}

	outputs, err = t.GetOutputsFromState(stateDownloadURL)
	if err != nil {
		return outputs, err
	}

	return outputs, nil
}
