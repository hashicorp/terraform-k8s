// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfe

import (
	"context"
	"fmt"
	"net/url"
)

// Compile-time proof of interface implementation.
var _ OrganizationMemberships = (*organizationMemberships)(nil)

// OrganizationMemberships describes all the organization membership related methods that
// the Terraform Enterprise API supports.
//
// TFE API docs:
// https://www.terraform.io/docs/cloud/api/organization-memberships.html
type OrganizationMemberships interface {
	// List all the organization memberships of the given organization.
	List(ctx context.Context, organization string, options *OrganizationMembershipListOptions) (*OrganizationMembershipList, error)

	// Create a new organization membership with the given options.
	Create(ctx context.Context, organization string, options OrganizationMembershipCreateOptions) (*OrganizationMembership, error)

	// Read an organization membership by ID
	Read(ctx context.Context, organizationMembershipID string) (*OrganizationMembership, error)

	// Read an organization membership by ID with options
	ReadWithOptions(ctx context.Context, organizationMembershipID string, options OrganizationMembershipReadOptions) (*OrganizationMembership, error)

	// Delete an organization membership by its ID.
	Delete(ctx context.Context, organizationMembershipID string) error
}

// organizationMemberships implements OrganizationMemberships.
type organizationMemberships struct {
	client *Client
}

// OrganizationMembershipStatus represents an organization membership status.
type OrganizationMembershipStatus string

const (
	OrganizationMembershipActive  OrganizationMembershipStatus = "active"
	OrganizationMembershipInvited OrganizationMembershipStatus = "invited"
)

// OrganizationMembershipList represents a list of organization memberships.
type OrganizationMembershipList struct {
	*Pagination
	Items []*OrganizationMembership
}

// OrganizationMembership represents a Terraform Enterprise organization membership.
type OrganizationMembership struct {
	ID     string                       `jsonapi:"primary,organization-memberships"`
	Status OrganizationMembershipStatus `jsonapi:"attr,status"`
	Email  string                       `jsonapi:"attr,email"`

	// Relations
	Organization *Organization `jsonapi:"relation,organization"`
	User         *User         `jsonapi:"relation,user"`
	Teams        []*Team       `jsonapi:"relation,teams"`
}

// OrgMembershipIncludeOpt represents the available options for include query params.
// https://www.terraform.io/cloud-docs/api-docs/organization-memberships#available-related-resources
type OrgMembershipIncludeOpt string

const (
	OrgMembershipUser OrgMembershipIncludeOpt = "user"
	OrgMembershipTeam OrgMembershipIncludeOpt = "teams"
)

// OrganizationMembershipListOptions represents the options for listing organization memberships.
type OrganizationMembershipListOptions struct {
	ListOptions
	// Optional: A list of relations to include. See available resources
	// https://www.terraform.io/cloud-docs/api-docs/organization-memberships#available-related-resources
	Include []OrgMembershipIncludeOpt `url:"include,omitempty"`

	// Optional: A list of organization member emails to filter by.
	Emails []string `url:"filter[email],omitempty"`
}

// OrganizationMembershipCreateOptions represents the options for creating an organization membership.
type OrganizationMembershipCreateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,organization-memberships"`

	// Required: User's email address.
	Email *string `jsonapi:"attr,email"`
}

// OrganizationMembershipReadOptions represents the options for reading organization memberships.
type OrganizationMembershipReadOptions struct {
	// Optional: A list of relations to include. See available resources
	// https://www.terraform.io/cloud-docs/api-docs/organization-memberships#available-related-resources
	Include []OrgMembershipIncludeOpt `url:"include,omitempty"`
}

// List all the organization memberships of the given organization.
func (s *organizationMemberships) List(ctx context.Context, organization string, options *OrganizationMembershipListOptions) (*OrganizationMembershipList, error) {
	if !validStringID(&organization) {
		return nil, ErrInvalidOrg
	}
	if err := options.valid(); err != nil {
		return nil, err
	}

	u := fmt.Sprintf("organizations/%s/organization-memberships", url.QueryEscape(organization))
	req, err := s.client.newRequest("GET", u, options)
	if err != nil {
		return nil, err
	}

	ml := &OrganizationMembershipList{}
	err = s.client.do(ctx, req, ml)
	if err != nil {
		return nil, err
	}

	return ml, nil
}

// Create an organization membership with the given options.
func (s *organizationMemberships) Create(ctx context.Context, organization string, options OrganizationMembershipCreateOptions) (*OrganizationMembership, error) {
	if !validStringID(&organization) {
		return nil, ErrInvalidOrg
	}
	if err := options.valid(); err != nil {
		return nil, err
	}

	u := fmt.Sprintf("organizations/%s/organization-memberships", url.QueryEscape(organization))
	req, err := s.client.newRequest("POST", u, &options)
	if err != nil {
		return nil, err
	}

	m := &OrganizationMembership{}
	err = s.client.do(ctx, req, m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// Read an organization membership by its ID.
func (s *organizationMemberships) Read(ctx context.Context, organizationMembershipID string) (*OrganizationMembership, error) {
	return s.ReadWithOptions(ctx, organizationMembershipID, OrganizationMembershipReadOptions{})
}

// Read an organization membership by ID with options
func (s *organizationMemberships) ReadWithOptions(ctx context.Context, organizationMembershipID string, options OrganizationMembershipReadOptions) (*OrganizationMembership, error) {
	if !validStringID(&organizationMembershipID) {
		return nil, ErrInvalidMembership
	}
	if err := options.valid(); err != nil {
		return nil, err
	}

	u := fmt.Sprintf("organization-memberships/%s", url.QueryEscape(organizationMembershipID))
	req, err := s.client.newRequest("GET", u, &options)
	if err != nil {
		return nil, err
	}

	mem := &OrganizationMembership{}
	err = s.client.do(ctx, req, mem)
	if err != nil {
		return nil, err
	}

	return mem, nil
}

// Delete an organization membership by its ID.
func (s *organizationMemberships) Delete(ctx context.Context, organizationMembershipID string) error {
	if !validStringID(&organizationMembershipID) {
		return ErrInvalidMembership
	}

	u := fmt.Sprintf("organization-memberships/%s", url.QueryEscape(organizationMembershipID))
	req, err := s.client.newRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}

func (o OrganizationMembershipCreateOptions) valid() error {
	if o.Email == nil {
		return ErrRequiredEmail
	}
	return nil
}

func (o *OrganizationMembershipListOptions) valid() error {
	if o == nil {
		return nil
	}

	if err := validateOrgMembershipIncludeParams(o.Include); err != nil {
		return err
	}

	if err := validateOrgMembershipEmailParams(o.Emails); err != nil {
		return err
	}

	return nil
}

func (o OrganizationMembershipReadOptions) valid() error {
	if err := validateOrgMembershipIncludeParams(o.Include); err != nil {
		return err
	}

	return nil
}

func validateOrgMembershipIncludeParams(params []OrgMembershipIncludeOpt) error {
	for _, p := range params {
		switch p {
		case OrgMembershipUser, OrgMembershipTeam:
			// do nothing
		default:
			return ErrInvalidIncludeValue
		}
	}

	return nil
}

func validateOrgMembershipEmailParams(emails []string) error {
	for _, email := range emails {
		if !validEmail(email) {
			return ErrInvalidEmail
		}
	}

	return nil
}
