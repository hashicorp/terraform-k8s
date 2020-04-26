package workspace

import (
	"fmt"
	"io/ioutil"
	"strings"

	tfc "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-k8s/operator/pkg/apis/app/v1alpha1"
	"golang.org/x/crypto/bcrypt"
)

const (
	BIT_HCL         = 1
	BIT_SENSITIVE   = 2
	BIT_ENVIRONMENT = 4
)

// Wrap tfe.Variable
type Variable struct {
	*tfc.Variable
	hashed       string
	AlwaysUpdate bool
}

func setVariableType(isEnvironmentVariable bool) tfc.CategoryType {
	if isEnvironmentVariable {
		return tfc.CategoryEnv
	}
	return tfc.CategoryTerraform
}

func setHCL(isHCL bool) bool {
	if isHCL {
		return true
	}
	return false
}

// MapToTFCVariable changes the controller spec to a wrapped TFC Variable
func MapToTFCVariable(specVariables []*v1alpha1.Variable) map[string]*Variable {
	tfcVariables := make(map[string]*Variable)
	for _, variable := range specVariables {
		tfcVariables[variable.Key] = &Variable{
			&tfc.Variable{
				Key:       variable.Key,
				Value:     strings.TrimSuffix(variable.Value, "\n"),
				Sensitive: variable.Sensitive,
				Category:  setVariableType(variable.EnvironmentVariable),
				HCL:       setHCL(variable.HCL),
			},
			"",
			variable.AlwaysUpdate,
		}
	}
	return tfcVariables
}

func (v *Variable) CheckAndRetrieveIfSensitive(t *TerraformCloudClient) error {
	if v.Sensitive && v.Value == "" {
		filePath := fmt.Sprintf("%s/%s", t.SecretsMountPath, v.Key)
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("could not get secret, %s", err)
		}
		secret := string(data)
		v.Value = secret
		if err != nil {
			return fmt.Errorf("error generating hash of sensitive value, %s", err)
		}
	}
	return nil
}

// Get (and memoize) hash of Variable Value
func (v *Variable) Hashed() string {
	if v.hashed != "" {
		return v.hashed
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(v.Value), bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	v.hashed = string(hashed)
	return v.hashed
}

// Determine if the variable has changed from the LastAppliedVariableValues
func (v *Variable) Changed(la *v1alpha1.LastApplied) bool {
	// Doesn't exist
	if _, ok := la.Values[v.Key]; !ok {
		return true
	}

	// Attributes changed
	if v.attributeConfig() != la.Attributes[v.Key] {
		return true
	}

	// Value changed
	if v.Sensitive {
		return bcrypt.CompareHashAndPassword([]byte(la.Values[v.Key]), []byte(v.Value)) != nil
	}
	return v.Value != la.Values[v.Key]
}

// Update values in LastAppliedVariableValues
func (v *Variable) SetStatus(la *v1alpha1.LastApplied) bool {
	if v.Sensitive {
		la.Values[v.Key] = v.Hashed()
	} else {
		la.Values[v.Key] = v.Value
	}
	la.Attributes[v.Key] = v.attributeConfig()
	return true
}

func (v *Variable) attributeConfig() byte {
	var config byte
	if v.HCL {
		config |= BIT_HCL
	}
	if v.Sensitive {
		config |= BIT_SENSITIVE
	}
	if v.Category == tfc.CategoryEnv {
		config |= BIT_ENVIRONMENT
	}
	return config
}
