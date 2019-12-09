package workspace

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"testing"

	"github.com/hashicorp/terraform-k8s/pkg/apis/app/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestShouldCreateTerraformWithMultipleVariables(t *testing.T) {
	expectedFile := `terraform {
		required_version = "~0.12"
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
		required_version = "~0.12"
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

func TestShouldCreateTarball(t *testing.T) {
	data := bytes.NewBufferString(`terraform {
		required_version = "~0.12"
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
	}`)
	err := createConfigurationTarGz(data)
	assert.Nil(t, err)
	files, err := readTarFile(ConfigurationTarball)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(files))
	assert.Equal(t, files[0], ConfigurationFileName)
	os.Remove(ConfigurationTarball)
}

func readTarFile(srcFile string) ([]string, error) {
	files := []string{}
	f, err := os.Open(srcFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	gzf, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}

	tarReader := tar.NewReader(gzf)

	i := 0
	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		name := header.Name

		switch header.Typeflag {
		case tar.TypeDir:
			continue
		case tar.TypeReg:
			files = append(files, name)
		default:
			return files, err
		}

		i++
	}
	return files, nil
}
