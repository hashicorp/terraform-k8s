package v1alpha1

import (
	"reflect"
	"strings"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateCustomResourceDefinition creates the CRD for the Workspace type
func CreateCustomResourceDefinition(clientSet apiextensionsclientset.Interface) (*apiextensionsv1beta1.CustomResourceDefinition, error) {
	kind := reflect.TypeOf(Workspace{}).Name()
	plural := strings.ToLower(kind + "s")
	name := strings.ToLower(plural + "." + SchemeGroupVersion.Group)

	crd := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   SchemeGroupVersion.Group,
			Version: SchemeGroupVersion.Version,
			Scope:   apiextensionsv1beta1.NamespaceScoped,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Plural: plural,
				Kind:   kind,
			},
		},
	}

	_, err := clientSet.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)

	if apierrors.IsAlreadyExists(err) {
		// TODO migration logic
		return crd, nil
	}

	if err != nil {
		return nil, err
	}

	return crd, nil
}
