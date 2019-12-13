package workspace

import (
	"testing"

	"github.com/hashicorp/terraform-k8s/pkg/apis/app/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestShouldCreateTerraformWithVariables(t *testing.T) {
	expectedFile := `terraform {
		backend "remote" {
			organization = "world"
	
			workspaces {
				name = "hello"
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
			Namespace: "world",
		},
		Spec: v1alpha1.WorkspaceSpec{
			Module: &v1alpha1.Module{
				Source:  "my_source",
				Version: "0.3.2",
			},
			Variables: []*v1alpha1.Variable{
				&v1alpha1.Variable{
					Key:                 "some_var",
					Value:               "here",
					Sensitive:           false,
					EnvironmentVariable: false,
				},
				&v1alpha1.Variable{
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
				name = "hello"
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
			Namespace: "world",
		},
		Spec: v1alpha1.WorkspaceSpec{
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
				name = "hello"
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
			Namespace: "world",
		},
		Spec: v1alpha1.WorkspaceSpec{
			Module: &v1alpha1.Module{
				Source:  "my_source",
				Version: "0.3.2",
			},
			Variables: []*v1alpha1.Variable{
				&v1alpha1.Variable{
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
				name = "hello"
			}
		}
	}
	output "module_output" {
		value = module.operator.my_output
	}
	output "ip" {
		value = module.operator.ip_address
	}
	module "operator" {
		source = "my_source"
		version = "0.3.2"
	}`

	workspace := &v1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hello",
			Namespace: "world",
		},
		Spec: v1alpha1.WorkspaceSpec{
			Module: &v1alpha1.Module{
				Source:  "my_source",
				Version: "0.3.2",
			},
			Outputs: []*v1alpha1.Output{
				&v1alpha1.Output{
					Key:       "module_output",
					Attribute: "my_output",
				},
				&v1alpha1.Output{
					Key:       "ip",
					Attribute: "ip_address",
				},
			},
		},
	}
	terraformFile, err := CreateTerraformTemplate(workspace)
	assert.Nil(t, err)
	assert.Equal(t, expectedFile, string(terraformFile))
}
