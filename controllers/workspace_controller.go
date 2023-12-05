package controllers

import (
	"context"

	appv1alpha1 "github.com/snyk/terraform-k8s/api/v1alpha1"
	"github.com/snyk/terraform-k8s/workspacehelper"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

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

// Add creates a new Workspace Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, NewWorkspaceReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func NewWorkspaceReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &WorkspaceReconciler{
		helper: workspacehelper.NewWorkspaceHelper(mgr),
	}
}

type WorkspaceReconciler struct {
	helper reconcile.Reconciler
}

// +kubebuilder:rbac:groups=app.terraform.io,resources=workspaces,verbs=get;list;watch;create;update;patch;delete,namespace=terraform-k8s
// +kubebuilder:rbac:groups=app.terraform.io,resources=workspaces/status,verbs=get;update;patch,namespace=terraform-k8s
func (r *WorkspaceReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	_ = context.Background()
	return r.helper.Reconcile(context.TODO(), req)
}

func (r *WorkspaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1alpha1.Workspace{}).
		Complete(r)
}
