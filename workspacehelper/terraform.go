package workspacehelper

import (
	"bytes"
	"text/template"

	"github.com/hashicorp/terraform-k8s/api/v1alpha1"
)

const (
	// TerraformConfigMap names the Terraform template ConfigMap
	TerraformConfigMap = "terraform"
	// TerraformOperator is for labelling
	TerraformOperator = "terraform-k8s"
)

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
		value = module.operator.{{.ModuleOutputName}}
	}
	{{- end}}
	module "operator" {
		source = "{{.Spec.Module.Source}}"
		{{- if .Spec.Module.Version }}
		version = "{{.Spec.Module.Version}}"
		{{- end}}
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
