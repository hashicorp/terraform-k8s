package workspace

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-k8s/operator/pkg/apis/app/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

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
	if err != nil && k8serrors.IsNotFound(err) {
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
			r.reqLogger.Error(err, "Failed to update ConfigMap", "Namespace", w.Namespace, "Name", w.Name)
			return updated, err
		}
		return true, nil
	}
	return updated, nil
}

// GetConfigMapForVariable retrieves the configmap value associated with the variable
func (r *ReconcileWorkspace) GetConfigMapForVariable(namespace string, variable *v1alpha1.Variable) error {
	if variable.Sensitive || variable.Value != "" {
		return nil
	}

	if variable.ValueFrom.ConfigMapKeyRef == nil {
		err := errors.New("Include ConfigMap in ValueFrom")
		r.reqLogger.Error(err, "No ConfigMap specified", "Namespace", namespace, "Variable", variable.Key)
		return err
	}

	r.reqLogger.Info("Checking ConfigMap for variable", "Namespace", namespace, "Variable", variable.Key)

	found := &v1.ConfigMap{}
	name := variable.ValueFrom.ConfigMapKeyRef.LocalObjectReference.Name
	key := variable.ValueFrom.ConfigMapKeyRef.Key
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, found)
	if err != nil {
		r.reqLogger.Error(err, "Did not find configmap", "Namespace", namespace, "Name", name)
		return err
	}
	value, ok := found.Data[key]
	if !ok {
		err := errors.New("Include ConfigMap key reference in ValueFrom")
		r.reqLogger.Error(err, "No ConfigMap key specified", "Namespace", namespace, "Name", name, "Key", key)
		return err
	}
	variable.Value = value
	return nil
}
