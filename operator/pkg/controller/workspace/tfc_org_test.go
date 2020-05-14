package workspace

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldGetTerraformVersionFromEnvVariable(t *testing.T) {
	expected := "0.11.0"
	os.Setenv("TF_VERSION", expected)
	actual := getTerraformVersion()
	os.Unsetenv("TF_VERSION")
	assert.Equal(t, expected, *actual)
}

func TestShouldGetTerraformVersionFromOperator(t *testing.T) {
	expected := "0.12.25"
	actual := getTerraformVersion()
	assert.Equal(t, expected, *actual)
}
