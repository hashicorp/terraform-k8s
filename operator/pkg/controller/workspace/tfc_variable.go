package workspace

import (
	"context"
	"fmt"
	"os"

	tfc "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-k8s/operator/pkg/apis/app/v1alpha1"
)

const (
	// PageSize is page size for TFC API
	PageSize = 500
)

func setVariableType(isEnvironmentVariable bool) tfc.CategoryType {
	if isEnvironmentVariable {
		return tfc.CategoryEnv
	}
	return tfc.CategoryTerraform
}

func setHCL(isHCL bool) bool {
	if isHCL {
		return true
	}
	return false
}

// CheckSecretsMountPath ensure the secrets mount path actually exists
func (t *TerraformCloudClient) CheckSecretsMountPath() error {
	if _, err := os.Stat(t.SecretsMountPath); os.IsNotExist(err) {
		return fmt.Errorf("Secrets Mount Path is invalid: %s", t.SecretsMountPath)
	}
	return nil
}

func (t *TerraformCloudClient) deleteVariablesFromTFC(specTFCVariables map[string]*Variable, workspaceVariables map[string]*Variable, la v1alpha1.LastAppliedVariableValues) error {
	for k, v := range workspaceVariables {
		if _, ok := specTFCVariables[k]; !ok {
			err := t.DeleteVariable(v)
			if err != nil {
				return err
			}
			delete(la, k)
		}
	}
	return nil
}

func (t *TerraformCloudClient) updateVariablesOnTFC(workspace *tfc.Workspace, specTFCVariables map[string]*Variable, workspaceVariables map[string]*Variable, la v1alpha1.LastAppliedVariableValues) (bool, error) {
	updated := false
	for k, v := range specTFCVariables {
		err := v.CheckAndRetrieveIfSensitive(t)
		if err != nil {
			return false, err
		}

		// Create Variable
		if _, ok := la[k]; !ok {
			err := t.CreateTerraformVariable(workspace, v)
			if err != nil {
				return false, err
			}
			updated = v.SetStatus(la)
			continue
		}

		// Update Variable
		if v.Changed(la) {
			err := t.UpdateTerraformVariable(workspaceVariables[k], v.Value)
			if err != nil {
				return false, err
			}
			updated = v.SetStatus(la)
			continue
		}

		// Update if not updated yet and AlwaysUpdate is true
		if v.AlwaysUpdate {
			err := t.UpdateTerraformVariable(workspaceVariables[k], v.Value)
			if err != nil {
				return false, err
			}
			updated = v.SetStatus(la)
		}
	}

	return updated, nil
}

// CheckVariables creates, updates, or deletes variables as needed
func (t *TerraformCloudClient) CheckVariables(workspace string, specTFCVariables map[string]*Variable, la v1alpha1.LastAppliedVariableValues) (bool, error) {
	tfcWorkspace, err := t.Client.Workspaces.Read(context.TODO(), t.Organization, workspace)
	if err != nil {
		return false, err
	}
	workspaceVariables, err := t.mapVariables(workspace)
	if err != nil {
		return false, err
	}
	if err := t.deleteVariablesFromTFC(specTFCVariables, workspaceVariables, la); err != nil {
		return false, err
	}

	return t.updateVariablesOnTFC(tfcWorkspace, specTFCVariables, workspaceVariables, la)
}

func (t *TerraformCloudClient) mapVariables(workspace string) (map[string]*Variable, error) {
	options := tfc.VariableListOptions{
		ListOptions:  tfc.ListOptions{PageSize: PageSize},
		Organization: &t.Organization,
		Workspace:    &workspace,
	}
	tfcVariables, err := t.Client.Variables.List(context.TODO(), options)
	if err != nil {
		return nil, err
	}
	variables := make(map[string]*Variable)
	for _, item := range tfcVariables.Items {
		variables[item.Key] = &Variable{item, "", false}
	}
	return variables, nil
}

// DeleteVariable removes the variable by ID from Terraform Cloud
func (t *TerraformCloudClient) DeleteVariable(variable *Variable) error {
	err := t.Client.Variables.Delete(context.TODO(), variable.ID)
	if err != nil {
		return err
	}
	return nil
}

// UpdateTerraformVariable updates a variable
func (t *TerraformCloudClient) UpdateTerraformVariable(variable *Variable, newValue string) error {
	options := tfc.VariableUpdateOptions{
		Key:       &variable.Key,
		Value:     &newValue,
		Sensitive: &variable.Sensitive,
	}
	_, err := t.Client.Variables.Update(context.TODO(), variable.ID, options)
	if err != nil {
		return err
	}
	return nil
}

// CreateTerraformVariable creates a Terraform variable based on key and value
func (t *TerraformCloudClient) CreateTerraformVariable(workspace *tfc.Workspace, variable *Variable) error {
	options := tfc.VariableCreateOptions{
		Key:       &variable.Key,
		Value:     &variable.Value,
		Category:  &variable.Category,
		Sensitive: &variable.Sensitive,
		HCL:       &variable.HCL,
		Workspace: workspace,
	}
	_, err := t.Client.Variables.Create(context.TODO(), options)
	if err != nil {
		return err
	}
	return nil
}
