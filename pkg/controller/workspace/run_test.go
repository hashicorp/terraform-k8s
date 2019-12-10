package workspace

import (
	"testing"

	"github.com/hashicorp/terraform-k8s/pkg/apis/app/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestShouldCreateTerraformWithMultipleVariables(t *testing.T) {
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
					Key:       "some_var",
					Value:     "here",
					Sensitive: false,
				},
				&v1alpha1.Variable{
					Key:       "hello",
					Value:     "world",
					Sensitive: false,
				},
			},
		},
	}
	terraformFile, err := createTerraformConfiguration(workspace)
	assert.Nil(t, err)
	assert.Equal(t, expectedFile, terraformFile.String())
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
	terraformFile, err := createTerraformConfiguration(workspace)
	assert.Nil(t, err)
	assert.Equal(t, expectedFile, terraformFile.String())
}
