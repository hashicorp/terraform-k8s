package workspace

import (
	"fmt"
	"strings"
	"testing"

	tfc "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-k8s/operator/pkg/apis/app/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestUpdateVariablesLastApplied(t *testing.T) {
	hashed, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	oldhashed, _ := bcrypt.GenerateFromPassword([]byte("oldhashed"), bcrypt.DefaultCost)
	tests := map[string]struct {
		lastApplied           [2]v1alpha1.LastAppliedVariableValues
		setWorkspaceVariables bool
		updated               bool
	}{
		"Variable Creation": {
			lastApplied: [2]v1alpha1.LastAppliedVariableValues{
				{},
				{},
			},
			setWorkspaceVariables: false,
			updated:               true,
		},
		"Variable Unchanged": {
			lastApplied: [2]v1alpha1.LastAppliedVariableValues{
				{"simple": "simple"},
				{"sensitive": string(hashed)},
			},
			setWorkspaceVariables: false,
			updated:               false,
		},
		"Variable Changed": {
			lastApplied: [2]v1alpha1.LastAppliedVariableValues{
				{"simple": "oldsimple"},
				{"sensitive": string(oldhashed)},
			},
			setWorkspaceVariables: true,
			updated:               true,
		},
	}

	mockVariablesClient := &MockVariablesClient{}
	mockVariablesClient.Mock.On(
		"Create",
		mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("tfe.VariableCreateOptions"),
	).Return(&tfc.Variable{}, nil).Twice()
	mockVariablesClient.Mock.On(
		"Update",
		mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("tfe.VariableUpdateOptions"),
	).Return(&tfc.Variable{}, nil).Twice()

	tfcClient := &TerraformCloudClient{}
	tfcClient.Client = &tfc.Client{Variables: mockVariablesClient}
	workspace := &tfc.Workspace{}

	for name, test := range tests {
		fmt.Printf("Running %s Tests\n", name)

		tts := simpleTests(test.lastApplied, string(hashed))
		workspaceVariables := make(map[string]*Variable)
		for key, tt := range tts {
			if test.setWorkspaceVariables {
				workspaceVariables[key] = &Variable{&tfc.Variable{Key: key, Sensitive: tt.sensitive, ID: key}, "", false}
			}
			lastApplied := tt.lastApplied
			specVariables := make(map[string]*Variable)
			specVariables[key] = newTestVariable(key, tt.value, tt.hashed, tt.sensitive, false, false, false)
			updated, _ := tfcClient.updateVariablesOnTFC(workspace, specVariables, workspaceVariables, lastApplied)
			assert.Equal(t, test.updated, updated)
			assert.Equal(t, tt.expected, lastApplied)
		}
	}

	mockVariablesClient.AssertExpectations(t)
}

func newTestVariable(key, value, hashed string, sensitive, environment, hcl, alwaysUpdate bool) *Variable {
	return &Variable{
		&tfc.Variable{
			Key:       key,
			Value:     strings.TrimSuffix(value, "\n"),
			Sensitive: sensitive,
			Category:  setVariableType(environment),
			HCL:       setHCL(hcl),
		},
		hashed,
		alwaysUpdate,
	}
}

func simpleTests(la [2]v1alpha1.LastAppliedVariableValues, hashed string) map[string]struct {
	value       string
	hashed      string
	sensitive   bool
	expected    v1alpha1.LastAppliedVariableValues
	lastApplied v1alpha1.LastAppliedVariableValues
} {
	return map[string]struct {
		value       string
		hashed      string
		sensitive   bool
		expected    v1alpha1.LastAppliedVariableValues
		lastApplied v1alpha1.LastAppliedVariableValues
	}{
		"simple": {
			value:       "simple",
			hashed:      "",
			sensitive:   false,
			expected:    v1alpha1.LastAppliedVariableValues{"simple": "simple"},
			lastApplied: la[0],
		},
		"sensitive": {
			value:       "password",
			hashed:      hashed,
			sensitive:   true,
			expected:    v1alpha1.LastAppliedVariableValues{"sensitive": hashed},
			lastApplied: la[1],
		},
	}
}
