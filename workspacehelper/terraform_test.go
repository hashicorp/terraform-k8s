package workspacehelper

import (
	"testing"

	"github.com/hashicorp/terraform-k8s/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestShouldCreateTerraformWithVariables(t *testing.T) {
	expectedFile := `terraform {
		backend "remote" {
			organization = "world"
	
			workspaces {
				name = "prod-hello"
			}
		}
	}
	variable "some_var" {}
	variable "hello" {}
	module "operator" {
		source = "my_source"
		version = "0.3.2"
		some_var = var.some_var
		hello = var.hello
	}`

	workspace := &v1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hello",
			Namespace: "prod",
		},
		Spec: v1alpha1.WorkspaceSpec{
			Organization: "world",
			Module: &v1alpha1.Module{
				Source:  "my_source",
				Version: "0.3.2",
			},
			Variables: []*v1alpha1.Variable{
				{
					Key:                 "some_var",
					Value:               "here",
					Sensitive:           false,
					EnvironmentVariable: false,
				},
				{
					Key:                 "hello",
					Value:               "world",
					Sensitive:           false,
					EnvironmentVariable: false,
				},
			},
		},
	}
	terraformFile, err := CreateTerraformTemplate(workspace)
	assert.Nil(t, err)
	assert.Equal(t, expectedFile, string(terraformFile))
}

func TestShouldCreateTerraformWithNoVariables(t *testing.T) {
	expectedFile := `terraform {
		backend "remote" {
			organization = "world"
	
			workspaces {
				name = "prod-hello"
			}
		}
	}
	module "operator" {
		source = "my_source"
		version = "0.3.2"
	}`

	workspace := &v1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hello",
			Namespace: "prod",
		},
		Spec: v1alpha1.WorkspaceSpec{
			Organization: "world",
			Module: &v1alpha1.Module{
				Source:  "my_source",
				Version: "0.3.2",
			},
		},
	}
	terraformFile, err := CreateTerraformTemplate(workspace)
	assert.Nil(t, err)
	assert.Equal(t, expectedFile, string(terraformFile))
}

func TestShouldCreateTerraformWithNoEnvironmentVariables(t *testing.T) {
	expectedFile := `terraform {
		backend "remote" {
			organization = "world"
	
			workspaces {
				name = "prod-hello"
			}
		}
	}
	module "operator" {
		source = "my_source"
		version = "0.3.2"
	}`

	workspace := &v1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hello",
			Namespace: "prod",
		},
		Spec: v1alpha1.WorkspaceSpec{
			Organization: "world",
			Module: &v1alpha1.Module{
				Source:  "my_source",
				Version: "0.3.2",
			},
			Variables: []*v1alpha1.Variable{
				{
					Key:                 "some_var",
					Value:               "here",
					Sensitive:           false,
					EnvironmentVariable: true,
				},
			},
		},
	}
	terraformFile, err := CreateTerraformTemplate(workspace)
	assert.Nil(t, err)
	assert.Equal(t, expectedFile, string(terraformFile))
}

func TestShouldCreateTerraformWithOutputs(t *testing.T) {
	expectedFile := `terraform {
		backend "remote" {
			organization = "world"
	
			workspaces {
				name = "prod-hello"
			}
		}
	}
	output "module_output" {
		value = module.operator.my_output
		sensitive = false
	}
	output "ip" {
		value = module.operator.ip_address
		sensitive = false
	}
	output "password" {
		value = module.operator.my_password
		sensitive = true
	}
	module "operator" {
		source = "my_source"
		version = "0.3.2"
	}`

	workspace := &v1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hello",
			Namespace: "prod",
		},
		Spec: v1alpha1.WorkspaceSpec{
			Organization: "world",
			Module: &v1alpha1.Module{
				Source:  "my_source",
				Version: "0.3.2",
			},
			Outputs: []*v1alpha1.OutputSpec{
				{
					Key:              "module_output",
					ModuleOutputName: "my_output",
				},
				{
					Key:              "ip",
					ModuleOutputName: "ip_address",
				},
				{
					Key:              "password",
					ModuleOutputName: "my_password",
					Sensitive:        true,
				},
			},
		},
	}
	terraformFile, err := CreateTerraformTemplate(workspace)
	assert.Nil(t, err)
	assert.Equal(t, expectedFile, string(terraformFile))
}

func TestShouldCreateTerraformWithNoModuleVersion(t *testing.T) {
	expectedFile := `terraform {
		backend "remote" {
			organization = "world"
	
			workspaces {
				name = "prod-hello"
			}
		}
	}
	module "operator" {
		source = "my_source"
	}`

	workspace := &v1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hello",
			Namespace: "prod",
		},
		Spec: v1alpha1.WorkspaceSpec{
			Organization: "world",
			Module: &v1alpha1.Module{
				Source: "my_source",
			},
		},
	}
	terraformFile, err := CreateTerraformTemplate(workspace)
	assert.Nil(t, err)
	assert.Equal(t, expectedFile, string(terraformFile))
}
