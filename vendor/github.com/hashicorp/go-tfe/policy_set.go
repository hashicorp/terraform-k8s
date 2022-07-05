package tfe

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// Compile-time proof of interface implementation.
var _ PolicySets = (*policySets)(nil)

// PolicySets describes all the policy set related methods that the Terraform
// Enterprise API supports.
//
// TFE API docs: https://www.terraform.io/docs/cloud/api/policy-sets.html
type PolicySets interface {
	// List all the policy sets for a given organization.
	List(ctx context.Context, organization string, options *PolicySetListOptions) (*PolicySetList, error)

	// Create a policy set and associate it with an organization.
	Create(ctx context.Context, organization string, options PolicySetCreateOptions) (*PolicySet, error)

	// Read a policy set by its ID.
	Read(ctx context.Context, policySetID string) (*PolicySet, error)

	// ReadWithOptions reads a policy set by its ID using the options supplied.
	ReadWithOptions(ctx context.Context, policySetID string, options *PolicySetReadOptions) (*PolicySet, error)

	// Update an existing policy set.
	Update(ctx context.Context, policySetID string, options PolicySetUpdateOptions) (*PolicySet, error)

	// Add policies to a policy set. This function can only be used when
	// there is no VCS repository associated with the policy set.
	AddPolicies(ctx context.Context, policySetID string, options PolicySetAddPoliciesOptions) error

	// Remove policies from a policy set. This function can only be used
	// when there is no VCS repository associated with the policy set.
	RemovePolicies(ctx context.Context, policySetID string, options PolicySetRemovePoliciesOptions) error

	// Add workspaces to a policy set.
	AddWorkspaces(ctx context.Context, policySetID string, options PolicySetAddWorkspacesOptions) error

	// Remove workspaces from a policy set.
	RemoveWorkspaces(ctx context.Context, policySetID string, options PolicySetRemoveWorkspacesOptions) error

	// Delete a policy set by its ID.
	Delete(ctx context.Context, policyID string) error
}

// policySets implements PolicySets.
type policySets struct {
	client *Client
}

// PolicySetList represents a list of policy sets.
type PolicySetList struct {
	*Pagination
	Items []*PolicySet
}

// PolicySet represents a Terraform Enterprise policy set.
type PolicySet struct {
	ID             string    `jsonapi:"primary,policy-sets"`
	Name           string    `jsonapi:"attr,name"`
	Description    string    `jsonapi:"attr,description"`
	Global         bool      `jsonapi:"attr,global"`
	PoliciesPath   string    `jsonapi:"attr,policies-path"`
	PolicyCount    int       `jsonapi:"attr,policy-count"`
	VCSRepo        *VCSRepo  `jsonapi:"attr,vcs-repo"`
	WorkspaceCount int       `jsonapi:"attr,workspace-count"`
	CreatedAt      time.Time `jsonapi:"attr,created-at,iso8601"`
	UpdatedAt      time.Time `jsonapi:"attr,updated-at,iso8601"`

	// Relations
	// The organization to which the policy set belongs to.
	Organization *Organization `jsonapi:"relation,organization"`
	// The workspaces to which the policy set applies.
	Workspaces []*Workspace `jsonapi:"relation,workspaces"`
	// Individually managed policies which are associated with the policy set.
	Policies []*Policy `jsonapi:"relation,policies"`
	// The most recently created policy set version, regardless of status.
	// Note that this relationship may include an errored and unusable version,
	// and is intended to allow checking for errors.
	NewestVersion *PolicySetVersion `jsonapi:"relation,newest-version"`
	// The most recent successful policy set version.
	CurrentVersion *PolicySetVersion `jsonapi:"relation,current-version"`
}

// PolicySetIncludeOpt represents the available options for include query params.
// https://www.terraform.io/cloud-docs/api-docs/policy-sets#available-related-resources
type PolicySetIncludeOpt string

const (
	PolicySetPolicies       PolicySetIncludeOpt = "policies"
	PolicySetWorkspaces     PolicySetIncludeOpt = "workspaces"
	PolicySetCurrentVersion PolicySetIncludeOpt = "current_version"
	PolicySetNewestVersion  PolicySetIncludeOpt = "newest_version"
)

// PolicySetListOptions represents the options for listing policy sets.
type PolicySetListOptions struct {
	ListOptions

	// Optional: A search string (partial policy set name) used to filter the results.
	Search string `url:"search[name],omitempty"`
}

// PolicySetReadOptions are read options.
// For a full list of relations, please see:
// https://www.terraform.io/docs/cloud/api/policy-sets.html#relationships
type PolicySetReadOptions struct {
	// Optional: A list of relations to include. See available resources
	// https://www.terraform.io/cloud-docs/api-docs/policy-sets#available-related-resources
	Include []PolicySetIncludeOpt `url:"include,omitempty"`
}

// PolicySetCreateOptions represents the options for creating a new policy set.
type PolicySetCreateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,policy-sets"`

	// Required: The name of the policy set.
	Name *string `jsonapi:"attr,name"`

	// Optional: The description of the policy set.
	Description *string `jsonapi:"attr,description,omitempty"`

	// Optional: Whether or not the policy set is global.
	Global *bool `jsonapi:"attr,global,omitempty"`

	// Optional: The sub-path within the attached VCS repository to ingress. All
	// files and directories outside of this sub-path will be ignored.
	// This option may only be specified when a VCS repo is present.
	PoliciesPath *string `jsonapi:"attr,policies-path,omitempty"`

	// Optional: The initial members of the policy set.
	Policies []*Policy `jsonapi:"relation,policies,omitempty"`

	// Optional: VCS repository information. When present, the policies and
	// configuration will be sourced from the specified VCS repository
	// instead of being defined within the policy set itself. Note that
	// this option is mutually exclusive with the Policies option and
	// both cannot be used at the same time.
	VCSRepo *VCSRepoOptions `jsonapi:"attr,vcs-repo,omitempty"`

	// Optional: The initial list of workspaces for which the policy set should be enforced.
	Workspaces []*Workspace `jsonapi:"relation,workspaces,omitempty"`
}

// PolicySetUpdateOptions represents the options for updating a policy set.
type PolicySetUpdateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,policy-sets"`

	// Optional: The name of the policy set.
	Name *string `jsonapi:"attr,name,omitempty"`

	// Optional: The description of the policy set.
	Description *string `jsonapi:"attr,description,omitempty"`

	// Optional: Whether or not the policy set is global.
	Global *bool `jsonapi:"attr,global,omitempty"`

	// Optional: The sub-path within the attached VCS repository to ingress. All
	// files and directories outside of this sub-path will be ignored.
	// This option may only be specified when a VCS repo is present.
	PoliciesPath *string `jsonapi:"attr,policies-path,omitempty"`

	// Optional: VCS repository information. When present, the policies and
	// configuration will be sourced from the specified VCS repository
	// instead of being defined within the policy set itself. Note that
	// specifying this option may only be used on policy sets with no
	// directly-attached policies (*PolicySet.Policies). Specifying this
	// option when policies are already present will result in an error.
	VCSRepo *VCSRepoOptions `jsonapi:"attr,vcs-repo,omitempty"`
}

// PolicySetAddPoliciesOptions represents the options for adding policies
// to a policy set.
type PolicySetAddPoliciesOptions struct {
	// The policies to add to the policy set.
	Policies []*Policy
}

// PolicySetRemovePoliciesOptions represents the options for removing
// policies from a policy set.
type PolicySetRemovePoliciesOptions struct {
	// The policies to remove from the policy set.
	Policies []*Policy
}

// PolicySetAddWorkspacesOptions represents the options for adding workspaces
// to a policy set.
type PolicySetAddWorkspacesOptions struct {
	// The workspaces to add to the policy set.
	Workspaces []*Workspace
}

// PolicySetRemoveWorkspacesOptions represents the options for removing
// workspaces from a policy set.
type PolicySetRemoveWorkspacesOptions struct {
	// The workspaces to remove from the policy set.
	Workspaces []*Workspace
}

// List all the policies for a given organization.
func (s *policySets) List(ctx context.Context, organization string, options *PolicySetListOptions) (*PolicySetList, error) {
	if !validStringID(&organization) {
		return nil, ErrInvalidOrg
	}

	u := fmt.Sprintf("organizations/%s/policy-sets", url.QueryEscape(organization))
	req, err := s.client.newRequest("GET", u, options)
	if err != nil {
		return nil, err
	}

	psl := &PolicySetList{}
	err = s.client.do(ctx, req, psl)
	if err != nil {
		return nil, err
	}

	return psl, nil
}

// Create a policy set and associate it with an organization.
func (s *policySets) Create(ctx context.Context, organization string, options PolicySetCreateOptions) (*PolicySet, error) {
	if !validStringID(&organization) {
		return nil, ErrInvalidOrg
	}
	if err := options.valid(); err != nil {
		return nil, err
	}

	u := fmt.Sprintf("organizations/%s/policy-sets", url.QueryEscape(organization))
	req, err := s.client.newRequest("POST", u, &options)
	if err != nil {
		return nil, err
	}

	ps := &PolicySet{}
	err = s.client.do(ctx, req, ps)
	if err != nil {
		return nil, err
	}

	return ps, err
}

// Read a policy set by its ID.
func (s *policySets) Read(ctx context.Context, policySetID string) (*PolicySet, error) {
	return s.ReadWithOptions(ctx, policySetID, nil)
}

// ReadWithOptions reads a policy by its ID using the options supplied.
func (s *policySets) ReadWithOptions(ctx context.Context, policySetID string, options *PolicySetReadOptions) (*PolicySet, error) {
	if !validStringID(&policySetID) {
		return nil, ErrInvalidPolicySetID
	}
	if err := options.valid(); err != nil {
		return nil, err
	}

	u := fmt.Sprintf("policy-sets/%s", url.QueryEscape(policySetID))
	req, err := s.client.newRequest("GET", u, options)
	if err != nil {
		return nil, err
	}

	ps := &PolicySet{}
	err = s.client.do(ctx, req, ps)
	if err != nil {
		return nil, err
	}

	return ps, err
}

// Update an existing policy set.
func (s *policySets) Update(ctx context.Context, policySetID string, options PolicySetUpdateOptions) (*PolicySet, error) {
	if !validStringID(&policySetID) {
		return nil, ErrInvalidPolicySetID
	}
	if err := options.valid(); err != nil {
		return nil, err
	}

	u := fmt.Sprintf("policy-sets/%s", url.QueryEscape(policySetID))
	req, err := s.client.newRequest("PATCH", u, &options)
	if err != nil {
		return nil, err
	}

	ps := &PolicySet{}
	err = s.client.do(ctx, req, ps)
	if err != nil {
		return nil, err
	}

	return ps, err
}

// AddPolicies adds policies to a policy set
func (s *policySets) AddPolicies(ctx context.Context, policySetID string, options PolicySetAddPoliciesOptions) error {
	if !validStringID(&policySetID) {
		return ErrInvalidPolicySetID
	}
	if err := options.valid(); err != nil {
		return err
	}

	u := fmt.Sprintf("policy-sets/%s/relationships/policies", url.QueryEscape(policySetID))
	req, err := s.client.newRequest("POST", u, options.Policies)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}

// RemovePolicies remove policies from a policy set
func (s *policySets) RemovePolicies(ctx context.Context, policySetID string, options PolicySetRemovePoliciesOptions) error {
	if !validStringID(&policySetID) {
		return ErrInvalidPolicySetID
	}
	if err := options.valid(); err != nil {
		return err
	}

	u := fmt.Sprintf("policy-sets/%s/relationships/policies", url.QueryEscape(policySetID))
	req, err := s.client.newRequest("DELETE", u, options.Policies)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}

// Addworkspaces adds workspaces to a policy set.
func (s *policySets) AddWorkspaces(ctx context.Context, policySetID string, options PolicySetAddWorkspacesOptions) error {
	if !validStringID(&policySetID) {
		return ErrInvalidPolicySetID
	}
	if err := options.valid(); err != nil {
		return err
	}

	u := fmt.Sprintf("policy-sets/%s/relationships/workspaces", url.QueryEscape(policySetID))
	req, err := s.client.newRequest("POST", u, options.Workspaces)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}

// RemoveWorkspaces removes workspaces from a policy set.
func (s *policySets) RemoveWorkspaces(ctx context.Context, policySetID string, options PolicySetRemoveWorkspacesOptions) error {
	if !validStringID(&policySetID) {
		return ErrInvalidPolicySetID
	}
	if err := options.valid(); err != nil {
		return err
	}

	u := fmt.Sprintf("policy-sets/%s/relationships/workspaces", url.QueryEscape(policySetID))
	req, err := s.client.newRequest("DELETE", u, options.Workspaces)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}

// Delete a policy set by its ID.
func (s *policySets) Delete(ctx context.Context, policySetID string) error {
	if !validStringID(&policySetID) {
		return ErrInvalidPolicySetID
	}

	u := fmt.Sprintf("policy-sets/%s", url.QueryEscape(policySetID))
	req, err := s.client.newRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}

func (o PolicySetCreateOptions) valid() error {
	if !validString(o.Name) {
		return ErrRequiredName
	}
	if !validStringID(o.Name) {
		return ErrInvalidName
	}
	return nil
}

func (o PolicySetRemoveWorkspacesOptions) valid() error {
	if o.Workspaces == nil {
		return ErrWorkspacesRequired
	}
	if len(o.Workspaces) == 0 {
		return ErrWorkspaceMinLimit
	}
	return nil
}

func (o PolicySetUpdateOptions) valid() error {
	if o.Name != nil && !validStringID(o.Name) {
		return ErrInvalidName
	}
	return nil
}

func (o PolicySetAddPoliciesOptions) valid() error {
	if o.Policies == nil {
		return ErrRequiredPolicies
	}
	if len(o.Policies) == 0 {
		return ErrInvalidPolicies
	}
	return nil
}

func (o PolicySetRemovePoliciesOptions) valid() error {
	if o.Policies == nil {
		return ErrRequiredPolicies
	}
	if len(o.Policies) == 0 {
		return ErrInvalidPolicies
	}
	return nil
}

func (o PolicySetAddWorkspacesOptions) valid() error {
	if o.Workspaces == nil {
		return ErrWorkspacesRequired
	}
	if len(o.Workspaces) == 0 {
		return ErrWorkspaceMinLimit
	}
	return nil
}

func (o *PolicySetReadOptions) valid() error {
	if o == nil {
		return nil // nothing to validate
	}

	if err := validatePolicySetIncludeParams(o.Include); err != nil {
		return err
	}

	return nil
}

func validatePolicySetIncludeParams(params []PolicySetIncludeOpt) error {
	for _, p := range params {
		switch p {
		case PolicySetPolicies, PolicySetWorkspaces, PolicySetCurrentVersion, PolicySetNewestVersion:
			// do nothing
		default:
			return ErrInvalidIncludeValue
		}
	}

	return nil
}
