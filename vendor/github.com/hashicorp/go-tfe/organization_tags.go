package tfe

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

var _ OrganizationTags = (*organizationTags)(nil)

type OrganizationTags interface {
	// List all tags within an organization
	List(ctx context.Context, organization string, options OrganizationTagsListOptions) (*OrganizationTagsList, error)

	// Delete tags from an organization
	Delete(ctx context.Context, organization string, options OrganizationTagsDeleteOptions) error

	// Associate an organization's workspace with a tag
	AddWorkspaces(ctx context.Context, tag string, options AddWorkspacesToTagOptions) error
}

type organizationTags struct {
	client *Client
}

// OrganizationTagsList represents a list of organization tags
type OrganizationTagsList struct {
	*Pagination
	Items []*OrganizationTag
}

// OrganizationTag represents a Terraform Enterprise Organization tag
type OrganizationTag struct {
	ID   string `jsonapi:"primary,tags"`
	Name string `jsonapi:"attr,name,omitempty"`

	// Number of workspaces that have this tag
	InstanceCount int `jsonapi:"attr,instance-count,omitempty"`

	// The org this tag belongs to
	Organization *Organization `jsonapi:"relation,organization"`
}

// OrganizationTagsListOptions represents the options for listing organization tags
type OrganizationTagsListOptions struct {
	ListOptions

	Filter *string `url:"filter[exclude][taggable][id],omitempty"`
}

// List all the tags in an organization. You can provide query params through OrganizationTagsListOptions
func (s *organizationTags) List(ctx context.Context, organization string, options OrganizationTagsListOptions) (*OrganizationTagsList, error) {
	if !validStringID(&organization) {
		return nil, ErrInvalidOrg
	}

	u := fmt.Sprintf("organizations/%s/tags", url.QueryEscape(organization))
	req, err := s.client.newRequest("GET", u, &options)
	if err != nil {
		return nil, err
	}

	tags := &OrganizationTagsList{}
	err = s.client.do(ctx, req, tags)
	if err != nil {
		return nil, err
	}

	return tags, nil
}

// OrganizationTagsDeleteOptions represents the request body for deleting a tag in an organization
type OrganizationTagsDeleteOptions struct {
	IDs []string
}

// this represents a single tag ID sent over the wire
type tagID struct {
	ID string `jsonapi:"primary,tags"`
}

func (opts *OrganizationTagsDeleteOptions) valid() error {
	if opts.IDs == nil || len(opts.IDs) == 0 {
		return errors.New("you must specify at least one tag id to remove")
	}

	for _, id := range opts.IDs {
		if !validStringID(&id) {
			errorMsg := fmt.Sprintf("%s is not a valid id value", id)
			return errors.New(errorMsg)
		}
	}

	return nil
}

// Delete tags from a Terraform Enterprise organization
func (s *organizationTags) Delete(ctx context.Context, organization string, options OrganizationTagsDeleteOptions) error {
	if !validStringID(&organization) {
		return ErrInvalidOrg
	}

	if err := options.valid(); err != nil {
		return err
	}

	u := fmt.Sprintf("organizations/%s/tags", url.QueryEscape(organization))
	var tagsToRemove []*tagID
	for _, id := range options.IDs {
		tagsToRemove = append(tagsToRemove, &tagID{ID: id})
	}

	req, err := s.client.newRequest("DELETE", u, tagsToRemove)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}

// AddWorkspacesToTagOptions represents the request body to add a workspace to a tag
type AddWorkspacesToTagOptions struct {
	WorkspaceIDs []string
}

func (w *AddWorkspacesToTagOptions) valid() error {
	if w.WorkspaceIDs == nil || len(w.WorkspaceIDs) == 0 {
		return errors.New("you must specify at least one workspace to add tag to")
	}

	for _, id := range w.WorkspaceIDs {
		if !validStringID(&id) {
			errorMsg := fmt.Sprintf("%s is not a valid id value", id)
			return errors.New(errorMsg)
		}
	}

	return nil
}

// this represents how workspace IDs will be sent over the wire
type workspaceID struct {
	ID string `jsonapi:"primary,workspaces"`
}

// Add workspaces to a tag
func (s *organizationTags) AddWorkspaces(ctx context.Context, tag string, options AddWorkspacesToTagOptions) error {
	if !validStringID(&tag) {
		return errors.New("invalid tag id")
	}

	if err := options.valid(); err != nil {
		return err
	}

	var workspaces []*workspaceID
	for _, id := range options.WorkspaceIDs {
		workspaces = append(workspaces, &workspaceID{ID: id})
	}

	u := fmt.Sprintf("tags/%s/relationships/workspaces", url.QueryEscape(tag))
	req, err := s.client.newRequest("POST", u, workspaces)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}
