package workspace

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"text/template"
	"time"

	tfc "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-k8s/operator/pkg/apis/app/v1alpha1"
)

var (
	autoQueueRuns         = false
	speculative           = false
	isDestroy             = true
	basepath              = "/tmp"
	moduleDirectory       = fmt.Sprintf("%s/%s", basepath, "module")
	configurationFilePath = fmt.Sprintf("%s/%s", moduleDirectory, "main.tf")
	interval              = 30 * time.Second
)

// GetMD5Hash retrieves the md5sum for a byte array
func GetMD5Hash(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

// UploadConfigurationFile uploads the main.tf to a configuration version
func (t *TerraformCloudClient) UploadConfigurationFile(uploadURL string) error {
	if err := t.Client.ConfigurationVersions.Upload(context.TODO(), uploadURL, moduleDirectory); err != nil {
		return fmt.Errorf("error, %v, %v", err, moduleDirectory)
	}
	return nil
}

// CreateConfigurationVersion creates a configuration version for a workspace
func (t *TerraformCloudClient) CreateConfigurationVersion(workspaceID string) (*tfc.ConfigurationVersion, error) {
	options := tfc.ConfigurationVersionCreateOptions{
		AutoQueueRuns: &autoQueueRuns,
		Speculative:   &speculative,
	}
	configVersion, err := t.Client.ConfigurationVersions.Create(context.TODO(), workspaceID, options)
	if err != nil {
		return nil, err
	}
	return configVersion, nil
}

// CreateTerraformTemplate creates a template for the Terraform configuration
func CreateTerraformTemplate(workspace *v1alpha1.Workspace) ([]byte, error) {
	tfTemplate, err := template.New("main.tf").Parse(`terraform {
		backend "remote" {
			organization = "{{.Spec.Organization}}"
	
			workspaces {
				name = "{{.ObjectMeta.Namespace}}-{{.ObjectMeta.Name}}"
			}
		}
	}
	{{- range .Spec.Variables}}
	{{- if not .EnvironmentVariable }}
	variable "{{.Key}}" {}
	{{- end}}
	{{- end}}
	{{- range .Spec.Outputs}}
	output "{{.Key}}" {
		value = module.operator.{{.Attribute}}
	}
	{{- end}}
	module "operator" {
		source = "{{.Spec.Module.Source}}"
		version = "{{.Spec.Module.Version}}"
		{{- range .Spec.Variables}}
		{{- if not .EnvironmentVariable }}
		{{.Key}} = var.{{.Key}}
		{{- end}}
		{{- end}}
	}`)
	if err != nil {
		return nil, err
	}
	var tpl bytes.Buffer
	if err := tfTemplate.Execute(&tpl, workspace); err != nil {
		return nil, err
	}
	return tpl.Bytes(), nil
}

// CreateRunForTerraformConfiguration runs a new Terraform Cloud configuration
func (t *TerraformCloudClient) CreateRunForTerraformConfiguration(workspace *v1alpha1.Workspace, terraform []byte) error {
	configVersion, err := t.CreateConfigurationVersion(workspace.Status.WorkspaceID)
	if err != nil {
		return err
	}

	os.Mkdir(moduleDirectory, 0777)
	if err := ioutil.WriteFile(configurationFilePath, terraform, 0777); err != nil {
		return err
	}

	if err := t.UploadConfigurationFile(configVersion.UploadURL); err != nil {
		return err
	}

	message := fmt.Sprintf("operator, apply, configHash, %s", GetMD5Hash(terraform))
	options := tfc.RunCreateOptions{
		Message:              &message,
		ConfigurationVersion: configVersion,
		Workspace: &tfc.Workspace{
			ID: workspace.Status.WorkspaceID,
		},
	}
	run, err := t.Client.Runs.Create(context.TODO(), options)
	if err != nil {
		return err
	}
	workspace.Status.RunID = run.ID
	return nil
}

// CheckRun gets the run status
func (t *TerraformCloudClient) CheckRun(workspace *v1alpha1.Workspace) error {
	run, err := t.Client.Runs.Read(context.TODO(), workspace.Status.RunID)
	if err != nil {
		return err
	}

	workspace.Status.RunStatus = string(run.Status)
	return nil
}

func isPending(status string) bool {
	state := tfc.RunStatus(status)
	switch state {
	case tfc.RunApplied:
		return false
	case tfc.RunPlannedAndFinished:
		return false
	case tfc.RunErrored:
		return false
	case tfc.RunCanceled:
		return false
	case tfc.RunDiscarded:
		return false
	default:
		return true
	}
}

// CreateRun generates a run with the last config version
func (t *TerraformCloudClient) CreateRun(workspace *v1alpha1.Workspace) error {
	message := fmt.Sprintf("operator, apply, variable change")
	options := tfc.RunCreateOptions{
		Message: &message,
		Workspace: &tfc.Workspace{
			ID: workspace.Status.WorkspaceID,
		},
	}
	run, err := t.Client.Runs.Create(context.TODO(), options)
	if err != nil {
		return err
	}
	workspace.Status.RunID = run.ID
	return nil
}

// DeleteRuns cancels runs that haven't been applied or planned
func (t *TerraformCloudClient) DeleteRuns(workspaceID string) error {
	message := "operator, finalizer, cancelling run"
	runs, err := t.Client.Runs.List(context.TODO(), workspaceID, tfc.RunListOptions{})
	if err != nil {
		return err
	}
	for _, run := range runs.Items {
		switch status := run.Status; status {
		case tfc.RunApplied:
			continue
		case tfc.RunPlannedAndFinished:
			continue
		default:
			err := t.Client.Runs.ForceCancel(context.TODO(), run.ID, tfc.RunForceCancelOptions{
				Comment: &message,
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// DeleteResources destroys the resources in a workspace
func (t *TerraformCloudClient) DeleteResources(workspaceID string) error {
	ws, err := t.Client.Workspaces.ReadByID(context.TODO(), workspaceID)
	if err != nil {
		return err
	}
	message := "operator, destroy, latest"
	options := tfc.RunCreateOptions{
		IsDestroy: &isDestroy,
		Message:   &message,
		Workspace: ws,
	}
	run, err := t.Client.Runs.Create(context.TODO(), options)
	if err != nil {
		return err
	}
	for {
		checkRun, err := t.Client.Runs.Read(context.TODO(), run.ID)
		if err != nil || checkRun.Status == tfc.RunErrored {
			return fmt.Errorf("destroy had error: %v", err)
		}
		if checkRun.Status == tfc.RunApplied {
			return nil
		}
		time.Sleep(interval)
	}
}
