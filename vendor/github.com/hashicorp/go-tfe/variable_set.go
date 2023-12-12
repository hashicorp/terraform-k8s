package tfe

import (
	"context"
	"fmt"
	"net/url"
)

// Compile-time proof of interface implementation.
var _ VariableSets = (*variableSets)(nil)

// VariableSets describes all the Variable Set related methods that the
// Terraform Enterprise API supports.
//
// TFE API docs: https://www.terraform.io/cloud-docs/api-docs/variable-sets
type VariableSets interface {
	// List all the variable sets within an organization.
	List(ctx context.Context, organization string, options *VariableSetListOptions) (*VariableSetList, error)

	// Create is used to create a new variable set.
	Create(ctx context.Context, organization string, options *VariableSetCreateOptions) (*VariableSet, error)

	// Read a variable set by its ID.
	Read(ctx context.Context, variableSetID string, options *VariableSetReadOptions) (*VariableSet, error)

	// Update an existing variable set.
	Update(ctx context.Context, variableSetID string, options *VariableSetUpdateOptions) (*VariableSet, error)

	// Delete a variable set by ID.
	Delete(ctx context.Context, variableSetID string) error

	// Apply variable set to workspaces in the supplied list.
	ApplyToWorkspaces(ctx context.Context, variableSetID string, options *VariableSetApplyToWorkspacesOptions) error

	// Remove variable set from workspaces in the supplied list.
	RemoveFromWorkspaces(ctx context.Context, variableSetID string, options *VariableSetRemoveFromWorkspacesOptions) error

	// Update list of workspaces to which the variable set is applied to match the supplied list.
	UpdateWorkspaces(ctx context.Context, variableSetID string, options *VariableSetUpdateWorkspacesOptions) (*VariableSet, error)
}

// variableSets implements VariableSets.
type variableSets struct {
	client *Client
}

// VariableSetList represents a list of variable sets.
type VariableSetList struct {
	*Pagination
	Items []*VariableSet
}

// VariableSet represents a Terraform Enterprise variable set.
type VariableSet struct {
	ID          string `jsonapi:"primary,varsets"`
	Name        string `jsonapi:"attr,name"`
	Description string `jsonapi:"attr,description"`
	Global      bool   `jsonapi:"attr,global"`

	// Relations
	Organization *Organization          `jsonapi:"relation,organization"`
	Workspaces   []*Workspace           `jsonapi:"relation,workspaces,omitempty"`
	Variables    []*VariableSetVariable `jsonapi:"relation,vars,omitempty"`
}

// A list of relations to include. See available resources
// https://www.terraform.io/docs/cloud/api/admin/organizations.html#available-related-resources
type VariableSetIncludeOpt string

const (
	VariableSetWorkspaces VariableSetIncludeOpt = "workspaces"
	VariableSetVars       VariableSetIncludeOpt = "vars"
)

// VariableSetListOptions represents the options for listing variable sets.
type VariableSetListOptions struct {
	ListOptions
	Include string `url:"include"`
}

// VariableSetCreateOptions represents the options for creating a new variable set within in a organization.
type VariableSetCreateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,varsets"`

	// The name of the variable set.
	// Affects variable precedence when there are conflicts between Variable Sets
	// https://www.terraform.io/cloud-docs/api-docs/variable-sets#apply-variable-set-to-workspaces
	Name *string `jsonapi:"attr,name"`

	// A description to provide context for the variable set.
	Description *string `jsonapi:"attr,description,omitempty"`

	// If true the variable set is considered in all runs in the organization.
	Global *bool `jsonapi:"attr,global,omitempty"`
}

// VariableSetReadOptions represents the options for reading variable sets.
type VariableSetReadOptions struct {
	Include *[]VariableSetIncludeOpt `url:"include:omitempty"`
}

// VariableSetUpdateOptions represents the options for updating a variable set.
type VariableSetUpdateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,varsets"`

	// The name of the variable set.
	// Affects variable precedence when there are conflicts between Variable Sets
	// https://www.terraform.io/cloud-docs/api-docs/variable-sets#apply-variable-set-to-workspaces
	Name *string `jsonapi:"attr,name,omitempty"`

	// A description to provide context for the variable set.
	Description *string `jsonapi:"attr,description,omitempty"`

	// If true the variable set is considered in all runs in the organization.
	Global *bool `jsonapi:"attr,global,omitempty"`
}

// VariableSetApplyToWorkspacesOptions represents the options for applying variable sets to workspaces.
type VariableSetApplyToWorkspacesOptions struct {
	// The workspaces to apply the variable set to (additive).
	Workspaces []*Workspace
}

// VariableSetRemoveFromWorkspacesOptions represents the options for removing variable sets from workspaces.
type VariableSetRemoveFromWorkspacesOptions struct {
	// The workspaces to remove the variable set from.
	Workspaces []*Workspace
}

// VariableSetUpdateWorkspacesOptions represents a subset of update options specifically for applying variable sets to workspaces
type VariableSetUpdateWorkspacesOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,varsets"`

	// The workspaces to be applied to. An empty set means remove all applied
	Workspaces []*Workspace `jsonapi:"relation,workspaces"`
}

type privateVariableSetUpdateWorkspacesOptions struct {
	Type       string       `jsonapi:"primary,varsets"`
	Global     bool         `jsonapi:"attr,global"`
	Workspaces []*Workspace `jsonapi:"relation,workspaces"`
}

// List all Variable Sets in the organization
func (s *variableSets) List(ctx context.Context, organization string, options *VariableSetListOptions) (*VariableSetList, error) {
	if !validStringID(&organization) {
		return nil, ErrInvalidOrg
	}
	if options != nil {
		if err := options.valid(); err != nil {
			return nil, err
		}
	}

	u := fmt.Sprintf("organizations/%s/varsets", url.QueryEscape(organization))
	req, err := s.client.newRequest("GET", u, options)
	if err != nil {
		return nil, err
	}

	vl := &VariableSetList{}
	err = s.client.do(ctx, req, vl)
	if err != nil {
		return nil, err
	}

	return vl, nil
}

// Create is used to create a new variable set.
func (s *variableSets) Create(ctx context.Context, organization string, options *VariableSetCreateOptions) (*VariableSet, error) {
	if !validStringID(&organization) {
		return nil, ErrInvalidOrg
	}
	if err := options.valid(); err != nil {
		return nil, err
	}

	u := fmt.Sprintf("organizations/%s/varsets", url.QueryEscape(organization))
	req, err := s.client.newRequest("POST", u, options)
	if err != nil {
		return nil, err
	}

	vl := &VariableSet{}
	err = s.client.do(ctx, req, vl)
	if err != nil {
		return nil, err
	}

	return vl, nil
}

// Read is used to inspect a given variable set based on ID
func (s *variableSets) Read(ctx context.Context, variableSetID string, options *VariableSetReadOptions) (*VariableSet, error) {
	if !validStringID(&variableSetID) {
		return nil, ErrInvalidVariableSetID
	}

	u := fmt.Sprintf("varsets/%s", url.QueryEscape(variableSetID))
	req, err := s.client.newRequest("GET", u, options)
	if err != nil {
		return nil, err
	}

	vs := &VariableSet{}
	err = s.client.do(ctx, req, vs)
	if err != nil {
		return nil, err
	}

	return vs, err
}

// Update an existing variable set.
func (s *variableSets) Update(ctx context.Context, variableSetID string, options *VariableSetUpdateOptions) (*VariableSet, error) {
	if !validStringID(&variableSetID) {
		return nil, ErrInvalidVariableSetID
	}

	u := fmt.Sprintf("varsets/%s", url.QueryEscape(variableSetID))
	req, err := s.client.newRequest("PATCH", u, options)
	if err != nil {
		return nil, err
	}

	v := &VariableSet{}
	err = s.client.do(ctx, req, v)
	if err != nil {
		return nil, err
	}

	return v, nil
}

// Delete a variable set by its ID.
func (s *variableSets) Delete(ctx context.Context, variableSetID string) error {
	if !validStringID(&variableSetID) {
		return ErrInvalidVariableSetID
	}

	u := fmt.Sprintf("varsets/%s", url.QueryEscape(variableSetID))
	req, err := s.client.newRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}

// Apply variable set to workspaces in the supplied list.
// Note: this method will return an error if the variable set has global = true.
func (s *variableSets) ApplyToWorkspaces(ctx context.Context, variableSetID string, options *VariableSetApplyToWorkspacesOptions) error {
	if !validStringID(&variableSetID) {
		return ErrInvalidVariableSetID
	}
	if err := options.valid(); err != nil {
		return err
	}

	u := fmt.Sprintf("varsets/%s/relationships/workspaces", url.QueryEscape(variableSetID))
	req, err := s.client.newRequest("POST", u, options.Workspaces)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}

// Remove variable set from workspaces in the supplied list.
// Note: this method will return an error if the variable set has global = true.
func (s *variableSets) RemoveFromWorkspaces(ctx context.Context, variableSetID string, options *VariableSetRemoveFromWorkspacesOptions) error {
	if !validStringID(&variableSetID) {
		return ErrInvalidVariableSetID
	}
	if err := options.valid(); err != nil {
		return err
	}

	u := fmt.Sprintf("varsets/%s/relationships/workspaces", url.QueryEscape(variableSetID))
	req, err := s.client.newRequest("DELETE", u, options.Workspaces)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}

// Update variable set to be applied to only the workspaces in the supplied list.
func (s *variableSets) UpdateWorkspaces(ctx context.Context, variableSetID string, options *VariableSetUpdateWorkspacesOptions) (*VariableSet, error) {
	if err := options.valid(); err != nil {
		return nil, err
	}

	// Use private struct to ensure global is set to false when applying to workspaces
	o := privateVariableSetUpdateWorkspacesOptions{
		Global:     bool(false),
		Workspaces: options.Workspaces,
	}

	// We force inclusion of workspaces as that is the primary data for which we are concerned with confirming changes.
	u := fmt.Sprintf("varsets/%s?include=%s", url.QueryEscape(variableSetID), VariableSetWorkspaces)
	req, err := s.client.newRequest("PATCH", u, &o)
	if err != nil {
		return nil, err
	}

	v := &VariableSet{}
	err = s.client.do(ctx, req, v)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (o *VariableSetListOptions) valid() error {
	return nil
}

func (o *VariableSetCreateOptions) valid() error {
	if o == nil {
		return nil
	}
	if !validString(o.Name) {
		return ErrRequiredName
	}
	if o.Global == nil {
		return ErrRequiredGlobalFlag
	}
	return nil
}

func (o *VariableSetApplyToWorkspacesOptions) valid() error {
	for _, s := range o.Workspaces {
		if !validStringID(&s.ID) {
			return ErrRequiredWorkspaceID
		}
	}
	return nil
}

func (o *VariableSetRemoveFromWorkspacesOptions) valid() error {
	for _, s := range o.Workspaces {
		if !validStringID(&s.ID) {
			return ErrRequiredWorkspaceID
		}
	}
	return nil
}

func (o *VariableSetUpdateWorkspacesOptions) valid() error {
	if o == nil || o.Workspaces == nil {
		return ErrRequiredWorkspacesList
	}
	return nil
}
