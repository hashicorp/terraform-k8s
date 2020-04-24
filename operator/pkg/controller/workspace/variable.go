package workspace

import (
	"fmt"
	"io/ioutil"
	"strings"

	tfc "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-k8s/operator/pkg/apis/app/v1alpha1"
	"golang.org/x/crypto/bcrypt"
)

// Wrap tfe.Variable
type Variable struct {
	*tfc.Variable
	hashed       string
	AlwaysUpdate bool
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
func (v *Variable) Changed(la v1alpha1.LastAppliedVariableValues) bool {
	if _, ok := la[v.Key]; !ok {
		return true
	}

	if v.Sensitive {
		return bcrypt.CompareHashAndPassword([]byte(la[v.Key]), []byte(v.Value)) != nil
	}
	return v.Value != la[v.Key]
}

// Update values in LastAppliedVariableValues
func (v *Variable) SetStatus(la v1alpha1.LastAppliedVariableValues) bool {
	if v.Sensitive {
		la[v.Key] = v.Hashed()
	} else {
		la[v.Key] = v.Value
	}
	return true
}
