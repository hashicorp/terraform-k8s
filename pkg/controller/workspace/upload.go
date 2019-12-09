package workspace

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"net/http"
	"os"
	"text/template"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-k8s/pkg/apis/app/v1alpha1"
)

const (
	ConfigurationFileEndpoint = "https://archivist.terraform.io/v1/object"
	ConfigurationFileName     = "main.tf"
	ConfigurationTarball      = "configuration.tar.gz"
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

func UploadConfigurationFiles(workspace *v1alpha1.Workspace) error {
	uuid, err := uuid.NewUUID()
	if err != nil {
		return err
	}
	client := &http.Client{}
	url := fmt.Sprintf("%s/%s", ConfigurationFileEndpoint, uuid.String())
	data, err := createTerraformConfiguration(workspace)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, url, data)
	if err != nil {
		return err
	}
	_, err = client.Do(req)
	if err != nil {
		return err
	}
	return nil
}
