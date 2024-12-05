package tfe

import (
	"context"
	"fmt"
	"net/url"
)

// Compile-time proof of interface implementation
var _ WorkspaceRunTasks = (*workspaceRunTasks)(nil)

// WorkspaceRunTasks represent all the run task related methods in the context of a workspace that the Terraform Cloud/Enterprise API supports.
// **Note: This API is still in BETA and subject to change.**
type WorkspaceRunTasks interface {
	// Add a run task to a workspace
	Create(ctx context.Context, workspaceID string, options WorkspaceRunTaskCreateOptions) (*WorkspaceRunTask, error)

	// List all run tasks for a workspace
	List(ctx context.Context, workspaceID string, options *WorkspaceRunTaskListOptions) (*WorkspaceRunTaskList, error)

	// Read a workspace run task by ID
	Read(ctx context.Context, workspaceID string, workspaceTaskID string) (*WorkspaceRunTask, error)

	// Update a workspace run task by ID
	Update(ctx context.Context, workspaceID string, workspaceTaskID string, options WorkspaceRunTaskUpdateOptions) (*WorkspaceRunTask, error)

	// Delete a workspace's run task by ID
	Delete(ctx context.Context, workspaceID string, workspaceTaskID string) error
}

// workspaceRunTasks implements WorkspaceRunTasks
type workspaceRunTasks struct {
	client *Client
}

// WorkspaceRunTask represents a TFC/E run task that belongs to a workspace
type WorkspaceRunTask struct {
	ID               string               `jsonapi:"primary,workspace-tasks"`
	EnforcementLevel TaskEnforcementLevel `jsonapi:"attr,enforcement-level"`

	RunTask   *RunTask   `jsonapi:"relation,task"`
	Workspace *Workspace `jsonapi:"relation,workspace"`
}

// WorkspaceRunTaskList represents a list of workspace run tasks
type WorkspaceRunTaskList struct {
	*Pagination
	Items []*WorkspaceRunTask
}

// WorkspaceRunTaskListOptions represents the set of options for listing workspace run tasks
type WorkspaceRunTaskListOptions struct {
	ListOptions
}

// WorkspaceRunTaskCreateOptions represents the set of options for creating a workspace run task
type WorkspaceRunTaskCreateOptions struct {
	Type string `jsonapi:"primary,workspace-tasks"`
	// Required: The enforcement level for a run task
	EnforcementLevel TaskEnforcementLevel `jsonapi:"attr,enforcement-level"`
	// Required: The run task to attach to the workspace
	RunTask *RunTask `jsonapi:"relation,task"`
}

// WorkspaceRunTaskUpdateOptions represent the set of options for updating a workspace run task.
type WorkspaceRunTaskUpdateOptions struct {
	Type             string               `jsonapi:"primary,workspace-tasks"`
	EnforcementLevel TaskEnforcementLevel `jsonapi:"attr,enforcement-level,omitempty"`
}

// List all run tasks attached to a workspace
func (s *workspaceRunTasks) List(ctx context.Context, workspaceID string, options *WorkspaceRunTaskListOptions) (*WorkspaceRunTaskList, error) {
	if !validStringID(&workspaceID) {
		return nil, ErrInvalidWorkspaceID
	}

	u := fmt.Sprintf("workspaces/%s/tasks", url.QueryEscape(workspaceID))
	req, err := s.client.newRequest("GET", u, options)
	if err != nil {
		return nil, err
	}

	rl := &WorkspaceRunTaskList{}
	err = s.client.do(ctx, req, rl)
	if err != nil {
		return nil, err
	}

	return rl, nil
}

// Read a workspace run task by ID
func (s *workspaceRunTasks) Read(ctx context.Context, workspaceID, workspaceTaskID string) (*WorkspaceRunTask, error) {
	if !validStringID(&workspaceID) {
		return nil, ErrInvalidWorkspaceID
	}

	if !validStringID(&workspaceTaskID) {
		return nil, ErrInvalidWorkspaceRunTaskID
	}

	u := fmt.Sprintf(
		"workspaces/%s/tasks/%s",
		url.QueryEscape(workspaceID),
		url.QueryEscape(workspaceTaskID),
	)
	req, err := s.client.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	wr := &WorkspaceRunTask{}
	err = s.client.do(ctx, req, wr)
	if err != nil {
		return nil, err
	}

	return wr, nil
}

// Create is used to attach a run task to a workspace, or in other words: create a workspace run task. The run task must exist in the workspace's organization.
func (s *workspaceRunTasks) Create(ctx context.Context, workspaceID string, options WorkspaceRunTaskCreateOptions) (*WorkspaceRunTask, error) {
	if !validStringID(&workspaceID) {
		return nil, ErrInvalidWorkspaceID
	}

	if err := options.valid(); err != nil {
		return nil, err
	}

	u := fmt.Sprintf("workspaces/%s/tasks", workspaceID)
	req, err := s.client.newRequest("POST", u, &options)
	if err != nil {
		return nil, err
	}

	wr := &WorkspaceRunTask{}
	err = s.client.do(ctx, req, wr)
	if err != nil {
		return nil, err
	}

	return wr, nil
}

// Update an existing workspace run task by ID
func (s *workspaceRunTasks) Update(ctx context.Context, workspaceID, workspaceTaskID string, options WorkspaceRunTaskUpdateOptions) (*WorkspaceRunTask, error) {
	if !validStringID(&workspaceID) {
		return nil, ErrInvalidWorkspaceID
	}

	if !validStringID(&workspaceTaskID) {
		return nil, ErrInvalidWorkspaceRunTaskID
	}

	u := fmt.Sprintf(
		"workspaces/%s/tasks/%s",
		url.QueryEscape(workspaceID),
		url.QueryEscape(workspaceTaskID),
	)
	req, err := s.client.newRequest("PATCH", u, &options)
	if err != nil {
		return nil, err
	}

	wr := &WorkspaceRunTask{}
	err = s.client.do(ctx, req, wr)
	if err != nil {
		return nil, err
	}

	return wr, nil
}

// Delete a workspace run task by ID
func (s *workspaceRunTasks) Delete(ctx context.Context, workspaceID, workspaceTaskID string) error {
	if !validStringID(&workspaceID) {
		return ErrInvalidWorkspaceID
	}

	if !validStringID(&workspaceTaskID) {
		return ErrInvalidWorkspaceRunTaskType
	}

	u := fmt.Sprintf(
		"workspaces/%s/tasks/%s",
		url.QueryEscape(workspaceID),
		url.QueryEscape(workspaceTaskID),
	)
	req, err := s.client.newRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}

func (o *WorkspaceRunTaskCreateOptions) valid() error {
	if o.RunTask.ID == "" {
		return ErrInvalidRunTaskID
	}

	return nil
}
