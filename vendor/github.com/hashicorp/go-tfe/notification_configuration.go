// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfe

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// Compile-time proof of interface implementation.
var _ NotificationConfigurations = (*notificationConfigurations)(nil)

// NotificationConfigurations describes all the Notification Configuration
// related methods that the Terraform Enterprise API supports.
//
// TFE API docs:
// https://www.terraform.io/docs/cloud/api/notification-configurations.html
type NotificationConfigurations interface {
	// List all the notification configurations within a workspace.
	List(ctx context.Context, workspaceID string, options *NotificationConfigurationListOptions) (*NotificationConfigurationList, error)

	// Create a new notification configuration with the given options.
	Create(ctx context.Context, workspaceID string, options NotificationConfigurationCreateOptions) (*NotificationConfiguration, error)

	// Read a notification configuration by its ID.
	Read(ctx context.Context, notificationConfigurationID string) (*NotificationConfiguration, error)

	// Update an existing notification configuration.
	Update(ctx context.Context, notificationConfigurationID string, options NotificationConfigurationUpdateOptions) (*NotificationConfiguration, error)

	// Delete a notification configuration by its ID.
	Delete(ctx context.Context, notificationConfigurationID string) error

	// Verify a notification configuration by its ID.
	Verify(ctx context.Context, notificationConfigurationID string) (*NotificationConfiguration, error)
}

// notificationConfigurations implements NotificationConfigurations.
type notificationConfigurations struct {
	client *Client
}

// NotificationTriggerType represents the different TFE notifications that can be sent
// as a run's progress transitions between different states
type NotificationTriggerType string

const (
	NotificationTriggerCreated        NotificationTriggerType = "run:created"
	NotificationTriggerPlanning       NotificationTriggerType = "run:planning"
	NotificationTriggerNeedsAttention NotificationTriggerType = "run:needs_attention"
	NotificationTriggerApplying       NotificationTriggerType = "run:applying"
	NotificationTriggerCompleted      NotificationTriggerType = "run:completed"
	NotificationTriggerErrored        NotificationTriggerType = "run:errored"
)

// NotificationDestinationType represents the destination type of the
// notification configuration.
type NotificationDestinationType string

// List of available notification destination types.
const (
	NotificationDestinationTypeEmail          NotificationDestinationType = "email"
	NotificationDestinationTypeGeneric        NotificationDestinationType = "generic"
	NotificationDestinationTypeSlack          NotificationDestinationType = "slack"
	NotificationDestinationTypeMicrosoftTeams NotificationDestinationType = "microsoft-teams"
)

// NotificationConfigurationList represents a list of Notification
// Configurations.
type NotificationConfigurationList struct {
	*Pagination
	Items []*NotificationConfiguration
}

// NotificationConfiguration represents a Notification Configuration.
type NotificationConfiguration struct {
	ID                string                      `jsonapi:"primary,notification-configurations"`
	CreatedAt         time.Time                   `jsonapi:"attr,created-at,iso8601"`
	DeliveryResponses []*DeliveryResponse         `jsonapi:"attr,delivery-responses"`
	DestinationType   NotificationDestinationType `jsonapi:"attr,destination-type"`
	Enabled           bool                        `jsonapi:"attr,enabled"`
	Name              string                      `jsonapi:"attr,name"`
	Token             string                      `jsonapi:"attr,token"`
	Triggers          []string                    `jsonapi:"attr,triggers"`
	UpdatedAt         time.Time                   `jsonapi:"attr,updated-at,iso8601"`
	URL               string                      `jsonapi:"attr,url"`

	// EmailAddresses is only available for TFE users. It is not available in TFC.
	EmailAddresses []string `jsonapi:"attr,email-addresses"`

	// Relations
	Subscribable *Workspace `jsonapi:"relation,subscribable"`
	EmailUsers   []*User    `jsonapi:"relation,users"`
}

// DeliveryResponse represents a notification configuration delivery response.
type DeliveryResponse struct {
	Body       string              `jsonapi:"attr,body"`
	Code       string              `jsonapi:"attr,code"`
	Headers    map[string][]string `jsonapi:"attr,headers"`
	SentAt     time.Time           `jsonapi:"attr,sent-at,rfc3339"`
	Successful string              `jsonapi:"attr,successful"`
	URL        string              `jsonapi:"attr,url"`
}

// NotificationConfigurationListOptions represents the options for listing
// notification configurations.
type NotificationConfigurationListOptions struct {
	ListOptions
}

// NotificationConfigurationCreateOptions represents the options for
// creating a new notification configuration.
type NotificationConfigurationCreateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,notification-configurations"`

	// Required: The destination type of the notification configuration
	DestinationType *NotificationDestinationType `jsonapi:"attr,destination-type"`

	// Required: Whether the notification configuration should be enabled or not
	Enabled *bool `jsonapi:"attr,enabled"`

	// Required: The name of the notification configuration
	Name *string `jsonapi:"attr,name"`

	// Optional: The token of the notification configuration
	Token *string `jsonapi:"attr,token,omitempty"`

	// Optional: The list of run events that will trigger notifications.
	Triggers []NotificationTriggerType `jsonapi:"attr,triggers,omitempty"`

	// Optional: The url of the notification configuration
	URL *string `jsonapi:"attr,url,omitempty"`

	// Optional: The list of email addresses that will receive notification emails.
	// EmailAddresses is only available for TFE users. It is not available in TFC.
	EmailAddresses []string `jsonapi:"attr,email-addresses,omitempty"`

	// Optional: The list of users belonging to the organization that will receive notification emails.
	EmailUsers []*User `jsonapi:"relation,users,omitempty"`
}

// NotificationConfigurationUpdateOptions represents the options for
// updating a existing notification configuration.
type NotificationConfigurationUpdateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,notification-configurations"`

	// Optional: Whether the notification configuration should be enabled or not
	Enabled *bool `jsonapi:"attr,enabled,omitempty"`

	// Optional: The name of the notification configuration
	Name *string `jsonapi:"attr,name,omitempty"`

	// Optional: The token of the notification configuration
	Token *string `jsonapi:"attr,token,omitempty"`

	// Optional: The list of run events that will trigger notifications.
	Triggers []NotificationTriggerType `jsonapi:"attr,triggers,omitempty"`

	// Optional: The url of the notification configuration
	URL *string `jsonapi:"attr,url,omitempty"`

	// Optional: The list of email addresses that will receive notification emails.
	// EmailAddresses is only available for TFE users. It is not available in TFC.
	EmailAddresses []string `jsonapi:"attr,email-addresses,omitempty"`

	// Optional: The list of users belonging to the organization that will receive notification emails.
	EmailUsers []*User `jsonapi:"relation,users,omitempty"`
}

// List all the notification configurations associated with a workspace.
func (s *notificationConfigurations) List(ctx context.Context, workspaceID string, options *NotificationConfigurationListOptions) (*NotificationConfigurationList, error) {
	if !validStringID(&workspaceID) {
		return nil, ErrInvalidWorkspaceID
	}

	u := fmt.Sprintf("workspaces/%s/notification-configurations", url.QueryEscape(workspaceID))
	req, err := s.client.newRequest("GET", u, options)
	if err != nil {
		return nil, err
	}

	ncl := &NotificationConfigurationList{}
	err = s.client.do(ctx, req, ncl)
	if err != nil {
		return nil, err
	}

	return ncl, nil
}

// Create a notification configuration with the given options.
func (s *notificationConfigurations) Create(ctx context.Context, workspaceID string, options NotificationConfigurationCreateOptions) (*NotificationConfiguration, error) {
	if !validStringID(&workspaceID) {
		return nil, ErrInvalidWorkspaceID
	}
	if err := options.valid(); err != nil {
		return nil, err
	}

	u := fmt.Sprintf("workspaces/%s/notification-configurations", url.QueryEscape(workspaceID))
	req, err := s.client.newRequest("POST", u, &options)
	if err != nil {
		return nil, err
	}

	nc := &NotificationConfiguration{}
	err = s.client.do(ctx, req, nc)
	if err != nil {
		return nil, err
	}

	return nc, nil
}

// Read a notification configuration by its ID.
func (s *notificationConfigurations) Read(ctx context.Context, notificationConfigurationID string) (*NotificationConfiguration, error) {
	if !validStringID(&notificationConfigurationID) {
		return nil, ErrInvalidNotificationConfigID
	}

	u := fmt.Sprintf("notification-configurations/%s", url.QueryEscape(notificationConfigurationID))
	req, err := s.client.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	nc := &NotificationConfiguration{}
	err = s.client.do(ctx, req, nc)
	if err != nil {
		return nil, err
	}

	return nc, nil
}

// Updates a notification configuration with the given options.
func (s *notificationConfigurations) Update(ctx context.Context, notificationConfigurationID string, options NotificationConfigurationUpdateOptions) (*NotificationConfiguration, error) {
	if !validStringID(&notificationConfigurationID) {
		return nil, ErrInvalidNotificationConfigID
	}

	if err := options.valid(); err != nil {
		return nil, err
	}

	u := fmt.Sprintf("notification-configurations/%s", url.QueryEscape(notificationConfigurationID))
	req, err := s.client.newRequest("PATCH", u, &options)
	if err != nil {
		return nil, err
	}

	nc := &NotificationConfiguration{}
	err = s.client.do(ctx, req, nc)
	if err != nil {
		return nil, err
	}

	return nc, nil
}

// Delete a notifications configuration by its ID.
func (s *notificationConfigurations) Delete(ctx context.Context, notificationConfigurationID string) error {
	if !validStringID(&notificationConfigurationID) {
		return ErrInvalidNotificationConfigID
	}

	u := fmt.Sprintf("notification-configurations/%s", url.QueryEscape(notificationConfigurationID))
	req, err := s.client.newRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}

// Verify a notification configuration by delivering a verification
// payload to the configured url.
func (s *notificationConfigurations) Verify(ctx context.Context, notificationConfigurationID string) (*NotificationConfiguration, error) {
	if !validStringID(&notificationConfigurationID) {
		return nil, ErrInvalidNotificationConfigID
	}

	u := fmt.Sprintf(
		"notification-configurations/%s/actions/verify", url.QueryEscape(notificationConfigurationID))
	req, err := s.client.newRequest("POST", u, nil)
	if err != nil {
		return nil, err
	}

	nc := &NotificationConfiguration{}
	err = s.client.do(ctx, req, nc)
	if err != nil {
		return nil, err
	}

	return nc, nil
}

func (o NotificationConfigurationCreateOptions) valid() error {
	if o.DestinationType == nil {
		return ErrRequiredDestinationType
	}
	if o.Enabled == nil {
		return ErrRequiredEnabled
	}
	if !validString(o.Name) {
		return ErrRequiredName
	}

	if !validNotificationTriggerType(o.Triggers) {
		return ErrInvalidNotificationTrigger
	}

	if *o.DestinationType == NotificationDestinationTypeGeneric ||
		*o.DestinationType == NotificationDestinationTypeSlack ||
		*o.DestinationType == NotificationDestinationTypeMicrosoftTeams {

		if o.URL == nil {
			return ErrRequiredURL
		}
	}
	return nil
}

func (o NotificationConfigurationUpdateOptions) valid() error {
	if o.Name != nil && !validString(o.Name) {
		return ErrRequiredName
	}

	if !validNotificationTriggerType(o.Triggers) {
		return ErrInvalidNotificationTrigger
	}

	return nil
}

func validNotificationTriggerType(triggers []NotificationTriggerType) bool {
	for _, t := range triggers {
		switch t {
		case NotificationTriggerApplying,
			NotificationTriggerNeedsAttention,
			NotificationTriggerCompleted,
			NotificationTriggerCreated,
			NotificationTriggerErrored,
			NotificationTriggerPlanning:
			continue
		default:
			return false
		}
	}

	return true
}
