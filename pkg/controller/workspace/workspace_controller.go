package workspace

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

var log = logf.Log.WithName("controller_workspace")

const workspaceFinalizer = "finalizer.workspace.app.terraform.io"

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Workspace Controller and adds it to the Manager. The Manager will set fields on the Controller
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
	return &ReconcileWorkspace{
		client:   mgr.GetClient(),
		scheme:   mgr.GetScheme(),
		tfclient: tfclient,
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("workspace-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Workspace
	err = c.Watch(&source.Kind{Type: &appv1alpha1.Workspace{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Workspace
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appv1alpha1.Workspace{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileWorkspace implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileWorkspace{}

// ReconcileWorkspace reconciles a Workspace object
type ReconcileWorkspace struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client   client.Client
	scheme   *runtime.Scheme
	tfclient *TerraformCloudClient
}

// Reconcile reads that state of the cluster for a Workspace object and makes changes based on the state read
// and what is in the Workspace.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileWorkspace) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Name", request.Name, "Request.Namespace", request.Namespace)
	reqLogger.Info("Reconciling Workspace")

	// Fetch the Workspace instance
	instance := &appv1alpha1.Workspace{}
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
	organization := instance.Spec.Organization
	r.tfclient.Organization = organization
	workspace := request.Name

	if err := r.tfclient.CheckOrganization(); err != nil {
		reqLogger.Error(err, "Could not find organization")
		return reconcile.Result{}, nil
	}

	reqLogger.Info("Checking workspace", "Organization", organization, "Name", workspace, "Namespace", request.Namespace)
	workspaceID, err := r.tfclient.CheckWorkspace(workspace)
	if err != nil {
		reqLogger.Error(err, "Could not update workspace")
		return reconcile.Result{}, err
	}
	reqLogger.Info("Found workspace", "Organization", organization, "Name", workspace, "ID", workspaceID, "Namespace", request.Namespace)
	instance.Status.WorkspaceID = workspaceID

	// Check if the Workspace instance is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	markedForDeletion := instance.GetDeletionTimestamp() != nil
	if markedForDeletion {
		if contains(instance.GetFinalizers(), workspaceFinalizer) {
			if err := r.finalizeWorkspace(reqLogger, instance); err != nil {
				return reconcile.Result{}, err
			}

			// Remove workspaceFinalizer. Once all finalizers have been
			// removed, the object will be deleted.
			instance.SetFinalizers(remove(instance.GetFinalizers(), workspaceFinalizer))
			err := r.client.Update(context.TODO(), instance)
			if err != nil {
				return reconcile.Result{}, err
			}
		}
		return reconcile.Result{}, nil
	}

	// Add finalizer for this CR
	if !contains(instance.GetFinalizers(), workspaceFinalizer) {
		if err := r.addFinalizer(reqLogger, instance); err != nil {
			return reconcile.Result{}, err
		}
	}

	r.tfclient.SecretsMountPath = instance.Spec.SecretsMountPath

	if err := r.tfclient.CheckSecretsMountPath(); err != nil {
		reqLogger.Error(err, "Could not find secrets mount path")
		return reconcile.Result{}, nil
	}

	reqLogger.Info("Check variables exist in workspace", "Organization", organization, "Name", workspace, "Namespace", request.Namespace)
	specTFCVariables := MapToTFCVariable(instance.Spec.Variables)
	updatedVariables, err := r.tfclient.CheckVariables(workspace, specTFCVariables)
	if err != nil {
		reqLogger.Error(err, "Could not update variables")
		return reconcile.Result{}, err
	}

	reqLogger.Info("Creating terraform template", "Organization", organization, "Name", workspace, "Namespace", request.Namespace)
	terraform, err := CreateTerraformTemplate(instance)
	if err != nil {
		reqLogger.Error(err, "Could not create Terraform configuration")
		return reconcile.Result{}, err
	}

	md5 := GetMD5Hash(terraform)
	if instance.Status.ConfigHash != md5 {
		reqLogger.Info("Starting run because template changed", "Organization", organization, "Name", workspace, "Namespace", request.Namespace)
		err := r.tfclient.CreateRunForTerraformConfiguration(instance, terraform)
		if err != nil {
			reqLogger.Error(err, "Could not run new Terraform configuration")
			return reconcile.Result{}, err
		}
		instance.Status.ConfigHash = md5
	} else if updatedVariables {
		reqLogger.Info("Starting run because variable changed", "Organization", organization, "Name", workspace, "Namespace", request.Namespace)
		err := r.tfclient.CreateRun(instance)
		if err != nil {
			reqLogger.Error(err, "Could not run new variable changes")
			return reconcile.Result{}, err
		}
	}

	reqLogger.Info("Update run status", "Organization", organization, "Name", workspace, "Namespace", request.Namespace)
	for isPending(instance.Status.RunStatus) {
		if err := r.tfclient.CheckRun(instance); err != nil {
			reqLogger.Error(err, "could not get run information")
			return reconcile.Result{}, err
		}

		if err := r.client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "Failed to update Workspace status")
			return reconcile.Result{}, err
		}
	}

	if err := r.client.Status().Update(context.TODO(), instance); err != nil {
		reqLogger.Error(err, "Failed to update Workspace status")
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileWorkspace) finalizeWorkspace(reqLogger logr.Logger, workspace *appv1alpha1.Workspace) error {
	reqLogger.Info("Deleting runs in workspace", "Name", workspace.Name, "Namespace", workspace.Namespace)
	if err := r.tfclient.DeleteRuns(workspace.Status.WorkspaceID); err != nil {
		return err
	}
	reqLogger.Info("Deleting resources in workspace", "Name", workspace.Name, "Namespace", workspace.Namespace)
	if err := r.tfclient.DeleteResources(workspace.Name); err != nil {
		return err
	}
	reqLogger.Info("Deleting workspace", "Name", workspace.Name, "Namespace", workspace.Namespace)
	if err := r.tfclient.DeleteWorkspace(workspace.Name); err != nil {
		reqLogger.Error(err, "Could not delete workspace")
	}
	reqLogger.Info("Successfully finalized organization")
	return nil
}

func (r *ReconcileWorkspace) addFinalizer(reqLogger logr.Logger, workspace *appv1alpha1.Workspace) error {
	reqLogger.Info("Adding Finalizer for the Workspace")
	workspace.SetFinalizers(append(workspace.GetFinalizers(), workspaceFinalizer))

	// Update CR
	err := r.client.Update(context.TODO(), workspace)
	if err != nil {
		reqLogger.Error(err, "Failed to update Workspace with finalizer")
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
