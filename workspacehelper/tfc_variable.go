package workspacehelper

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	tfc "github.com/hashicorp/go-tfe"
	"github.com/snyk/terraform-k8s/api/v1alpha1"
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

// MapToTFCVariable changes the controller spec to a TFC Variable
func MapToTFCVariable(specVariables []*v1alpha1.Variable) []*tfc.Variable {
	tfcVariables := []*tfc.Variable{}
	for _, variable := range specVariables {
		tfcVariables = append(tfcVariables, &tfc.Variable{
			Key:       variable.Key,
			Value:     strings.TrimSuffix(variable.Value, "\n"),
			Sensitive: variable.Sensitive,
			Category:  setVariableType(variable.EnvironmentVariable),
			HCL:       variable.HCL,
		})
	}
	return tfcVariables
}

// CheckSecretsMountPath ensure the secrets mount path actually exists
func (t *TerraformCloudClient) CheckSecretsMountPath() error {
	if _, err := os.Stat(t.SecretsMountPath); os.IsNotExist(err) {
		return fmt.Errorf("Secrets Mount Path is invalid: %s", t.SecretsMountPath)
	}
	return nil
}

func (t *TerraformCloudClient) deleteVariablesFromTFC(specTFCVariables []*tfc.Variable, workspaceVariables []*tfc.Variable) error {
	for _, v := range workspaceVariables {
		index := find(specTFCVariables, v.Key)
		if index < 0 {
			err := t.DeleteVariable(v)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *TerraformCloudClient) createVariablesOnTFC(workspace *tfc.Workspace, specTFCVariables []*tfc.Variable, workspaceVariables []*tfc.Variable) (bool, error) {
	updated := false
	for _, v := range specTFCVariables {
		index := find(workspaceVariables, v.Key)
		if index < 0 {
			err := t.CreateTerraformVariable(workspace, v)
			if err != nil {
				return false, err
			}
			updated = true
			continue
		}
	}
	return updated, nil
}

func checkIfVariableChanged(specVariable *tfc.Variable, workspaceVariable *tfc.Variable) bool {
	if specVariable.Value != workspaceVariable.Value {
		return true
	}
	if specVariable.HCL != workspaceVariable.HCL {
		return true
	}
	if !specVariable.Sensitive && workspaceVariable.Sensitive {
		return true
	}
	return false
}

func getNonSensitiveVariablesToUpdate(specTFCVariables []*tfc.Variable, workspaceVariables []*tfc.Variable) []*tfc.Variable {
	variablesToUpdate := []*tfc.Variable{}
	for _, v := range specTFCVariables {
		index := find(workspaceVariables, v.Key)
		if index < 0 || workspaceVariables[index].Sensitive {
			continue
		}
		if checkIfVariableChanged(v, workspaceVariables[index]) {
			v.ID = workspaceVariables[index].ID
			v.Workspace = workspaceVariables[index].Workspace
			variablesToUpdate = append(variablesToUpdate, v)
		}
	}
	return variablesToUpdate
}

func getSensitiveVariablesToUpdate(specTFCVariables []*tfc.Variable, workspaceVariables []*tfc.Variable, secretsMountPath string) ([]*tfc.Variable, error) {
	variablesToUpdate := []*tfc.Variable{}
	for _, v := range specTFCVariables {
		index := find(workspaceVariables, v.Key)
		if index < 0 {
			continue
		}
		if workspaceVariables[index].Sensitive {
			if err := checkAndRetrieveIfSensitive(v, secretsMountPath); err != nil {
				return nil, err
			}
			v.ID = workspaceVariables[index].ID
			v.Workspace = workspaceVariables[index].Workspace
			v.Sensitive = true
			variablesToUpdate = append(variablesToUpdate, v)
		}
	}
	return variablesToUpdate, nil
}

func generateUpdateVariableList(specTFCVariables []*tfc.Variable, workspaceVariables []*tfc.Variable, secretsMountPath string) ([]*tfc.Variable, error) {
	updateList := []*tfc.Variable{}

	nonSensitiveVariablesToUpdate := getNonSensitiveVariablesToUpdate(specTFCVariables, workspaceVariables)
	if len(nonSensitiveVariablesToUpdate) == 0 {
		return updateList, nil
	}

	sensitiveVariablesToUpdate, err := getSensitiveVariablesToUpdate(specTFCVariables, workspaceVariables, secretsMountPath)
	if err != nil {
		return nonSensitiveVariablesToUpdate, err
	}

	updateList = append(nonSensitiveVariablesToUpdate, sensitiveVariablesToUpdate...)

	return updateList, nil
}

// CheckVariables creates, updates, or deletes variables as needed
func (t *TerraformCloudClient) CheckVariables(workspace string, specTFCVariables []*tfc.Variable) (bool, error) {
	tfcWorkspace, err := t.Client.Workspaces.Read(context.TODO(), t.Organization, workspace)
	if err != nil {
		return false, err
	}
	workspaceVariables, err := t.listVariables(tfcWorkspace.ID)
	if err != nil {
		return false, err
	}
	if err := t.deleteVariablesFromTFC(specTFCVariables, workspaceVariables); err != nil {
		return false, err
	}

	createdVariables, err := t.createVariablesOnTFC(tfcWorkspace, specTFCVariables, workspaceVariables)
	if err != nil {
		return false, err
	}

	variablesToUpdate, err := generateUpdateVariableList(specTFCVariables, workspaceVariables, t.SecretsMountPath)
	if err != nil || len(variablesToUpdate) == 0 {
		return false, err
	}

	if err = t.UpdateTerraformVariables(variablesToUpdate); err != nil {
		return false, err
	}

	return createdVariables || len(variablesToUpdate) > 0, nil
}

func find(tfcVariables []*tfc.Variable, key string) int {
	for index, variable := range tfcVariables {
		if variable.Key == key {
			return index
		}
	}
	return -1
}

func (t *TerraformCloudClient) listVariables(workspaceID string) ([]*tfc.Variable, error) {
	options := tfc.VariableListOptions{
		ListOptions: tfc.ListOptions{PageSize: PageSize},
	}
	variables, err := t.Client.Variables.List(context.TODO(), workspaceID, options)
	if err != nil {
		return nil, err
	}
	return variables.Items, nil
}

// DeleteVariable removes the variable by ID from Terraform Cloud
func (t *TerraformCloudClient) DeleteVariable(variable *tfc.Variable) error {
	err := t.Client.Variables.Delete(context.TODO(), variable.Workspace.ID, variable.ID)
	if err != nil {
		return err
	}
	return nil
}

// UpdateTerraformVariables updates a list of variable
func (t *TerraformCloudClient) UpdateTerraformVariables(variables []*tfc.Variable) error {
	if len(variables) == 0 {
		return nil
	}
	for _, v := range variables {
		options := tfc.VariableUpdateOptions{
			Key:       &v.Key,
			Value:     &v.Value,
			HCL:       &v.HCL,
			Sensitive: &v.Sensitive,
		}
		_, err := t.Client.Variables.Update(context.TODO(), v.Workspace.ID, v.ID, options)
		if err != nil {
			return err
		}
	}
	return nil
}

func checkAndRetrieveIfSensitive(variable *tfc.Variable, secretsMountPath string) error {
	// Try to read variables with empty value from file. If the value isn't empty,
	// it was already read fromValue.SecretKeyRef.
	if variable.Sensitive && variable.Value == "" {
		filePath := fmt.Sprintf("%s/%s", secretsMountPath, variable.Key)

		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("could not get secret, %s", err)
		}
		secret := string(data)
		variable.Value = secret
	}
	return nil
}

// CreateTerraformVariable creates a Terraform variable based on key and value
func (t *TerraformCloudClient) CreateTerraformVariable(workspace *tfc.Workspace, variable *tfc.Variable) error {
	if err := checkAndRetrieveIfSensitive(variable, t.SecretsMountPath); err != nil {
		return err
	}
	options := tfc.VariableCreateOptions{
		Key:       &variable.Key,
		Value:     &variable.Value,
		Category:  &variable.Category,
		Sensitive: &variable.Sensitive,
		HCL:       &variable.HCL,
	}
	_, err := t.Client.Variables.Create(context.TODO(), workspace.ID, options)
	if err != nil {
		return err
	}
	return nil
}
