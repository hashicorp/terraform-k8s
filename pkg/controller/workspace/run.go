package workspace

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"text/template"

	tfc "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-k8s/pkg/apis/app/v1alpha1"
)

const (
	ConfigurationFileName    = "main.tf"
	ConfigurationTarballName = "configuration.tar.gz"
)

var (
	AutoQueueRuns         = false
	Speculative           = false
	_, b, _, _            = runtime.Caller(0)
	basepath              = filepath.Dir(b)
	moduleDirectory       = fmt.Sprintf("%s/%s", basepath, "module")
	ConfigurationFilePath = basepath + "/" + ConfigurationTarballName
)

type RunConfiguration struct {
	ConfigurationHash string
	RunID             string
}

func createTerraformConfiguration(workspace *v1alpha1.Workspace) (*bytes.Buffer, error) {
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
	if err := ioutil.WriteFile(moduleDirectory+"/"+ConfigurationFileName, data.Bytes(), 0777); err != nil {
		return err
	}
	return nil
}

func createConfigurationTarGz(data *bytes.Buffer) error {
	file, err := os.Create(basepath + "/" + ConfigurationTarballName)
	if err != nil {
		return err
	}
	defer file.Close()

	gzw := gzip.NewWriter(file)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	hdr := &tar.Header{
		Name: ConfigurationFileName,
		Mode: 0744,
		Size: int64(len(data.Bytes())),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	if _, err := tw.Write(data.Bytes()); err != nil {
		return err
	}
	if err := tw.Close(); err != nil {
		return err
	}
	return nil
}

func (t *TerraformCloudClient) UploadConfigurationFile(data *bytes.Buffer, uploadURL string) error {
	if err := writeToFile(data); err != nil {
		return err
	}
	if err := t.Client.ConfigurationVersions.Upload(context.TODO(), uploadURL, moduleDirectory); err != nil {
		return fmt.Errorf("error, %v, %v", err, moduleDirectory)
	}
	return nil
}

func (t *TerraformCloudClient) CreateConfigurationVersion(workspaceID string) (*tfc.ConfigurationVersion, error) {
	options := tfc.ConfigurationVersionCreateOptions{
		AutoQueueRuns: &AutoQueueRuns,
		Speculative:   &Speculative,
	}
	configVersion, err := t.Client.ConfigurationVersions.Create(context.TODO(), workspaceID, options)
	if err != nil {
		return nil, err
	}
	return configVersion, nil
}

func (t *TerraformCloudClient) CheckConfiguration(workspace *v1alpha1.Workspace) error {
	data, err := createTerraformConfiguration(workspace)
	if err != nil {
		return err
	}
	hash := md5.Sum(data.Bytes())
	md5Sum := hex.EncodeToString(hash[:])
	if workspace.Status.ConfigHash == md5Sum {
		return nil
	}

	workspace.Status.ConfigHash = md5Sum
	configVersion, err := t.CreateConfigurationVersion(workspace.Status.WorkspaceID)
	if err != nil {
		return err
	}
	if err := t.UploadConfigurationFile(data, configVersion.UploadURL); err != nil {
		return err
	}
	message := fmt.Sprintf("Operator creating plan for configuration hash: %s", md5Sum)
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

func (t *TerraformCloudClient) CheckPlanAndApply(workspace *v1alpha1.Workspace) error {
	run, err := t.Client.Runs.Read(context.TODO(), workspace.Status.RunID)
	if err != nil {
		return err
	}

	if run.Status == tfc.RunPlannedAndFinished {
		message := fmt.Sprintf("Operator applying run: %s", workspace.Status.RunID)
		options := tfc.RunApplyOptions{
			Comment: &message,
		}
		err := t.Client.Runs.Apply(context.TODO(), workspace.Status.RunID, options)
		if err != nil {
			return err
		}
	}

	if run.Status == tfc.RunErrored {
		return fmt.Errorf("run has error, runID, %s", run.ID)
	}

	return nil
}
