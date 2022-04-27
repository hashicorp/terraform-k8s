package workspacehelper

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-k8s/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// GetSecretData retrieves the data from a secret in a given namespace
func (r *WorkspaceHelper) GetSecretData(namespace string, name string) (map[string][]byte, error) {
	// If no secretName defined, return empty map
	if name == "" {
		return make(map[string][]byte), nil
	}

	r.reqLogger.Info("Getting Secret", "Namespace", namespace, "Name", name)

	secret := &corev1.Secret{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, secret)

	if err != nil {
		r.reqLogger.Error(err, "Failed to get Secret", "Namespace", namespace, "Name", name)

		return nil, err
	}

	return secret.Data, nil
}

// GetSecretForVariable retrieves the sensitive value associated with the variable from a secret
func (r *WorkspaceHelper) GetSecretForVariable(namespace string, variable *v1alpha1.Variable) error {
	if variable.Sensitive == false || variable.ValueFrom == nil {
		return nil
	}

	if variable.ValueFrom.SecretKeyRef == nil {
		err := errors.New("Include Secret in ValueFrom")

		r.reqLogger.Error(err, "No Secret specified", "Namespace", namespace, "Variable", variable.Key)

		return err
	}

	r.reqLogger.Info("Checking Secret for variable", "Namespace", namespace, "Variable", variable.Key)

	name := variable.ValueFrom.SecretKeyRef.LocalObjectReference.Name
	key := variable.ValueFrom.SecretKeyRef.Key

	data, err := r.GetSecretData(namespace, name)
	if err != nil {
		return err
	}

	value, ok := data[key]
	if !ok {
		err := errors.New("Include Secret key reference in ValueFrom")

		r.reqLogger.Error(err, "No Secret key specified", "Namespace", namespace, "Name", name, "Key", key)

		return err
	}

	variable.Value = string(value)

	return nil
}
