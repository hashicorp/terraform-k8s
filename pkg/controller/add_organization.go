package controller

import (
	"github.com/hashicorp/terraform-k8s/pkg/controller/organization"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, organization.Add)
}
