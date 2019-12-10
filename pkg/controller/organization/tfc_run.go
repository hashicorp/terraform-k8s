package organization

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"text/template"
	"time"

	tfc "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-k8s/pkg/apis/app/v1alpha1"
)

var (
	autoQueueRuns         = false
	speculative           = false
	isDestroy             = true
	_, b, _, _            = runtime.Caller(0)
	basepath              = filepath.Dir(b)
	moduleDirectory       = fmt.Sprintf("%s/%s", basepath, "module")
	configurationFilePath = fmt.Sprintf("%s/%s", moduleDirectory, "main.tf")
	interval              = 30 * time.Second
)

func createTerraformConfiguration(workspace *v1alpha1.Organization) (*bytes.Buffer, error) {
	tfTemplate, err := template.New("main.tf").Parse(`terraform {
		backend "remote" {
			organization = "{{.ObjectMeta.Namespace}}"
	
			workspaces {
				name = "{{.ObjectMeta.Name}}"
			}
		}
	}
	{{- range .Spec.Variables}}
	variable "{{.Key}}" {}
	{{- end}}

	module "operator" {
		source = "{{.Spec.Module.Source}}"
		version = "{{.Spec.Module.Version}}"
		{{- range .Spec.Variables}}
		{{.Key}} = var.{{.Key}}
		{{- end}}
	}`)
	if err != nil {
		return nil, err
	}
	var tpl bytes.Buffer
	if err := tfTemplate.Execute(&tpl, workspace); err != nil {
		return nil, err
	}
	return &tpl, nil
}

func writeToFile(data *bytes.Buffer) error {
	os.Mkdir(moduleDirectory, 0777)
	if err := ioutil.WriteFile(configurationFilePath, data.Bytes(), 0777); err != nil {
		return err
	}
	return nil
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

// CheckRunConfiguration examines if there is a change in the Terraform configuration and
// runs the workspace if there is
func (t *TerraformCloudClient) CheckRunConfiguration(workspace *v1alpha1.Organization) error {
	data, err := createTerraformConfiguration(workspace)
	if err != nil {
		return err
	}
	hash := md5.Sum(data.Bytes())
	md5Sum := hex.EncodeToString(hash[:])
	if workspace.Status.ConfigHash == md5Sum {
		return nil
	}

	configVersion, err := t.CreateConfigurationVersion(workspace.Status.WorkspaceID)
	if err != nil {
		return err
	}

	workspace.Status.ConfigHash = md5Sum
	if err := writeToFile(data); err != nil {
		return err
	}
	if err := t.UploadConfigurationFile(configVersion.UploadURL); err != nil {
		return err
	}

	message := fmt.Sprintf("operator, apply, configHash, %s", md5Sum)
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

// CheckRunForError examines for errors in the run
func (t *TerraformCloudClient) CheckRunForError(workspace *v1alpha1.Organization) error {
	run, err := t.Client.Runs.Read(context.TODO(), workspace.Status.RunID)
	if err != nil {
		return err
	}

	if run.Status == tfc.RunErrored {
		return fmt.Errorf("run has error, runID, %s", run.ID)
	}

	return nil
}

// RunDelete destroys the latest configuration in a workspace
func (t *TerraformCloudClient) RunDelete(workspace string) error {
	ws, err := t.Client.Workspaces.Read(context.TODO(), t.Organization, workspace)
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
