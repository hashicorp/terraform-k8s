package workspacehelper

import (
	"context"
	"errors"
	"reflect"

	"github.com/hashicorp/terraform-k8s/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
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

func secretForOutputs(name string, namespace string, outputs []*v1alpha1.OutputStatus) *corev1.Secret {
	data := outputsToMap(outputs)
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: data,
	}
}

// UpsertTerraformConfig creates a ConfigMap for the Terraform template if it doesn't exist already
func (r *WorkspaceHelper) UpsertTerraformConfig(w *v1alpha1.Workspace, template []byte) (bool, error) {
	found := &corev1.ConfigMap{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: w.Name, Namespace: w.Namespace}, found)
	if err != nil && k8serrors.IsNotFound(err) {
		configMap := configMapForTerraform(w.Name, w.Namespace, template)
		err := controllerutil.SetControllerReference(w, configMap, r.scheme)
		if err != nil {
			return false, err
		}
		r.reqLogger.Info("Writing to new Terraform ConfigMap")
		if err := r.client.Create(context.TODO(), configMap); err != nil {
			r.reqLogger.Error(err, "Failed to create new Terraform ConfigMap")
			return false, err
		}
		return true, nil
	} else if err != nil {
		r.reqLogger.Error(err, "Failed to get Terraform ConfigMap")
		return false, err
	}

	if found.Data[TerraformConfigMap] == string(template) {
		return false, nil
	}

	found.Data[TerraformConfigMap] = string(template)
	if err := r.client.Update(context.TODO(), found); err != nil {
		r.reqLogger.Error(err, "Failed to update Terraform ConfigMap", "Namespace", w.Namespace, "Name", w.Name)
		return false, err
	}
	return true, nil
}

// GetConfigMapForVariable retrieves the configmap value associated with the variable
func (r *WorkspaceHelper) GetConfigMapForVariable(namespace string, variable *v1alpha1.Variable) error {
	if variable.Sensitive || variable.Value != "" {
		return nil
	}

	if variable.ValueFrom == nil {
		err := errors.New("non-sensitive variables require a value")
		r.reqLogger.Error(err, "No value specified", "Namespace", namespace, "Variable", variable.Key)
		return err
	}

	if variable.ValueFrom.ConfigMapKeyRef == nil {
		err := errors.New("Include ConfigMap in ValueFrom")
		r.reqLogger.Error(err, "No ConfigMap specified", "Namespace", namespace, "Variable", variable.Key)
		return err
	}

	r.reqLogger.Info("Checking ConfigMap for variable", "Namespace", namespace, "Variable", variable.Key)

	found := &corev1.ConfigMap{}
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

func outputsToMap(outputs []*v1alpha1.OutputStatus) map[string][]byte {
	data := map[string][]byte{}
	for _, output := range outputs {
		data[output.Key] = []byte(output.Value)
	}
	return data
}

// CleanPreviousSecretOutputsIfChanged cleans previous secret if output has changed
func (r *WorkspaceHelper) CleanPreviousSecretOutputsIfChanged(outputName string, outputNamespace string, status *v1alpha1.WorkspaceStatus) error {
	currentName := status.OutputName
	currentNamespace := status.OutputNamespace

	// no previous name was set, do nothing
	if currentName == "" || currentNamespace == "" {
		return nil
	}

	// no changes to name/namespace
	if currentName == outputName && currentNamespace == outputNamespace {
		return nil
	}

	// attempt to cleanup whatever is configured, if it exists
	r.reqLogger.Info("Terraform Output Secret name has changed. Removing previous version.")
	found := &corev1.Secret{}

	err := r.client.Get(context.TODO(), types.NamespacedName{Name: currentName, Namespace: currentNamespace}, found)
	if err != nil && k8serrors.IsNotFound(err) {
		// resource already removed, do nothing
		return nil
	} else if err != nil {
		return err
	}

	if err := r.client.Delete(context.TODO(), found); err != nil {
		// unable to remove object
		return err
	}

	return nil
}

// UpsertSecretOutputs creates a Secret for the outputs
func (r *WorkspaceHelper) UpsertSecretOutputs(outputName string, outputNamespace string, w *v1alpha1.Workspace) error {
	found := &corev1.Secret{}

	if err := r.CleanPreviousSecretOutputsIfChanged(outputName, outputNamespace, &w.Status); err != nil {
		r.reqLogger.Error(err, "Failed to clean previous output Secret")
		return err
	}

	err := r.client.Get(context.TODO(), types.NamespacedName{Name: outputName, Namespace: outputNamespace}, found)
	if err != nil && k8serrors.IsNotFound(err) {
		secret := secretForOutputs(outputName, outputNamespace, w.Status.Outputs)
		err = controllerutil.SetControllerReference(w, secret, r.scheme)
		if err != nil {
			return err
		}
		r.reqLogger.Info("Writing outputs to new Secret")
		if err := r.client.Create(context.TODO(), secret); err != nil {
			r.reqLogger.Error(err, "Failed to create new output secrets")
			return err
		}
		return nil
	} else if err != nil {
		r.reqLogger.Error(err, "Failed to get output secrets")
		return err
	}

	currentOutputs := outputsToMap(w.Status.Outputs)
	if !reflect.DeepEqual(found.Data, currentOutputs) {
		r.reqLogger.Info("Updating secrets", "name", outputName)
		found.Data = currentOutputs
		if err := r.client.Update(context.TODO(), found); err != nil {
			r.reqLogger.Error(err, "Failed to update output secrets", "Namespace", outputNamespace, "Name", outputName)
			return err
		}
		return nil
	}
	return nil
}
