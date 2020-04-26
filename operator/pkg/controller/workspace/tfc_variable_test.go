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
		lastApplied           [2]*v1alpha1.LastApplied
		setWorkspaceVariables bool
		updated               bool
	}{
		"Variable Creation": {
			lastApplied: [2]*v1alpha1.LastApplied{
				{Values: map[string]string{}, Attributes: map[string]byte{"simple": 0}},
				{Values: map[string]string{}, Attributes: map[string]byte{"sensitive": 2}},
			},
			setWorkspaceVariables: false,
			updated:               true,
		},
		"Variable Unchanged": {
			lastApplied: [2]*v1alpha1.LastApplied{
				{Values: map[string]string{"simple": "simple"}, Attributes: map[string]byte{"simple": 0}},
				{Values: map[string]string{"sensitive": string(hashed)}, Attributes: map[string]byte{"sensitive": 2}},
			},
			setWorkspaceVariables: false,
			updated:               false,
		},
		"Variable Changed": {
			lastApplied: [2]*v1alpha1.LastApplied{
				{Values: map[string]string{"simple": "oldsimple"}, Attributes: map[string]byte{"simple": 0}},
				{Values: map[string]string{"sensitive": string(oldhashed)}, Attributes: map[string]byte{"sensitive": 2}},
			},
			setWorkspaceVariables: true,
			updated:               true,
		},
		"Attribute Changed": {
			lastApplied: [2]*v1alpha1.LastApplied{
				{Values: map[string]string{"simple": "simple"}, Attributes: map[string]byte{"simple": 1}},
				{Values: map[string]string{"sensitive": string(hashed)}, Attributes: map[string]byte{"sensitive": 3}},
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
	).Return(&tfc.Variable{}, nil).Times(4)

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

func simpleTests(la [2]*v1alpha1.LastApplied, hashed string) map[string]struct {
	value       string
	hashed      string
	sensitive   bool
	hcl         bool
	environment bool
	expected    *v1alpha1.LastApplied
	lastApplied *v1alpha1.LastApplied
} {
	return map[string]struct {
		value       string
		hashed      string
		sensitive   bool
		hcl         bool
		environment bool
		expected    *v1alpha1.LastApplied
		lastApplied *v1alpha1.LastApplied
	}{
		"simple": {
			value:       "simple",
			hashed:      "",
			sensitive:   false,
			hcl:         false,
			environment: false,
			expected:    &v1alpha1.LastApplied{Values: map[string]string{"simple": "simple"}, Attributes: map[string]byte{"simple": 0}},
			lastApplied: la[0],
		},
		"sensitive": {
			value:       "password",
			hashed:      hashed,
			sensitive:   true,
			hcl:         false,
			environment: false,
			expected:    &v1alpha1.LastApplied{Values: map[string]string{"sensitive": hashed}, Attributes: map[string]byte{"sensitive": 2}},
			lastApplied: la[1],
		},
	}
}
