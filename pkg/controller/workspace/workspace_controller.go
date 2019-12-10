package workspace

import (
	"context"

	appv1alpha1 "github.com/hashicorp/terraform-k8s/pkg/apis/app/v1alpha1"
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
	reqLogger := log.WithValues("Request.Organization", request.Namespace, "Request.Workspace", request.Name)
	reqLogger.Info("Reconciling Workspace")

	// Fetch the Workspace instance
	instance := &appv1alpha1.Workspace{}
	r.tfclient.Organization = request.Namespace
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("Deleting resources", "Organization", request.Namespace, "Name", request.Name)
			err := r.tfclient.RunDelete(request.Name)
			reqLogger.Info("Deleting workspace", "Organization", request.Namespace, "Name", request.Name)
			err = r.tfclient.DeleteWorkspace(request.Name)
			if err != nil {
				reqLogger.Error(err, "Could not delete workspace")
			}
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if err := r.tfclient.CheckOrganization(); err != nil {
		reqLogger.Error(err, "Could not find organization")
		return reconcile.Result{}, nil
	}

	reqLogger.Info("Checking workspace", "Organization", request.Namespace, "Name", request.Name)
	workspaceID, err := r.tfclient.CheckWorkspace(request.Name)
	if err != nil {
		reqLogger.Error(err, "Could not update workspace")
		return reconcile.Result{}, err
	}
	reqLogger.Info("Found workspace", "Organization", request.Namespace, "Name", request.Name, "ID", workspaceID)
	instance.Status.WorkspaceID = workspaceID

	reqLogger.Info("Check variables exist in workspace", "Organization", request.Namespace, "Name", request.Name)
	if err := r.tfclient.CheckVariables(request.Name, instance.Spec.Variables); err != nil {
		reqLogger.Error(err, "Could not update variables")
		return reconcile.Result{}, err
	}
	reqLogger.Info("Updated variables", "Organization", request.Namespace, "Name", request.Name)

	reqLogger.Info("Run if needed", "Organization", request.Namespace, "Name", request.Name)
	if err := r.tfclient.CheckRunConfiguration(instance); err != nil {
		reqLogger.Error(err, "Could not execute run")
		return reconcile.Result{}, err
	}

	reqLogger.Info("Check run status", "Organization", request.Namespace, "Name", request.Name, "Run", instance.Status.RunID)
	if err := r.tfclient.CheckRunForError(instance); err != nil {
		reqLogger.Error(err, "Run has error")
		return reconcile.Result{}, err
	}
	reqLogger.Info("Plan and apply executed", "Organization", request.Namespace, "Name", request.Name, "Run", instance.Status.RunID)

	if err := r.client.Status().Update(context.TODO(), instance); err != nil {
		reqLogger.Error(err, "Failed to update Workspace status")
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}
