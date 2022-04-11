package workspacehelper

import (
	"testing"

	tfc "github.com/hashicorp/go-tfe"
	"github.com/stretchr/testify/assert"
)

const (
	secretsMount = "mocks"
)

var (
	specVariables = []*tfc.Variable{
		{
			Key:       "test",
			HCL:       false,
			Value:     "needs update",
			Sensitive: true,
		},
		{
			Key:       "test2",
			HCL:       false,
			Sensitive: true,
		},
	}
	workspaceVariables = []*tfc.Variable{
		{
			Key:       "test",
			HCL:       true,
			Value:     "test",
			Sensitive: false,
		},
		{
			Key:       "test2",
			HCL:       false,
			Sensitive: true,
		},
	}
)

func TestShouldReturn1VariableNeedsUpdate(t *testing.T) {
	updatedVariables := getNonSensitiveVariablesToUpdate(specVariables, workspaceVariables)
	assert.Len(t, updatedVariables, 1)
	assert.Equal(t, updatedVariables[0].HCL, specVariables[0].HCL)
	assert.Equal(t, updatedVariables[0].Value, specVariables[0].Value)
	assert.Equal(t, updatedVariables[0].Sensitive, specVariables[0].Sensitive)
}

func TestShouldReturnNoVariableNeedsUpdateWhenChangingSensitive(t *testing.T) {
	specVariables := []*tfc.Variable{
		{
			Key:       "test",
			Sensitive: false,
		},
	}
	workspaceVariables := []*tfc.Variable{
		{
			Key:       "test",
			Sensitive: true,
		},
	}
	updatedVariables := getNonSensitiveVariablesToUpdate(specVariables, workspaceVariables)
	assert.Len(t, updatedVariables, 0)
}

func TestShouldReturnNoVariableNeedsUpdate(t *testing.T) {
	specVariables := []*tfc.Variable{
		{
			Key:       "test2",
			HCL:       false,
			Value:     "test2",
			Sensitive: false,
		},
	}
	workspaceVariables := []*tfc.Variable{
		{
			Key:       "test2",
			HCL:       false,
			Value:     "test2",
			Sensitive: false,
		},
	}
	updatedVariables := getNonSensitiveVariablesToUpdate(specVariables, workspaceVariables)
	assert.Len(t, updatedVariables, 0)
}

func TestShouldGetSensitiveVariablesForUpdate(t *testing.T) {
	specVariables := []*tfc.Variable{
		{
			Key:       "test",
			Sensitive: true,
		},
	}
	workspaceVariables := []*tfc.Variable{
		{
			Key:       "test",
			Sensitive: true,
		},
	}
	secretData := make(map[string][]byte)
	update, err := getSensitiveVariablesToUpdate(specVariables, workspaceVariables, secretsMount, secretData)
	assert.NoError(t, err)
	assert.Len(t, update, 1)
	assert.Equal(t, update[0].Key, specVariables[0].Key)
}

func TestShouldUpdateVariables(t *testing.T) {
	specVariables := []*tfc.Variable{
		{
			Key:       "test",
			Sensitive: true,
			HCL:       false,
		},
		{
			Key:       "test2",
			Value:     `{hello="world"}`,
			Sensitive: false,
			HCL:       true,
		},
	}
	workspaceVariables := []*tfc.Variable{
		{
			Key:       "test",
			Sensitive: true,
			HCL:       false,
		},
		{
			Key:       "test2",
			Value:     "hello world",
			Sensitive: false,
			HCL:       false,
		},
	}
	secretData := make(map[string][]byte)
	update, err := generateUpdateVariableList(specVariables, workspaceVariables, secretsMount, secretData)
	assert.NoError(t, err)
	assert.Len(t, update, 2)
	assert.False(t, update[0].Sensitive)
	assert.Equal(t, update[0].Key, specVariables[1].Key)
	assert.Equal(t, update[0].Value, specVariables[1].Value)
	assert.Equal(t, update[0].HCL, specVariables[1].HCL)
	assert.True(t, update[1].Sensitive)
	assert.Equal(t, update[1].Key, specVariables[0].Key)
	assert.Equal(t, update[1].HCL, specVariables[0].HCL)
}

func TestShouldNotUpdateVariables(t *testing.T) {
	specVariables := []*tfc.Variable{
		{
			Key:       "test",
			Sensitive: true,
			HCL:       false,
		},
		{
			Key:       "test2",
			Value:     `{hello="world"}`,
			Sensitive: false,
			HCL:       true,
		},
	}
	workspaceVariables := []*tfc.Variable{
		{
			Key:       "test",
			Sensitive: true,
			HCL:       false,
		},
		{
			Key:       "test2",
			Value:     `{hello="world"}`,
			Sensitive: false,
			HCL:       true,
		},
	}
	secretData := make(map[string][]byte)
	update, err := generateUpdateVariableList(specVariables, workspaceVariables, secretsMount, secretData)
	assert.NoError(t, err)
	assert.Len(t, update, 0)
}
