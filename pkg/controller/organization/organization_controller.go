package organization

import (
	"context"

	"github.com/go-logr/logr"
	appv1alpha1 "github.com/hashicorp/terraform-k8s/pkg/apis/app/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_organization")

const organizationFinalizer = "finalizer.app.terraform.io"

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Organization Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	tfclient := &TerraformCloudClient{}
	err := tfclient.GetClient()
	if err != nil {
		log.Error(err, "could not create Terraform Cloud client")
	}
	return &ReconcileOrganization{
		client:   mgr.GetClient(),
		scheme:   mgr.GetScheme(),
		tfclient: tfclient,
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("organization-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Organization
	err = c.Watch(&source.Kind{Type: &appv1alpha1.Organization{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Organization
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appv1alpha1.Organization{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileOrganization implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileOrganization{}

// ReconcileOrganization reconciles a Organization object
type ReconcileOrganization struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client   client.Client
	scheme   *runtime.Scheme
	tfclient *TerraformCloudClient
}

// Reconcile reads that state of the cluster for a Organization object and makes changes based on the state read
// and what is in the Organization.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileOrganization) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	organization := request.Name
	workspace := request.Namespace
	reqLogger := log.WithValues("Request.Organization", organization, "Request.Workspace", workspace)
	reqLogger.Info("Reconciling Workspace")

	instance := &appv1alpha1.Organization{}
	r.tfclient.Organization = organization
	r.tfclient.Workspace = workspace
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Check if the Organization instance is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	markedForDeletion := instance.GetDeletionTimestamp() != nil
	if markedForDeletion {
		if contains(instance.GetFinalizers(), organizationFinalizer) {
			if err := r.finalizeOrganization(reqLogger, instance); err != nil {
				return reconcile.Result{}, err
			}

			// Remove organizationFinalizer. Once all finalizers have been
			// removed, the object will be deleted.
			instance.SetFinalizers(remove(instance.GetFinalizers(), organizationFinalizer))
			err := r.client.Update(context.TODO(), instance)
			if err != nil {
				return reconcile.Result{}, err
			}
		}
		return reconcile.Result{}, nil
	}

	// Add finalizer for this CR
	if !contains(instance.GetFinalizers(), organizationFinalizer) {
		if err := r.addFinalizer(reqLogger, instance); err != nil {
			return reconcile.Result{}, err
		}
	}

	r.tfclient.SecretsMountPath = instance.Spec.SecretsMountPath

	if err := r.tfclient.CheckOrganization(); err != nil {
		reqLogger.Error(err, "Could not find organization")
		return reconcile.Result{}, nil
	}

	if err := r.tfclient.CheckSecretsMountPath(); err != nil {
		reqLogger.Error(err, "Could not find secrets mount path")
		return reconcile.Result{}, nil
	}

	reqLogger.Info("Checking workspace", "Organization", organization, "Name", workspace)
	workspaceID, err := r.tfclient.CheckWorkspace(workspace)
	if err != nil {
		reqLogger.Error(err, "Could not update workspace")
		return reconcile.Result{}, err
	}
	reqLogger.Info("Found workspace", "Organization", organization, "Name", workspace, "ID", workspaceID)
	instance.Status.WorkspaceID = workspaceID

	reqLogger.Info("Check variables exist in workspace", "Organization", organization, "Name", workspace)
	updatedVariables, err := r.tfclient.CheckVariables(workspace, instance.Spec.Variables)
	if err != nil {
		reqLogger.Error(err, "Could not update variables")
		return reconcile.Result{}, err
	}
	reqLogger.Info("Updated variables", "Organization", organization, "Name", workspace)

	reqLogger.Info("Run if needed", "Organization", organization, "Name", workspace)
	if err := r.tfclient.CheckRunConfiguration(instance, updatedVariables); err != nil {
		reqLogger.Error(err, "Could not execute run")
		return reconcile.Result{}, err
	}

	reqLogger.Info("Check run status", "Organization", organization, "Name", workspace, "Run", instance.Status.RunID)
	if err := r.tfclient.CheckRunForError(instance); err != nil {
		reqLogger.Error(err, "Run has error")
		return reconcile.Result{}, err
	}
	reqLogger.Info("Plan and apply executed", "Organization", organization, "Name", workspace, "Run", instance.Status.RunID)

	if err := r.client.Status().Update(context.TODO(), instance); err != nil {
		reqLogger.Error(err, "Failed to update Workspace status")
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileOrganization) finalizeOrganization(reqLogger logr.Logger, m *appv1alpha1.Organization) error {
	reqLogger.Info("Deleting resources", "Organization", m.Name, "Name", m.Namespace)
	err := r.tfclient.RunDelete(m.Namespace)
	reqLogger.Info("Deleting workspace", "Organization", m.Name, "Name", m.Namespace)
	err = r.tfclient.DeleteWorkspace(m.Namespace)
	if err != nil {
		reqLogger.Error(err, "Could not delete workspace")
	}
	reqLogger.Info("Successfully finalized organization")
	return nil
}

func (r *ReconcileOrganization) addFinalizer(reqLogger logr.Logger, m *appv1alpha1.Organization) error {
	reqLogger.Info("Adding Finalizer for the Organization")
	m.SetFinalizers(append(m.GetFinalizers(), organizationFinalizer))

	// Update CR
	err := r.client.Update(context.TODO(), m)
	if err != nil {
		reqLogger.Error(err, "Failed to update Organization with finalizer")
		return err
	}
	return nil
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func remove(list []string, s string) []string {
	for i, v := range list {
		if v == s {
			list = append(list[:i], list[i+1:]...)
		}
	}
	return list
}
