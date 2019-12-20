package workspace

import (
	"bytes"
	"context"
	"text/template"

	"github.com/hashicorp/terraform-k8s/operator/pkg/apis/app/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	// TerraformConfigMap names the Terraform template ConfigMap
	TerraformConfigMap = "terraform"
	// TerraformOperator is for labelling
	TerraformOperator = "terraform-k8s"
)

// CreateTerraformTemplate creates a template for the Terraform configuration
func CreateTerraformTemplate(workspace *v1alpha1.Workspace) ([]byte, error) {
	tfTemplate, err := template.New("main.tf").Parse(`terraform {
		backend "remote" {
			organization = "{{.Spec.Organization}}"
	
			workspaces {
				name = "{{.ObjectMeta.Namespace}}-{{.ObjectMeta.Name}}"
			}
		}
	}
	{{- range .Spec.Variables}}
	{{- if not .EnvironmentVariable }}
	variable "{{.Key}}" {}
	{{- end}}
	{{- end}}
	{{- range .Spec.Outputs}}
	output "{{.Key}}" {
		value = module.operator.{{.ModuleOutputName}}
	}
	{{- end}}
	module "operator" {
		source = "{{.Spec.Module.Source}}"
		version = "{{.Spec.Module.Version}}"
		{{- range .Spec.Variables}}
		{{- if not .EnvironmentVariable }}
		{{.Key}} = var.{{.Key}}
		{{- end}}
		{{- end}}
	}`)
	if err != nil {
		return nil, err
	}
	var tpl bytes.Buffer
	if err := tfTemplate.Execute(&tpl, workspace); err != nil {
		return nil, err
	}
	return tpl.Bytes(), nil
}

func configMapForTerraform(name string, namespace string, template []byte) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: map[string]string{
			TerraformConfigMap: string(template),
		},
	}
}

// UpsertConfigMap creates a ConfigMap for the Terraform template if it doesn't exist already
func (r *ReconcileWorkspace) UpsertConfigMap(w *v1alpha1.Workspace, template []byte) (bool, error) {
	updated := false
	found := &v1.ConfigMap{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: w.Name, Namespace: w.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		configMap := configMapForTerraform(w.Name, w.Namespace, template)
		controllerutil.SetControllerReference(w, configMap, r.scheme)
		r.reqLogger.Info("Writing terraform to new ConfigMap")
		if err := r.client.Create(context.TODO(), configMap); err != nil {
			r.reqLogger.Error(err, "Failed to create new ConfigMap")
			return updated, err
		}
		return true, nil
	} else if err != nil {
		r.reqLogger.Error(err, "Failed to get ConfigMap")
		return updated, err
	}

	if found.Data[TerraformConfigMap] != string(template) {
		found.Data[TerraformConfigMap] = string(template)
		if err := r.client.Update(context.TODO(), found); err != nil {
			r.reqLogger.Error(err, "Failed to update ConfigMap", "Namespace", w.Name, "Name", w.Namespace)
			return updated, err
		}
		return true, nil
	}
	return updated, nil
}
