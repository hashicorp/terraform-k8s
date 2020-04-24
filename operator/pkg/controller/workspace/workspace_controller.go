package workspace

import (
	"context"
	"fmt"
	"reflect"

	"github.com/go-logr/logr"
	"github.com/hashicorp/terraform-k8s/operator/pkg/apis/app/v1alpha1"
	appv1alpha1 "github.com/hashicorp/terraform-k8s/operator/pkg/apis/app/v1alpha1"
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

var log = logf.Log.WithName(TerraformOperator)

const workspaceFinalizer = "finalizer.workspace.app.terraform.io"

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
		client:    mgr.GetClient(),
		scheme:    mgr.GetScheme(),
		tfclient:  tfclient,
		reqLogger: log,
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

	err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForOwner{
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
	client    client.Client
	scheme    *runtime.Scheme
	tfclient  *TerraformCloudClient
	reqLogger logr.Logger
}

// Reconcile reads that state of the cluster for a Workspace object and makes changes based on the state read
// and what is in the Workspace.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileWorkspace) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	r.reqLogger.WithValues("Name", request.Name, "Namespace", request.Namespace)
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
	workspace := fmt.Sprintf("%s-%s", request.Namespace, request.Name)

	if err := r.tfclient.CheckOrganization(); err != nil {
		r.reqLogger.Error(err, "Could not find organization")
		return reconcile.Result{}, nil
	}

	r.tfclient.SecretsMountPath = instance.Spec.SecretsMountPath
	if err := r.tfclient.CheckSecretsMountPath(); err != nil {
		r.reqLogger.Error(err, "Could not find secrets mount path")
		return reconcile.Result{}, nil
	}

	workspaceID, err := r.tfclient.CheckWorkspace(workspace)
	if err != nil {
		r.reqLogger.Error(err, "Could not update workspace")
		return reconcile.Result{}, err
	}

	if instance.Status.WorkspaceID != workspaceID {
		instance.Status.WorkspaceID = workspaceID
		instance.Status.Outputs = []*v1alpha1.OutputStatus{}
		if err := r.client.Status().Update(context.TODO(), instance); err != nil {
			r.reqLogger.Error(err, "Failed to update output status")
			return reconcile.Result{}, err
		}
		r.reqLogger.Info("Updated workspace ID", "Organization", organization, "WorkspaceID", instance.Status.WorkspaceID)
	}

	// Check if the Workspace instance is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	markedForDeletion := instance.GetDeletionTimestamp() != nil
	err = r.tfclient.CheckWorkspacebyID(instance.Status.WorkspaceID)
	if markedForDeletion || err != nil {
		if contains(instance.GetFinalizers(), workspaceFinalizer) {
			if err := r.finalizeWorkspace(r.reqLogger, instance); err != nil {
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
		if err := r.addFinalizer(r.reqLogger, instance); err != nil {
			return reconcile.Result{}, err
		}
	}

	if isPending(instance.Status.RunStatus) {
		r.reqLogger.Info("Run incomplete", "Organization", organization, "RunID", instance.Status.RunID, "RunStatus", instance.Status.RunStatus)
		runStatus, err := r.tfclient.CheckRun(instance.Status.RunID)
		if err != nil {
			r.reqLogger.Error(err, "Could not get run ID")
			return reconcile.Result{}, err
		}

		if instance.Status.RunStatus != runStatus {
			instance.Status.RunStatus = runStatus
			if err := r.client.Status().Update(context.TODO(), instance); err != nil {
				r.reqLogger.Error(err, "Failed to update Workspace status")
				return reconcile.Result{}, err
			}
		}
		return reconcile.Result{Requeue: true}, nil
	}

	r.reqLogger.Info("Checking outputs", "Organization", organization, "WorkspaceID", instance.Status.WorkspaceID, "RunID", instance.Status.RunID)
	if !isError(instance.Status.RunStatus) {
		outputs, err := r.tfclient.CheckOutputs(instance.Status.WorkspaceID, instance.Status.RunID)
		if err != nil {
			r.reqLogger.Error(err, "Could not get run ID")
			return reconcile.Result{}, err
		}

		if !reflect.DeepEqual(outputs, instance.Status.Outputs) {
			instance.Status.Outputs = outputs
			err := r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				r.reqLogger.Error(err, "Failed to update output status")
				return reconcile.Result{}, err
			}
			r.reqLogger.Info("Updated outputs", "Organization", organization, "WorkspaceID", instance.Status.WorkspaceID)
		}

		if err = r.UpsertOutputs(instance, instance.Status.Outputs); err != nil {
			r.reqLogger.Error(err, "Error with creating ConfigMap for Terraform Outputs")
			return reconcile.Result{}, err
		}
	}

	terraform, err := CreateTerraformTemplate(instance)
	if err != nil {
		r.reqLogger.Error(err, "Could not create Terraform configuration")
		return reconcile.Result{}, err
	}

	updatedTerraform, err := r.UpsertTerraformConfig(instance, terraform)
	if err != nil {
		r.reqLogger.Error(err, "Error with creating ConfigMap for Terraform Configuration")
		return reconcile.Result{}, err
	}

	for _, variable := range instance.Spec.Variables {
		err := r.GetConfigMapForVariable(request.Namespace, variable)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	specTFCVariables := MapToTFCVariable(instance.Spec.Variables)

	updatedVariables, err := r.tfclient.CheckVariables(workspace, specTFCVariables, instance.Status.LastAppliedVariableValues)
	if err != nil {
		r.reqLogger.Error(err, "Could not update variables")
		return reconcile.Result{}, err
	}

	if updatedTerraform || updatedVariables || instance.Status.RunID == "" {
		r.reqLogger.Info("Starting run because template changed", "Organization", organization, "Name", workspace, "Namespace", request.Namespace)

		if err := r.tfclient.CreateRun(instance, terraform); err != nil {
			r.reqLogger.Error(err, "Could not run new Terraform configuration")
			return reconcile.Result{}, err
		}

		err := r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			r.reqLogger.Error(err, "Failed to update run ID")
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileWorkspace) finalizeWorkspace(reqLogger logr.Logger, workspace *appv1alpha1.Workspace) error {
	if err := r.tfclient.CheckWorkspacebyID(workspace.Status.WorkspaceID); err == nil {
		reqLogger.Info("Stopping runs in workspace", "Name", workspace.Name, "Namespace", workspace.Namespace)
		if err := r.tfclient.DeleteRuns(workspace.Status.WorkspaceID); err != nil {
			return err
		}
		reqLogger.Info("Deleting resources in workspace", "Name", workspace.Name, "Namespace", workspace.Namespace)
		if err := r.tfclient.DeleteResources(workspace.Status.WorkspaceID); err != nil {
			return err
		}
		reqLogger.Info("Deleting workspace", "Name", workspace.Name, "Namespace", workspace.Namespace)
		if err := r.tfclient.DeleteWorkspace(workspace.Status.WorkspaceID); err != nil {
			reqLogger.Error(err, "Could not delete workspace")
		}
	}
	reqLogger.Info("Successfully finalized workspace")
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
