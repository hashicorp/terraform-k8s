package tfe

import (
	"context"
	"fmt"
	"net/url"
)

// Compile-time proof of interface implementation
var _ RunTasks = (*runTasks)(nil)

// RunTasks represents all the run task related methods in the context of an organization
// that the Terraform Cloud/Enterprise API supports.
// **Note: This API is still in BETA and subject to change.**
// https://www.terraform.io/cloud-docs/api-docs/run-tasks#run-tasks-api
type RunTasks interface {
	// Create a run task for an organization
	Create(ctx context.Context, organization string, options RunTaskCreateOptions) (*RunTask, error)

	// List all run tasks for an organization
	List(ctx context.Context, organization string, options *RunTaskListOptions) (*RunTaskList, error)

	// Read an organization's run task by ID
	Read(ctx context.Context, runTaskID string) (*RunTask, error)

	// Read an organization's run task by ID with given options
	ReadWithOptions(ctx context.Context, runTaskID string, options *RunTaskReadOptions) (*RunTask, error)

	// Update a run task for an organization
	Update(ctx context.Context, runTaskID string, options RunTaskUpdateOptions) (*RunTask, error)

	// Delete an organization's run task
	Delete(ctx context.Context, runTaskID string) error

	// Attach a run task to an organization's workspace
	AttachToWorkspace(ctx context.Context, workspaceID string, runTaskID string, enforcementLevel TaskEnforcementLevel) (*WorkspaceRunTask, error)
}

// runTasks implements  RunTasks
type runTasks struct {
	client *Client
}

// RunTask represents a TFC/E run task
type RunTask struct {
	ID       string  `jsonapi:"primary,tasks"`
	Name     string  `jsonapi:"attr,name"`
	URL      string  `jsonapi:"attr,url"`
	Category string  `jsonapi:"attr,category"`
	HMACKey  *string `jsonapi:"attr,hmac-key,omitempty"`
	Enabled  bool    `jsonapi:"attr,enabled"`

	Organization      *Organization       `jsonapi:"relation,organization"`
	WorkspaceRunTasks []*WorkspaceRunTask `jsonapi:"relation,workspace-tasks"`
}

// RunTaskList represents a list of run tasks
type RunTaskList struct {
	*Pagination
	Items []*RunTask
}

// RunTaskIncludeOpt represents the available options for include query params.
// https://www.terraform.io/cloud-docs/api-docs/run-tasks#list-run-tasks
type RunTaskIncludeOpt string

const (
	RunTaskWorkspaceTasks RunTaskIncludeOpt = "workspace_tasks"
	RunTaskWorkspace      RunTaskIncludeOpt = "workspace_tasks.workspace"
)

// RunTaskListOptions represents the set of options for listing run tasks
type RunTaskListOptions struct {
	ListOptions
	// Optional: A list of relations to include with a run task. See available resources:
	// https://www.terraform.io/cloud-docs/api-docs/run-tasks#list-run-tasks
	Include []RunTaskIncludeOpt `url:"include,omitempty"`
}

// RunTaskReadOptions represents the set of options for reading a run task
type RunTaskReadOptions struct {
	// Optional: A list of relations to include with a run task. See available resources:
	// https://www.terraform.io/cloud-docs/api-docs/run-tasks#list-run-tasks
	Include []RunTaskIncludeOpt `url:"include,omitempty"`
}

// RunTaskCreateOptions represents the set of options for creating a run task
type RunTaskCreateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,tasks"`

	// Required: The name of the run task
	Name string `jsonapi:"attr,name"`

	// Required: The URL to send a run task payload
	URL string `jsonapi:"attr,url"`

	// Required: Must be "task"
	Category string `jsonapi:"attr,category"`

	// Optional: An HMAC key to verify the run task
	HMACKey *string `jsonapi:"attr,hmac-key,omitempty"`

	// Optional: Whether the task should be enabled
	Enabled *bool `jsonapi:"attr,enabled,omitempty"`
}

// RunTaskUpdateOptions represents the set of options for updating an organization's run task
type RunTaskUpdateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,tasks"`

	// Optional: The name of the run task, defaults to previous value
	Name *string `jsonapi:"attr,name,omitempty"`

	// Optional: The URL to send a run task payload, defaults to previous value
	URL *string `jsonapi:"attr,url,omitempty"`

	// Optional: Must be "task", defaults to "task"
	Category *string `jsonapi:"attr,category,omitempty"`

	// Optional: An HMAC key to verify the run task
	HMACKey *string `jsonapi:"attr,hmac-key,omitempty"`

	// Optional: Whether the task should be enabled
	Enabled *bool `jsonapi:"attr,enabled,omitempty"`
}

// Create is used to create a new run task for an organization
func (s *runTasks) Create(ctx context.Context, organization string, options RunTaskCreateOptions) (*RunTask, error) {
	if !validStringID(&organization) {
		return nil, ErrInvalidOrg
	}

	if err := options.valid(); err != nil {
		return nil, err
	}

	u := fmt.Sprintf("organizations/%s/tasks", url.QueryEscape(organization))
	req, err := s.client.newRequest("POST", u, &options)
	if err != nil {
		return nil, err
	}

	r := &RunTask{}
	err = s.client.do(ctx, req, r)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// List all the run tasks for an organization
func (s *runTasks) List(ctx context.Context, organization string, options *RunTaskListOptions) (*RunTaskList, error) {
	if !validStringID(&organization) {
		return nil, ErrInvalidOrg
	}
	if err := options.valid(); err != nil {
		return nil, err
	}

	u := fmt.Sprintf("organizations/%s/tasks", url.QueryEscape(organization))
	req, err := s.client.newRequest("GET", u, options)
	if err != nil {
		return nil, err
	}

	rl := &RunTaskList{}
	err = s.client.do(ctx, req, rl)
	if err != nil {
		return nil, err
	}

	return rl, nil
}

// Read is used to read an organization's run task by ID
func (s *runTasks) Read(ctx context.Context, runTaskID string) (*RunTask, error) {
	return s.ReadWithOptions(ctx, runTaskID, nil)
}

// Read is used to read an organization's run task by ID with options
func (s *runTasks) ReadWithOptions(ctx context.Context, runTaskID string, options *RunTaskReadOptions) (*RunTask, error) {
	if !validStringID(&runTaskID) {
		return nil, ErrInvalidRunTaskID
	}
	if err := options.valid(); err != nil {
		return nil, err
	}

	u := fmt.Sprintf("tasks/%s", url.QueryEscape(runTaskID))
	req, err := s.client.newRequest("GET", u, options)
	if err != nil {
		return nil, err
	}

	r := &RunTask{}
	err = s.client.do(ctx, req, r)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// Update an existing run task for an organization by ID
func (s *runTasks) Update(ctx context.Context, runTaskID string, options RunTaskUpdateOptions) (*RunTask, error) {
	if !validStringID(&runTaskID) {
		return nil, ErrInvalidRunTaskID
	}

	if err := options.valid(); err != nil {
		return nil, err
	}

	u := fmt.Sprintf("tasks/%s", url.QueryEscape(runTaskID))
	req, err := s.client.newRequest("PATCH", u, &options)
	if err != nil {
		return nil, err
	}

	r := &RunTask{}
	err = s.client.do(ctx, req, r)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// Delete an existing run task for an organization by ID
func (s *runTasks) Delete(ctx context.Context, runTaskID string) error {
	if !validStringID(&runTaskID) {
		return ErrInvalidRunTaskID
	}

	u := fmt.Sprintf("tasks/%s", runTaskID)
	req, err := s.client.newRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}

// AttachToWorkspace is a convenient method to attach a run task to a workspace. See: WorkspaceRunTasks.Create()
func (s *runTasks) AttachToWorkspace(ctx context.Context, workspaceID, runTaskID string, enforcement TaskEnforcementLevel) (*WorkspaceRunTask, error) {
	return s.client.WorkspaceRunTasks.Create(ctx, workspaceID, WorkspaceRunTaskCreateOptions{
		EnforcementLevel: enforcement,
		RunTask:          &RunTask{ID: runTaskID},
	})
}

func (o *RunTaskCreateOptions) valid() error {
	if !validString(&o.Name) {
		return ErrRequiredName
	}

	if !validString(&o.URL) {
		return ErrInvalidRunTaskURL
	}

	if o.Category != "task" {
		return ErrInvalidRunTaskCategory
	}

	return nil
}

func (o *RunTaskUpdateOptions) valid() error {
	if o.Name != nil && !validString(o.Name) {
		return ErrRequiredName
	}

	if o.URL != nil && !validString(o.URL) {
		return ErrInvalidRunTaskURL
	}

	if o.Category != nil && *o.Category != "task" {
		return ErrInvalidRunTaskCategory
	}

	return nil
}

func (o *RunTaskListOptions) valid() error {
	if o == nil {
		return nil // nothing to validate
	}

	if err := validateRunTaskIncludeParams(o.Include); err != nil {
		return err
	}

	return nil
}

func (o *RunTaskReadOptions) valid() error {
	if o == nil {
		return nil // nothing to validate
	}

	if err := validateRunTaskIncludeParams(o.Include); err != nil {
		return err
	}

	return nil
}

func validateRunTaskIncludeParams(params []RunTaskIncludeOpt) error {
	for _, p := range params {
		switch p {
		case RunTaskWorkspaceTasks, RunTaskWorkspace:
			// do nothing
		default:
			return ErrInvalidIncludeValue
		}
	}

	return nil
}
