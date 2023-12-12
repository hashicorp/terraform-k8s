// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfe

import (
	"context"
	"fmt"
	"net/url"
)

// Compile-time proof of interface implementation.
var _ AdminOrganizations = (*adminOrganizations)(nil)

// AdminOrganizations describes all of the admin organization related methods that the Terraform
// Enterprise API supports. Note that admin settings are only available in Terraform Enterprise.
//
// TFE API docs: https://www.terraform.io/docs/cloud/api/admin/organizations.html
type AdminOrganizations interface {
	// List all the organizations visible to the current user.
	List(ctx context.Context, options *AdminOrganizationListOptions) (*AdminOrganizationList, error)

	// Read attributes of an existing organization via admin API.
	Read(ctx context.Context, organization string) (*AdminOrganization, error)

	// Update attributes of an existing organization via admin API.
	Update(ctx context.Context, organization string, options AdminOrganizationUpdateOptions) (*AdminOrganization, error)

	// Delete an organization by its name via admin API
	Delete(ctx context.Context, organization string) error

	// ListModuleConsumers lists specific organizations in the Terraform Enterprise installation that have permission to use an organization's modules.
	ListModuleConsumers(ctx context.Context, organization string, options *AdminOrganizationListModuleConsumersOptions) (*AdminOrganizationList, error)

	// UpdateModuleConsumers specifies a list of organizations that can use modules from the sharing organization's private registry. Setting a list of module consumers will turn off global module sharing for an organization.
	UpdateModuleConsumers(ctx context.Context, organization string, consumerOrganizations []string) error
}

// adminOrganizations implements AdminOrganizations.
type adminOrganizations struct {
	client *Client
}

// AdminOrganization represents a Terraform Enterprise organization returned from the Admin API.
type AdminOrganization struct {
	Name                             string `jsonapi:"primary,organizations"`
	AccessBetaTools                  bool   `jsonapi:"attr,access-beta-tools"`
	ExternalID                       string `jsonapi:"attr,external-id"`
	GlobalModuleSharing              *bool  `jsonapi:"attr,global-module-sharing"`
	IsDisabled                       bool   `jsonapi:"attr,is-disabled"`
	NotificationEmail                string `jsonapi:"attr,notification-email"`
	SsoEnabled                       bool   `jsonapi:"attr,sso-enabled"`
	TerraformBuildWorkerApplyTimeout string `jsonapi:"attr,terraform-build-worker-apply-timeout"`
	TerraformBuildWorkerPlanTimeout  string `jsonapi:"attr,terraform-build-worker-plan-timeout"`
	TerraformWorkerSudoEnabled       bool   `jsonapi:"attr,terraform-worker-sudo-enabled"`

	// Relations
	Owners []*User `jsonapi:"relation,owners"`
}

// AdminOrganizationUpdateOptions represents the admin options for updating an organization.
// https://www.terraform.io/docs/cloud/api/admin/organizations.html#request-body
type AdminOrganizationUpdateOptions struct {
	AccessBetaTools                  *bool   `jsonapi:"attr,access-beta-tools,omitempty"`
	GlobalModuleSharing              *bool   `jsonapi:"attr,global-module-sharing,omitempty"`
	IsDisabled                       *bool   `jsonapi:"attr,is-disabled,omitempty"`
	TerraformBuildWorkerApplyTimeout *string `jsonapi:"attr,terraform-build-worker-apply-timeout,omitempty"`
	TerraformBuildWorkerPlanTimeout  *string `jsonapi:"attr,terraform-build-worker-plan-timeout,omitempty"`
	TerraformWorkerSudoEnabled       bool    `jsonapi:"attr,terraform-worker-sudo-enabled,omitempty"`
}

// AdminOrganizationList represents a list of organizations via Admin API.
type AdminOrganizationList struct {
	*Pagination
	Items []*AdminOrganization
}

// AdminOrgIncludeOpt represents the available options for include query params.
// https://www.terraform.io/docs/cloud/api/admin/organizations.html#available-related-resources
type AdminOrgIncludeOpt string

const AdminOrgOwners AdminOrgIncludeOpt = "owners"

// AdminOrganizationListOptions represents the options for listing organizations via Admin API.
type AdminOrganizationListOptions struct {
	ListOptions

	// Optional: A query string used to filter organizations.
	// Any organizations with a name or notification email partially matching this value will be returned.
	Query string `url:"q,omitempty"`
	// Optional: A list of relations to include. See available resources
	// https://www.terraform.io/docs/cloud/api/admin/organizations.html#available-related-resources
	Include []AdminOrgIncludeOpt `url:"include,omitempty"`
}

// AdminOrganizationListModuleConsumersOptions represents the options for listing organization module consumers through the Admin API
type AdminOrganizationListModuleConsumersOptions struct {
	ListOptions
}

type AdminOrganizationID struct {
	ID string `jsonapi:"primary,organizations"`
}

// List all the organizations visible to the current user.
func (s *adminOrganizations) List(ctx context.Context, options *AdminOrganizationListOptions) (*AdminOrganizationList, error) {
	if err := options.valid(); err != nil {
		return nil, err
	}
	u := "admin/organizations"
	req, err := s.client.newRequest("GET", u, options)
	if err != nil {
		return nil, err
	}

	orgl := &AdminOrganizationList{}
	err = s.client.do(ctx, req, orgl)
	if err != nil {
		return nil, err
	}

	return orgl, nil
}

// ListModuleConsumers lists specific organizations in the Terraform Enterprise installation that have permission to use an organization's modules.
func (s *adminOrganizations) ListModuleConsumers(ctx context.Context, organization string, options *AdminOrganizationListModuleConsumersOptions) (*AdminOrganizationList, error) {
	if !validStringID(&organization) {
		return nil, ErrInvalidOrg
	}

	u := fmt.Sprintf("admin/organizations/%s/relationships/module-consumers", url.QueryEscape(organization))

	req, err := s.client.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	orgl := &AdminOrganizationList{}
	err = s.client.do(ctx, req, orgl)
	if err != nil {
		return nil, err
	}

	return orgl, nil
}

// Read an organization by its name.
func (s *adminOrganizations) Read(ctx context.Context, organization string) (*AdminOrganization, error) {
	if !validStringID(&organization) {
		return nil, ErrInvalidOrg
	}

	u := fmt.Sprintf("admin/organizations/%s", url.QueryEscape(organization))
	req, err := s.client.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	org := &AdminOrganization{}
	err = s.client.do(ctx, req, org)
	if err != nil {
		return nil, err
	}

	return org, nil
}

// Update an organization by its name.
func (s *adminOrganizations) Update(ctx context.Context, organization string, options AdminOrganizationUpdateOptions) (*AdminOrganization, error) {
	if !validStringID(&organization) {
		return nil, ErrInvalidOrg
	}

	u := fmt.Sprintf("admin/organizations/%s", url.QueryEscape(organization))
	req, err := s.client.newRequest("PATCH", u, &options)
	if err != nil {
		return nil, err
	}

	org := &AdminOrganization{}
	err = s.client.do(ctx, req, org)
	if err != nil {
		return nil, err
	}

	return org, nil
}

// UpdateModuleConsumers updates an organization to specify a list of organizations that can use modules from the sharing organization's private registry.
func (s *adminOrganizations) UpdateModuleConsumers(ctx context.Context, organization string, consumerOrganizationIDs []string) error {
	if !validStringID(&organization) {
		return ErrInvalidOrg
	}

	u := fmt.Sprintf("admin/organizations/%s/relationships/module-consumers", url.QueryEscape(organization))

	var organizations []*AdminOrganizationID
	for _, id := range consumerOrganizationIDs {
		if !validStringID(&id) {
			return ErrInvalidOrg
		}
		organizations = append(organizations, &AdminOrganizationID{ID: id})
	}

	req, err := s.client.newRequest("PATCH", u, organizations)
	if err != nil {
		return err
	}

	err = s.client.do(ctx, req, nil)
	if err != nil {
		return err
	}

	return nil
}

// Delete an organization by its name.
func (s *adminOrganizations) Delete(ctx context.Context, organization string) error {
	if !validStringID(&organization) {
		return ErrInvalidOrg
	}

	u := fmt.Sprintf("admin/organizations/%s", url.QueryEscape(organization))
	req, err := s.client.newRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}

func (o *AdminOrganizationListOptions) valid() error {
	if o == nil {
		return nil // nothing to validate
	}

	if err := validateAdminOrgIncludeParams(o.Include); err != nil {
		return err
	}

	return nil
}

func validateAdminOrgIncludeParams(params []AdminOrgIncludeOpt) error {
	for _, p := range params {
		switch p {
		case AdminOrgOwners:
			// do nothing
		default:
			return ErrInvalidIncludeValue
		}
	}

	return nil
}
