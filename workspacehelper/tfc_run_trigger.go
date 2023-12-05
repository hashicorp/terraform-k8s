package workspacehelper

import (
	tfc "github.com/hashicorp/go-tfe"

	"context"
	"github.com/snyk/terraform-k8s/api/v1alpha1"
)

// MapToTFCRunTrigger changes controller spec to a TFC RunTrigger
func MapToTFCRunTrigger(workspace string, specRunTriggers []*v1alpha1.RunTrigger) []*tfc.RunTrigger {
	tfcRunTriggers := []*tfc.RunTrigger{}
	for _, runTrigger := range specRunTriggers {
		tfcRunTriggers = append(tfcRunTriggers, &tfc.RunTrigger{
			SourceableName: runTrigger.SourceableName,
			WorkspaceName:  workspace,
		})
	}
	return tfcRunTriggers
}

// Deletes run triggers in TFC that were not defined in controller spec
func (t *TerraformCloudClient) deleteRunTriggersFromTFC(specTFCRunTriggers []*tfc.RunTrigger, workspaceRunTriggers []*tfc.RunTrigger) error {
	for _, rt := range workspaceRunTriggers {
		index := findRT(specTFCRunTriggers, rt.SourceableName)
		if index < 0 {
			err := t.DeleteRunTrigger(rt)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Creates run triggers in TFC that were defined in controller spec but not created
func (t *TerraformCloudClient) createRunTriggersOnTFC(workspace *tfc.Workspace, specTFCRunTriggers []*tfc.RunTrigger, workspaceRunTriggers []*tfc.RunTrigger) (bool, error) {
	updated := false
	for _, rt := range specTFCRunTriggers {
		index := findRT(workspaceRunTriggers, rt.SourceableName)
		if index < 0 {
			err := t.CreateTerraformRunTrigger(workspace, rt)
			if err != nil {
				return false, err
			}
			updated = true
			continue
		}
	}
	return updated, nil
}

// CheckRunTriggers deletes and update TFC run triggers as needed
func (t *TerraformCloudClient) CheckRunTriggers(workspace string, specRunTriggers []*v1alpha1.RunTrigger) (bool, error) {
	tfcWorkspace, err := t.Client.Workspaces.Read(context.TODO(), t.Organization, workspace)
	if err != nil {
		return false, err
	}

	specTFCRunTriggers := MapToTFCRunTrigger(workspace, specRunTriggers)
	workspaceRunTriggers, err := t.listRunTriggers(tfcWorkspace.ID)
	if err != nil {
		return false, err
	}
	if err := t.deleteRunTriggersFromTFC(specTFCRunTriggers, workspaceRunTriggers); err != nil {
		return false, err
	}
	createdRunTriggers, err := t.createRunTriggersOnTFC(tfcWorkspace, specTFCRunTriggers, workspaceRunTriggers)
	if err != nil {
		return false, err
	}
	return createdRunTriggers, err
}

func (t *TerraformCloudClient) listRunTriggers(workspaceID string) ([]*tfc.RunTrigger, error) {
	options := tfc.RunTriggerListOptions{
		RunTriggerType: tfc.String("inbound"),
	}
	runTriggers, err := t.Client.RunTriggers.List(context.TODO(), workspaceID, options)
	if err != nil {
		return nil, err
	}
	return runTriggers.Items, nil
}

func findRT(tfcRunTriggers []*tfc.RunTrigger, sourceableName string) int {
	for index, runTrigger := range tfcRunTriggers {
		if runTrigger.SourceableName == sourceableName {
			return index
		}
	}
	return -1
}

func (t *TerraformCloudClient) DeleteRunTrigger(runTrigger *tfc.RunTrigger) error {
	err := t.Client.RunTriggers.Delete(context.TODO(), runTrigger.ID)
	if err != nil {
		return err
	}
	return nil
}

func (t *TerraformCloudClient) CreateTerraformRunTrigger(workspace *tfc.Workspace, runTrigger *tfc.RunTrigger) error {
	tfcSourceWorkspace, err := t.Client.Workspaces.Read(context.TODO(), t.Organization, runTrigger.SourceableName)
	if err != nil {
		return err
	}
	options := tfc.RunTriggerCreateOptions{
		Sourceable: tfcSourceWorkspace,
	}
	t.Client.RunTriggers.Create(context.TODO(), workspace.ID, options)
	if err != nil {
		return err
	}
	return nil
}
