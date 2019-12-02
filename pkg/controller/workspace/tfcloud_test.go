package workspace

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func setup() *TerraformCloudClient {
	tfc := &TerraformCloudClient{}
	tfc.GetClient()
	return tfc
}

func TestShouldGetClientConfig(t *testing.T) {
	tfc := setup()
	assert.NotNil(t, tfc.Client)
}
