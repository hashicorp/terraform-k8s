package organization

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	tfc "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-k8s/pkg/apis/app/v1alpha1"
)

const (
	// PageSize is page size for TFC API
	PageSize = 500
)

var (
	// TerraformVariable is a variable
	TerraformVariable = tfc.CategoryTerraform
	// EnvironmentVariable is an environment variable
	EnvironmentVariable = tfc.CategoryEnv
	// Sensitive defaults to false
	Sensitive = false
)

func changeTypeToTFCVariable(specVariables []*v1alpha1.Variable) []*tfc.Variable {
	tfcVariables := []*tfc.Variable{}
	for _, variable := range specVariables {
		tfcVariables = append(tfcVariables, &tfc.Variable{
			Key:       variable.Key,
			Value:     variable.Value,
			Sensitive: variable.Sensitive,
		})
	}
	return tfcVariables
}

func (t *TerraformCloudClient) CheckSecretsMountPath() error {
	if _, err := os.Stat(t.SecretsMountPath); os.IsNotExist(err) {
		return fmt.Errorf("Secrets Mount Path is invalid: %s", t.SecretsMountPath)
	}
	return nil
}

// CheckVariables creates, updates, or deletes variables as needed
func (t *TerraformCloudClient) CheckVariables(workspace string, specVariables []*v1alpha1.Variable) (bool, error) {
	variablesUpdated := false
	specTFCVariables := changeTypeToTFCVariable(specVariables)
	tfcWorkspace, err := t.Client.Workspaces.Read(context.TODO(), t.Organization, workspace)
	if err != nil {
		return variablesUpdated, err
	}
	workspaceVariables, err := t.listVariables(workspace)
	if err != nil {
		return variablesUpdated, err
	}
	for _, v := range workspaceVariables {
		index := find(specTFCVariables, v.Key)
		if index < 0 {
			err := t.DeleteVariable(v)
			if err != nil {
				return variablesUpdated, err
			}
		}
	}
	for _, v := range specTFCVariables {
		index := find(workspaceVariables, v.Key)
		if index < 0 {
			err := t.CreateTerraformVariable(tfcWorkspace, v)
			if err != nil {
				return variablesUpdated, err
			}
			variablesUpdated = true
			continue
		}
		if v.Value != workspaceVariables[index].Value {
			err := t.UpdateTerraformVariable(workspaceVariables[index], v.Value)
			if err != nil {
				return true, err
			}
			variablesUpdated = true
		}
	}
	return variablesUpdated, nil
}

func find(tfcVariables []*tfc.Variable, key string) int {
	for index, variable := range tfcVariables {
		if variable.Key == key {
			return index
		}
	}
	return -1
}

func (t *TerraformCloudClient) listVariables(workspace string) ([]*tfc.Variable, error) {
	options := tfc.VariableListOptions{
		ListOptions:  tfc.ListOptions{PageSize: PageSize},
		Organization: &t.Organization,
		Workspace:    &workspace,
	}
	variables, err := t.Client.Variables.List(context.TODO(), options)
	if err != nil {
		return nil, err
	}
	return variables.Items, nil
}

// DeleteVariable removes the variable by ID from Terraform Cloud
func (t *TerraformCloudClient) DeleteVariable(variable *tfc.Variable) error {
	err := t.Client.Variables.Delete(context.TODO(), variable.ID)
	if err != nil {
		return err
	}
	return nil
}

// UpdateTerraformVariable updates a variable
func (t *TerraformCloudClient) UpdateTerraformVariable(variable *tfc.Variable, newValue string) error {
	options := tfc.VariableUpdateOptions{
		Key:   &variable.Key,
		Value: &newValue,
	}
	_, err := t.Client.Variables.Update(context.TODO(), variable.ID, options)
	if err != nil {
		return err
	}
	return nil
}

func (t *TerraformCloudClient) checkAndRetrieveIfSensitive(variable *tfc.Variable) error {
	if variable.Sensitive {
		filePath := fmt.Sprintf("%s/%s", t.SecretsMountPath, variable.Key)
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
	t.checkAndRetrieveIfSensitive(variable)
	options := tfc.VariableCreateOptions{
		Key:       &variable.Key,
		Value:     &variable.Value,
		Category:  &TerraformVariable,
		Sensitive: &variable.Sensitive,
		Workspace: workspace,
	}
	_, err := t.Client.Variables.Create(context.TODO(), options)
	if err != nil {
		return err
	}
	return nil
}
