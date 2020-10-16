package workspace

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupClient(t *testing.T, tfAddress string) (*TerraformCloudClient, error) {
	testOrganization := os.Getenv("TF_ORG")
	if os.Getenv("TF_ACC") == "" || os.Getenv("TF_CLI_CONFIG_FILE") == "" {
		t.Skipf("this test requires Terraform Cloud and Enterprise access and credentials;" +
			"set TF_ACC=1 and TF_CLI_CONFIG_FILE to run it")
	}
	tfClient := &TerraformCloudClient{
		Organization: testOrganization,
	}
	err := tfClient.GetClient(tfAddress)
	return tfClient, err
}

func TestFailsForNonMatchingTerraformEnterpriseHostname(t *testing.T) {
	_, err := setupClient(t, "notahostname")
	assert.Error(t, err)
}

func TestOrganizationTerraformCloud(t *testing.T) {
	tfClient, err := setupClient(t, "")
	assert.NoError(t, err)

	err = tfClient.CheckOrganization()
	assert.NoError(t, err)
}

func TestOrganizationTerraformEnterprise(t *testing.T) {
	if os.Getenv("TF_URL") == "" {
		t.Skipf("this test requires TF_URL for Terraform Enterprise")
	}
	tfClient, err := setupClient(t, os.Getenv("TF_URL"))
	assert.NoError(t, err)

	err = tfClient.CheckOrganization()
	assert.NoError(t, err)
}

func TestOrganizationTerraformEnterpriseNotFound(t *testing.T) {
	if os.Getenv("TF_URL") == "" {
		t.Skipf("this test requires TF_URL for Terraform Enterprise")
	}
	tfClient, err := setupClient(t, os.Getenv("TF_URL"))
	assert.NoError(t, err)

	tfClient.Organization = "doesnotexist"
	err = tfClient.CheckOrganization()
	assert.Error(t, err)
}

func TestShouldGetTerraformVersionFromEnvVariable(t *testing.T) {
	expected := "0.11.0"
	os.Setenv("TF_VERSION", expected)
	actual := getTerraformVersion()
	os.Unsetenv("TF_VERSION")
	assert.Equal(t, expected, *actual)
}

func TestShouldGetTerraformVersionFromOperator(t *testing.T) {
	expected := "0.13.4"
	actual := getTerraformVersion()
	assert.Equal(t, expected, *actual)
}
