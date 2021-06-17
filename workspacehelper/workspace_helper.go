package workspacehelper

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	"github.com/hashicorp/go-tfe"
	tfc "github.com/hashicorp/go-tfe"
	appv1alpha1 "github.com/hashicorp/terraform-k8s/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var log = logf.Log.WithName(TerraformOperator)

const workspaceFinalizer = "finalizer.workspace.app.terraform.io"
const requeueInterval = time.Minute * 1

type WorkspaceHelper struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client    client.Client
	scheme    *runtime.Scheme
	tfclient  *TerraformCloudClient
	reqLogger logr.Logger
	recorder  record.EventRecorder
}

func NewWorkspaceHelper(mgr manager.Manager) reconcile.Reconciler {
	tfclient := &TerraformCloudClient{}
	err := tfclient.GetClient(os.Getenv("TF_URL"))
	if err != nil {
		log.Error(err, "could not create Terraform Cloud or Enterprise client")
		os.Exit(1)
	}
	return &WorkspaceHelper{
		client:    mgr.GetClient(),
		scheme:    mgr.GetScheme(),
		tfclient:  tfclient,
		reqLogger: log,
		recorder:  mgr.GetEventRecorderFor("workspace"),
	}
}

func (r *WorkspaceHelper) finalizeWorkspace(reqLogger logr.Logger, workspace *appv1alpha1.Workspace) error {
	if err := r.tfclient.CheckWorkspacebyID(workspace.Status.WorkspaceID); err == nil {
		reqLogger.Info("Stopping runs in workspace",
			"Name", workspace.Name, "Namespace", workspace.Namespace)
		if err := r.tfclient.DeleteRuns(workspace.Status.WorkspaceID); err != nil {
			return err
		}
		reqLogger.Info("Deleting resources in workspace", "Name", workspace.Name,
			"Namespace", workspace.Namespace)
		if err := r.tfclient.DeleteResources(workspace.Status.WorkspaceID); err != nil {
			return err
		}
		reqLogger.Info("Deleting workspace", "Name", workspace.Name,
			"Namespace", workspace.Namespace)
		if err := r.tfclient.DeleteWorkspace(workspace.Status.WorkspaceID); err != nil {
			reqLogger.Error(err, "Could not delete workspace")
		}
	}
	reqLogger.Info("Successfully finalized workspace")
	return nil
}

func (r *WorkspaceHelper) addFinalizer(reqLogger logr.Logger, workspace *appv1alpha1.Workspace) error {
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

func (r *WorkspaceHelper) initializeReconciliation(request reconcile.Request) (*appv1alpha1.Workspace, error) {
	// Fetch the Workspace instance
	instance := &appv1alpha1.Workspace{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return nil, nil
		}
		// Error reading the object - requeue the request.
		return nil, err
	}

	if instance.Spec.VCS == nil && instance.Spec.Module == nil {
		msg := fmt.Sprintf("Either VCS or Module need to be specified in spec for workspace %s", instance.Name)
		r.recorder.Event(instance, corev1.EventTypeWarning, "WorkspaceEvent", msg)
	}

	r.tfclient.Organization = instance.Spec.Organization
	if err := r.tfclient.CheckOrganization(); err != nil {
		r.reqLogger.Error(err, "Could not find organization", "Organization", instance.Spec.Organization)
		return nil, err
	}

	r.tfclient.SecretsMountPath = instance.Spec.SecretsMountPath
	if err := r.tfclient.CheckSecretsMountPath(); err != nil {
		r.reqLogger.Error(err, "Could not find secrets mount path")
		return nil, err
	}

	return instance, nil
}

func (r *WorkspaceHelper) reconcileWorkspace(instance *appv1alpha1.Workspace) error {
	workspace := fmt.Sprintf("%s-%s", instance.Namespace, instance.Name)
	organization := instance.Spec.Organization

	ws, err := r.tfclient.CheckWorkspace(workspace, instance)
	if err != nil {
		r.reqLogger.Error(err, "Could not update workspace")
		return err
	}
	workspaceID := ws.ID

	if instance.Status.WorkspaceID != workspaceID {
		instance.Status.WorkspaceID = workspaceID
		instance.Status.Outputs = []*appv1alpha1.OutputStatus{}
		if err := r.client.Status().Update(context.TODO(), instance); err != nil {
			r.reqLogger.Error(err, "Failed to update output status")
			return err
		}
		r.reqLogger.Info("Updated workspace ID", "Organization", organization,
			"WorkspaceID", instance.Status.WorkspaceID)
	}

	if instance.Status.RunID != "" && ws.CurrentRun != nil && instance.Status.RunID != ws.CurrentRun.ID {
		instance.Status.RunID = ws.CurrentRun.ID
		if err := r.client.Status().Update(context.TODO(), instance); err != nil {
			r.reqLogger.Error(err, "Failed to update workspace status")
			return err
		}
		r.recorder.Event(instance, corev1.EventTypeNormal, "WorkspaceEvent",
			"Updated outputs after out of band run applied")
		r.reqLogger.Info("Updated Run ID", "Organization", organization,
			"WorkspaceID", instance.Status.WorkspaceID, "RunID", instance.Status.RunID)
	}

	// Add finalizer for this CR
	if !contains(instance.GetFinalizers(), workspaceFinalizer) {
		if err := r.addFinalizer(r.reqLogger, instance); err != nil {
			return err
		}
	}

	return nil
}

func (r *WorkspaceHelper) condDeleteWorkspace(instance *appv1alpha1.Workspace) (bool, error) {
	markedForDeletion := instance.GetDeletionTimestamp() != nil
	if !markedForDeletion {
		return false, nil
	}

	err := r.tfclient.CheckWorkspacebyID(instance.Status.WorkspaceID)
	if err != nil && err != tfe.ErrResourceNotFound {
		return false, err
	}

	if contains(instance.GetFinalizers(), workspaceFinalizer) {
		if err := r.finalizeWorkspace(r.reqLogger, instance); err != nil {
			return false, err
		}
	}

	// Remove workspaceFinalizer. Once all finalizers have been
	// removed, the object will be deleted.
	instance.SetFinalizers(remove(instance.GetFinalizers(), workspaceFinalizer))
	err = r.client.Update(context.TODO(), instance)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *WorkspaceHelper) runInProgress(instance *appv1alpha1.Workspace) (bool, error) {
	if !isPending(instance.Status.RunStatus) {
		return false, nil
	}

	r.reqLogger.Info("Run incomplete", "Organization", instance.Spec.Organization,
		"RunID", instance.Status.RunID, "RunStatus", instance.Status.RunStatus)
	runStatus, err := r.tfclient.CheckRun(instance.Status.RunID)
	if err != nil {
		r.reqLogger.Error(err, "Could not get run ID")
		return false, err
	}

	if instance.Status.RunStatus != runStatus {
		instance.Status.RunStatus = runStatus
		if err := r.client.Status().Update(context.TODO(), instance); err != nil {
			r.reqLogger.Error(err, "Failed to update Workspace status")
			return false, err
		}
	}
	return true, nil
}

func (r *WorkspaceHelper) processFinishedRun(instance *appv1alpha1.Workspace) error {
	if isError(instance.Status.RunStatus) {
		msg := fmt.Sprintf("Run %q for workspace %q failed to complete",
			instance.Status.RunID, instance.Status.WorkspaceID)
		r.recorder.Event(instance, corev1.EventTypeNormal,
			"WorkspaceEvent", msg)

		// stick to the original implementation
		// perhaps we should throw an error instead
		return nil
	}

	r.reqLogger.Info("Checking outputs", "Organization",
		instance.Spec.Organization, "WorkspaceID", instance.Status.WorkspaceID,
		"RunID", instance.Status.RunID)
	outputs, err := r.tfclient.CheckOutputs(instance.Status.WorkspaceID, instance.Status.RunID)
	if err != nil {
		r.reqLogger.Error(err, "Could not get run ID")
		return err
	}

	if !reflect.DeepEqual(outputs, instance.Status.Outputs) {
		instance.Status.Outputs = outputs
		err := r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			r.reqLogger.Error(err, "Failed to update output status")
			return err
		}
		r.reqLogger.Info("Updated outputs", "Organization",
			instance.Spec.Organization, "WorkspaceID", instance.Status.WorkspaceID)
		r.recorder.Event(instance, corev1.EventTypeNormal, "WorkspaceEvent",
			fmt.Sprintf("Updated outputs for run %s", instance.Status.RunID))
	}
	if err = r.UpsertSecretOutputs(instance, instance.Status.Outputs); err != nil {
		r.reqLogger.Error(err, "Error with creating ConfigMap for Terraform Outputs")
		return err
	}
	return nil
}

func (r *WorkspaceHelper) updateTerraformTemplate(instance *appv1alpha1.Workspace) (bool, error) {
	if instance.Spec.VCS != nil {
		return false, nil
	}

	terraform, err := CreateTerraformTemplate(instance)
	if err != nil {
		r.reqLogger.Error(err, "Could not create Terraform configuration")
		return false, err
	}

	updated, err := r.UpsertTerraformConfig(instance, terraform)
	if err != nil {
		r.reqLogger.Error(err, "Error with creating ConfigMap for Terraform Configuration")
		return false, err
	}
	return updated, nil
}

func (r *WorkspaceHelper) updateVariables(instance *appv1alpha1.Workspace) (bool, error) {
	workspace := fmt.Sprintf("%s-%s", instance.Namespace, instance.Name)

	for _, variable := range instance.Spec.Variables {
		err := r.GetConfigMapForVariable(instance.Namespace, variable)
		if err != nil {
			return false, err
		}
	}

	specTFCVariables := MapToTFCVariable(instance.Spec.Variables)
	updatedVariables, err := r.tfclient.CheckVariables(workspace, specTFCVariables)
	if err != nil {
		r.reqLogger.Error(err, "Could not update variables")
		return false, err
	}

	return updatedVariables, nil
}

func (r *WorkspaceHelper) prepareModuleRun(instance *appv1alpha1.Workspace, options tfe.RunCreateOptions) (bool, error) {
	r.reqLogger.Info("Starting module backed run", "Organization",
		instance.Spec.Organization, "Name", instance.Name, "Namespace", instance.Namespace)

	var (
		configVersion *tfe.ConfigurationVersion
		err           error
	)
	if instance.Status.ConfigVersionID == "" {
		cfgMap := &corev1.ConfigMap{}
		err = r.client.Get(context.TODO(),
			types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, cfgMap)
		if err != nil {
			return true, err
		}

		tf := []byte(cfgMap.Data[TerraformConfigMap])
		configVersion, err = r.tfclient.CreateConfigurationVersion(instance.Status.WorkspaceID)
		if err != nil {
			return true, err
		}

		if _, err := os.Stat(moduleDirectory); os.IsNotExist(err) {
			if err = os.Mkdir(moduleDirectory, 0777); err != nil {
				return true, err
			}
		}

		if err = ioutil.WriteFile(configurationFilePath, tf, 0777); err != nil {
			return true, err
		}

		if err = r.tfclient.UploadConfigurationFile(configVersion.UploadURL); err != nil {
			return true, err
		}

		instance.Status.ConfigVersionID = configVersion.ID
		if err = r.client.Status().Update(context.TODO(), instance); err != nil {
			r.reqLogger.Error(err, "Failed to update Workspace status")
			return true, err
		}
	}

	configVersion, err = r.tfclient.Client.ConfigurationVersions.Read(context.TODO(), instance.Status.ConfigVersionID)
	if err != nil {
		return true, err
	}
	if !(configVersion.Status == tfc.ConfigurationUploaded) {
		return true, err
	}

	// Reset Configuration Version ID on the Workspace Status since we are done with this one
	instance.Status.ConfigVersionID = ""
	if err = r.client.Status().Update(context.TODO(), instance); err != nil {
		r.reqLogger.Error(err, "Failed to update Workspace status")
		return true, err
	}

	options.ConfigurationVersion = configVersion

	return false, nil
}

func (r *WorkspaceHelper) prepareVCSRun(instance *appv1alpha1.Workspace) (bool, error) {
	r.reqLogger.Info("Starting VCS backed run", "Organization",
		instance.Spec.Organization, "Name", instance.Name, "Namespace", instance.Namespace)

	configVersions, err := r.tfclient.Client.ConfigurationVersions.List(context.TODO(),
		instance.Status.WorkspaceID, tfe.ConfigurationVersionListOptions{})
	if err != nil {
		return false, err
	}

	if len(configVersions.Items) == 0 {
		r.reqLogger.Info("ConfigVersion is not available yet", "Organization",
			instance.Spec.Organization, "Name", instance.Name, "Namespace", instance.Namespace)
		return true, nil
	}
	return false, nil
}

func (r *WorkspaceHelper) startRun(instance *appv1alpha1.Workspace) error {
	message := fmt.Sprintf("%s, apply", TerraformOperator)
	options := tfe.RunCreateOptions{
		Message: &message,
		Workspace: &tfe.Workspace{
			ID: instance.Status.WorkspaceID,
		},
	}

	var err error
	requeue := false
	if instance.Spec.VCS != nil {
		requeue, err = r.prepareVCSRun(instance)
	} else if instance.Spec.Module != nil {
		requeue, err = r.prepareModuleRun(instance, options)
	}

	if err != nil || requeue {
		// When requeue is true it means that the workspace is not ready yet.
		return err
	}

	runResult, err := r.tfclient.Client.Runs.Create(context.TODO(), options)
	if err != nil {
		return err
	}

	instance.Status.RunID = runResult.ID
	instance.Status.RunStatus = string(runResult.Status)
	if err := r.client.Status().Update(context.TODO(), instance); err != nil {
		r.reqLogger.Error(err, "Failed to update Workspace status")
		return err
	}
	return nil
}

func (r *WorkspaceHelper) reconcileNotifications(instance *appv1alpha1.Workspace) error {
	err := r.tfclient.CheckNotifications(instance)
	if err != nil {
		r.reqLogger.Error(err, "while checking notifications")
		return err
	}
	return nil
}

// Reconcile reads that state of the cluster for a Workspace object and makes changes based on the state read
// and what is in the Workspace.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *WorkspaceHelper) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Get instance, check if org and secrets exist or not
	instance, err := r.initializeReconciliation(request)
	if err != nil {
		return reconcile.Result{}, err
	} else if instance == nil {
		// Instance got garbage collected
		return reconcile.Result{}, nil
	}

	// Check if the object is pending deletion and act accordingly
	deleted, err := r.condDeleteWorkspace(instance)
	if err != nil {
		return reconcile.Result{}, err
	} else if deleted {
		return reconcile.Result{}, nil
	}

	// Create the workspace update instance with workspace information such as Workspace ID and Run ID.
	err = r.reconcileWorkspace(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Check if notifications exist, create them if they don't
	err = r.reconcileNotifications(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	// check the run status and update the instance
	// returns instantly if the run is not in progress
	shouldRequeue, err := r.runInProgress(instance)
	if err != nil {
		return reconcile.Result{}, err
	} else if shouldRequeue {
		return reconcile.Result{Requeue: true}, nil
	}

	// process the run result
	err = r.processFinishedRun(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Figure out if the terraform config has been updated for non VCS backed workspaces
	updatedTerraform, err := r.updateTerraformTemplate(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	// make sure the variables in the tfc workspace match the ones in the workspace
	// k8s object and update them if there is a difference.
	//
	// the k8s object is always the source of truth.
	updatedVariables, err := r.updateVariables(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	if updatedTerraform || updatedVariables || instance.Status.RunID == "" || instance.Status.ConfigVersionID != "" {
		err := r.startRun(instance)
		if err != nil {
			return reconcile.Result{}, err
		}

		r.recorder.Event(instance, corev1.EventTypeNormal, "WorkspaceEvent",
			fmt.Sprintf("Started new Terraform job with id %s", instance.Status.RunID))
		return reconcile.Result{Requeue: true}, nil
	}

	// We're asking the operator to reconcile again after a while because one of the following
	// may happen:
	//
	// - If we're using a VCS backed workspace it could be possible that a VCS event triggered
	//   a new run.
	// - A new run was manually queued via the UI.
	// - Someone changed the workspace variables via the UI and we need to change them back.
	//
	// It is important to note that if any event takes place between requeues will not be blocked.
	return reconcile.Result{RequeueAfter: requeueInterval}, nil
}
