package tfe

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/google/go-querystring/query"
	retryablehttp "github.com/hashicorp/go-retryablehttp"
)

// Compile-time proof of interface implementation
var _ AuditTrails = (*auditTrails)(nil)

// AuditTrails describes all the audit event related methods that the Terraform
// Cloud API supports.
// **Note:** These methods require the client to be configured with an organization token for
// an organization in the Business tier. Furthermore, these methods are only available in Terraform Cloud.
//
// TFC API Docs: https://www.terraform.io/cloud-docs/api-docs/audit-trails
type AuditTrails interface {
	// Read all the audit events in an organization.
	List(ctx context.Context, options *AuditTrailListOptions) (*AuditTrailList, error)
}

// auditTrails implements AuditTrails
type auditTrails struct {
	client *Client
}

// AuditTrailRequest represents the request details of the audit event.
type AuditTrailRequest struct {
	ID string `json:"id"`
}

// AuditTrailAuth represents the details of the actor that invoked the audit event.
type AuditTrailAuth struct {
	AccessorID     string  `json:"accessor_id"`
	Description    string  `json:"description"`
	Type           string  `json:"type"`
	ImpersonatorID *string `json:"impersonator_id"`
	OrganizationID string  `json:"organization_id"`
}

// AuditTrailResource represents the details of the API resource in the audit event.
type AuditTrailResource struct {
	ID     string                 `json:"id"`
	Type   string                 `json:"type"`
	Action string                 `json:"action"`
	Meta   map[string]interface{} `json:"meta"`
}

// AuditTrail represents an event in the TFC audit log.
type AuditTrail struct {
	ID        string    `json:"id"`
	Version   string    `json:"version"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`

	Auth     AuditTrailAuth     `json:"auth"`
	Request  AuditTrailRequest  `json:"request"`
	Resource AuditTrailResource `json:"resource"`
}

// AuditTrailList represents a list of audit trails.
type AuditTrailList struct {
	*Pagination

	Items []*AuditTrail `json:"data"`
}

// AuditTrailListOptions represents the options for listing audit trails.
type AuditTrailListOptions struct {
	// Optional: Returns only audit trails created after this date
	Since time.Time `url:"since,omitempty"`
	*ListOptions
}

// List all the audit events in an organization.
func (s *auditTrails) List(ctx context.Context, options *AuditTrailListOptions) (*AuditTrailList, error) {
	u, err := s.client.baseURL.Parse("/api/v2/organization/audit-trail")
	if err != nil {
		return nil, err
	}

	headers := make(http.Header)
	headers.Set("User-Agent", _userAgent)
	headers.Set("Authorization", "Bearer "+s.client.token)
	headers.Set("Content-Type", "application/json")

	if options != nil {
		q, err := query.Values(options)
		if err != nil {
			return nil, err
		}

		u.RawQuery = encodeQueryParams(q)
	}

	req, err := retryablehttp.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	// Attach the headers to the request
	for k, v := range headers {
		req.Header[k] = v
	}

	if err := s.client.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	resp, err := s.client.http.Do(req.WithContext(ctx))
	if err != nil {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			return nil, err
		}
	}
	defer resp.Body.Close()

	if err := checkResponseCode(resp); err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	atl := &AuditTrailList{}
	if err := json.Unmarshal(body, atl); err != nil {
		return nil, err
	}

	return atl, nil
}
