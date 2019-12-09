package workspace

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"net/http"
	"os"
	"text/template"

	tfc "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-k8s/pkg/apis/app/v1alpha1"
)

const (
	ConfigurationFileName = "main.tf"
	ConfigurationTarball  = "configuration.tar.gz"
)

var (
	AutoQueueRuns = true
	Speculative   = false
)

func createTerraformConfiguration(workspace *v1alpha1.Workspace) (*bytes.Buffer, error) {
	tfTemplate, err := template.New("main.tf").Parse(`terraform {
		required_version = "~0.12"
		backend "remote" {

			organization = "{{.ObjectMeta.Namespace}}"
	
			workspaces {
				name = "{{.ObjectMeta.Name}}"
			}
		}
	}

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

func createConfigurationTarGz(data *bytes.Buffer) error {
	file, err := os.Create("configuration.tar.gz")
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

func UploadConfigurationFiles(configVersion *tfc.ConfigurationVersion, workspace *v1alpha1.Workspace) error {
	client := &http.Client{}
	data, err := createTerraformConfiguration(workspace)
	if err != nil {
		return err
	}
	if err := createConfigurationTarGz(data); err != nil {
		return err
	}
	file, err := os.Open(ConfigurationTarball)
	if err != nil {
		return err
	}
	defer file.Close()
	req, err := http.NewRequest(http.MethodPut, configVersion.UploadURL, file)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/octet-stream")
	_, err = client.Do(req)
	if err != nil {
		return err
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
