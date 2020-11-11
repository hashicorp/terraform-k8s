package workspacehelper

import (
	"testing"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-k8s/api/v1alpha1"
	"github.com/hashicorp/terraform-k8s/workspacehelper/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestRunStartsDuringRequeueWhenConfigVersionIDSet(t *testing.T) {
	// Setup Workspace Resource
	workspace := &v1alpha1.Workspace{
		Spec: v1alpha1.WorkspaceSpec{
			Organization:     "Dummy Org",
			SecretsMountPath: "/tmp",
			Module: &v1alpha1.Module{
				Source:  "the-one-module-to-rule-them-all",
				Version: "1.19.0",
			},
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "awesome-workspace",
			Namespace: "terraform-system",
		},
	}

	reconcileWorkspace, assertMocks := buildReconcileWorkspace(workspace)

	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource .
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "awesome-workspace",
			Namespace: "terraform-system",
		},
	}
	res, err := reconcileWorkspace.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	assert.True(t, res.Requeue, "Reconcile should have been requeued")

	// Manually call Reconcile again (Requeue)
	res, err = reconcileWorkspace.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	// Make sure ConfigVersionID Status was reset
	assert.Equal(t, workspace.Status.ConfigVersionID, "", "Workspace Status ConfigVersionID should be empty")

	mock.AssertExpectationsForObjects(t, assertMocks...)
}

func buildReconcileWorkspace(workspace *v1alpha1.Workspace) (*WorkspaceHelper, []interface{}) {
	// Objects to track in the fake client.
	objs := []runtime.Object{workspace}

	gv := schema.GroupVersion{Group: "app.terraform.io", Version: "v1alpha1"}
	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(gv, workspace)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)

	// Mock the needed TFC Calls
	organizations := &mocks.Organizations{}
	organizations.On("Read", mock.Anything, "Dummy Org").Return(
		&tfe.Organization{Name: "Dummy Org"},
		nil,
	)

	workspaces := &mocks.Workspaces{}
	workspaces.On("Read", mock.Anything, mock.Anything, mock.Anything).Return(
		&tfe.Workspace{ID: "workspace-id"},
		nil,
	)

	configurationVersions := &mocks.ConfigurationVersions{}
	configurationVersions.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(
		&tfe.ConfigurationVersion{ID: "configuration-version-id"},
		nil,
	)
	configurationVersions.On("Upload", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	// First Call returns "pending"
	configurationVersions.On("Read", mock.Anything, "configuration-version-id").Return(
		&tfe.ConfigurationVersion{
			ID:     "configuration-version-id",
			Status: tfe.ConfigurationPending,
		},
		nil,
	).Once()
	// Second Call returns "uploaded"
	configurationVersions.On("Read", mock.Anything, "configuration-version-id").Return(
		&tfe.ConfigurationVersion{
			ID:     "configuration-version-id",
			Status: tfe.ConfigurationUploaded,
		},
		nil,
	)

	runs := &mocks.Runs{}
	runs.On("Create", mock.Anything, mock.Anything).Return(
		&tfe.Run{
			ID:     "awesome-run",
			Status: tfe.RunPlanning,
		},
		nil,
	).Once()

	variables := &mocks.Variables{}
	variables.On("List", mock.Anything, mock.Anything, mock.Anything).Return(&tfe.VariableList{}, nil)

	eventRecorder := &mocks.EventRecorder{}
	eventRecorder.On("Event", mock.Anything, mock.Anything, mock.Anything, mock.Anything)

	// Mocks to assert in test
	assertMocks := []interface{}{
		organizations,
		workspaces,
		configurationVersions,
		runs,
		variables,
	}

	r := &WorkspaceHelper{
		client: cl,
		scheme: s,
		tfclient: &TerraformCloudClient{
			Client: &tfe.Client{
				Organizations:         organizations,
				Workspaces:            workspaces,
				ConfigurationVersions: configurationVersions,
				Runs:                  runs,
				Variables:             variables,
			},
		},
		reqLogger: log,
		recorder:  eventRecorder,
	}

	return r, assertMocks
}
