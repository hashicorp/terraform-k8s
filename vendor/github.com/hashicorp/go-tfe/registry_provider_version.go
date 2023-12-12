package tfe

import (
	"context"
	"fmt"
	"net/url"
)

// Compile-time proof of interface implementation.
var _ RegistryProviderVersions = (*registryProviderVersions)(nil)

// RegistryProviderVersions describes the registry provider version methods that
// the Terraform Enterprise API supports.
//
// TFE API docs: https://www.terraform.io/cloud-docs/api-docs/private-registry/provider-versions-platforms
type RegistryProviderVersions interface {
	// List all versions for a single provider.
	List(ctx context.Context, providerID RegistryProviderID, options *RegistryProviderVersionListOptions) (*RegistryProviderVersionList, error)

	// Create a registry provider version.
	Create(ctx context.Context, providerID RegistryProviderID, options RegistryProviderVersionCreateOptions) (*RegistryProviderVersion, error)

	// Read a registry provider version.
	Read(ctx context.Context, versionID RegistryProviderVersionID) (*RegistryProviderVersion, error)

	// Delete a registry provider version.
	Delete(ctx context.Context, versionID RegistryProviderVersionID) error
}

// registryProvidersVersions implements RegistryProvidersVersions
type registryProviderVersions struct {
	client *Client
}

// RegistryProviderVersion represents a registry provider version
type RegistryProviderVersion struct {
	ID                 string                             `jsonapi:"primary,registry-provider-versions"`
	Version            string                             `jsonapi:"attr,version"`
	CreatedAt          string                             `jsonapi:"attr,created-at,iso8601"`
	UpdatedAt          string                             `jsonapi:"attr,updated-at,iso8601"`
	KeyID              string                             `jsonapi:"attr,key-id"`
	Protocols          []string                           `jsonapi:"attr,protocols"`
	Permissions        RegistryProviderVersionPermissions `jsonapi:"attr,permissions"`
	ShasumsUploaded    bool                               `jsonapi:"attr,shasums-uploaded"`
	ShasumsSigUploaded bool                               `jsonapi:"attr,shasums-sig-uploaded"`

	// Relations
	RegistryProvider          *RegistryProvider           `jsonapi:"relation,registry-provider"`
	RegistryProviderPlatforms []*RegistryProviderPlatform `jsonapi:"relation,platforms"`

	// Links
	Links map[string]interface{} `jsonapi:"links,omitempty"`
}

// RegistryProviderVersionID is the multi key ID for addressing a version provider
type RegistryProviderVersionID struct {
	RegistryProviderID
	Version string
}

type RegistryProviderVersionPermissions struct {
	CanDelete      bool `jsonapi:"attr,can-delete"`
	CanUploadAsset bool `jsonapi:"attr,can-upload-asset"`
}

type RegistryProviderVersionList struct {
	*Pagination
	Items []*RegistryProviderVersion
}

type RegistryProviderVersionListOptions struct {
	ListOptions
}

type RegistryProviderVersionCreateOptions struct {
	// Required: A valid semver version string.
	Version   string   `jsonapi:"attr,version"`

	// Required: A valid gpg-key string.
	KeyID     string   `jsonapi:"attr,key-id"`

	// Required: An array of Terraform provider API versions that this version supports.
	Protocols []string `jsonapi:"attr,protocols"`
}

// List registry provider versions
func (r *registryProviderVersions) List(ctx context.Context, providerID RegistryProviderID, options *RegistryProviderVersionListOptions) (*RegistryProviderVersionList, error) {
	if err := providerID.valid(); err != nil {
		return nil, err
	}
	if err := options.valid(); err != nil {
		return nil, err
	}

	u := fmt.Sprintf(
		"organizations/%s/registry-providers/%s/%s/%s/versions",
		url.QueryEscape(providerID.OrganizationName),
		url.QueryEscape(string(providerID.RegistryName)),
		url.QueryEscape(providerID.Namespace),
		url.QueryEscape(providerID.Name),
	)
	req, err := r.client.newRequest("GET", u, options)
	if err != nil {
		return nil, err
	}

	pvl := &RegistryProviderVersionList{}
	err = r.client.do(ctx, req, pvl)
	if err != nil {
		return nil, err
	}

	return pvl, nil
}

// Create a registry provider version
func (r *registryProviderVersions) Create(ctx context.Context, providerID RegistryProviderID, options RegistryProviderVersionCreateOptions) (*RegistryProviderVersion, error) {
	if err := providerID.valid(); err != nil {
		return nil, err
	}

	if providerID.RegistryName != PrivateRegistry {
		return nil, ErrRequiredPrivateRegistry
	}

	if err := options.valid(); err != nil {
		return nil, err
	}

	u := fmt.Sprintf(
		"organizations/%s/registry-providers/%s/%s/%s/versions",
		url.QueryEscape(providerID.OrganizationName),
		url.QueryEscape(string(providerID.RegistryName)),
		url.QueryEscape(providerID.Namespace),
		url.QueryEscape(providerID.Name),
	)
	req, err := r.client.newRequest("POST", u, &options)
	if err != nil {
		return nil, err
	}

	prvv := &RegistryProviderVersion{}
	err = r.client.do(ctx, req, prvv)
	if err != nil {
		return nil, err
	}

	return prvv, nil
}

// Read a registry provider version
func (r *registryProviderVersions) Read(ctx context.Context, versionID RegistryProviderVersionID) (*RegistryProviderVersion, error) {
	if err := versionID.valid(); err != nil {
		return nil, err
	}

	u := fmt.Sprintf(
		"organizations/%s/registry-providers/%s/%s/%s/versions/%s",
		url.QueryEscape(versionID.OrganizationName),
		url.QueryEscape(string(versionID.RegistryName)),
		url.QueryEscape(versionID.Namespace),
		url.QueryEscape(versionID.Name),
		url.QueryEscape(versionID.Version),
	)
	req, err := r.client.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	prvv := &RegistryProviderVersion{}
	err = r.client.do(ctx, req, prvv)
	if err != nil {
		return nil, err
	}

	return prvv, nil
}

// Delete a registry provider version
func (r *registryProviderVersions) Delete(ctx context.Context, versionID RegistryProviderVersionID) error {
	if err := versionID.valid(); err != nil {
		return err
	}

	u := fmt.Sprintf(
		"organizations/%s/registry-providers/%s/%s/%s/versions/%s",
		url.QueryEscape(versionID.OrganizationName),
		url.QueryEscape(string(versionID.RegistryName)),
		url.QueryEscape(versionID.Namespace),
		url.QueryEscape(versionID.Name),
		url.QueryEscape(versionID.Version),
	)
	req, err := r.client.newRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	return r.client.do(ctx, req, nil)
}

// ShasumsUploadURL returns the upload URL to upload shasums if one is available
func (v *RegistryProviderVersion) ShasumsUploadURL() (string, error) {
	uploadURL, ok := v.Links["shasums-upload"].(string)
	if !ok {
		return uploadURL, fmt.Errorf("the Registry Provider Version does not contain a shasums upload link")
	}
	if uploadURL == "" {
		return uploadURL, fmt.Errorf("the Registry Provider Version shasums upload URL is empty")
	}
	return uploadURL, nil
}

// ShasumsSigUploadURL returns the URL to upload a shasums sig
func (v *RegistryProviderVersion) ShasumsSigUploadURL() (string, error) {
	uploadURL, ok := v.Links["shasums-sig-upload"].(string)
	if !ok {
		return uploadURL, fmt.Errorf("the Registry Provider Version does not contain a shasums sig upload link")
	}
	if uploadURL == "" {
		return uploadURL, fmt.Errorf("the Registry Provider Version shasums sig upload URL is empty")
	}
	return uploadURL, nil
}

// ShasumsDownloadURL returns the URL to download the shasums for the registry version
func (v *RegistryProviderVersion) ShasumsDownloadURL() (string, error) {
	downloadURL, ok := v.Links["shasums-download"].(string)
	if !ok {
		return downloadURL, fmt.Errorf("the Registry Provider Version does not contain a shasums download link")
	}
	if downloadURL == "" {
		return downloadURL, fmt.Errorf("the Registry Provider Version shasums download URL is empty")
	}
	return downloadURL, nil
}

// ShasumsSigDownloadURL returns the URL to download the shasums sig for the registry version
func (v *RegistryProviderVersion) ShasumsSigDownloadURL() (string, error) {
	downloadURL, ok := v.Links["shasums-sig-download"].(string)
	if !ok {
		return downloadURL, fmt.Errorf("the Registry Provider Version does not contain a shasums sig download link")
	}
	if downloadURL == "" {
		return downloadURL, fmt.Errorf("the Registry Provider Version shasums sig download URL is empty")
	}
	return downloadURL, nil
}

func (id RegistryProviderVersionID) valid() error {
	if !validStringID(&id.Version) {
		return ErrInvalidVersion
	}
	if id.RegistryName != PrivateRegistry {
		return ErrRequiredPrivateRegistry
	}
	if err := id.RegistryProviderID.valid(); err != nil {
		return err
	}
	return nil
}

func (o *RegistryProviderVersionListOptions) valid() error {
	return nil
}

func (o RegistryProviderVersionCreateOptions) valid() error {
	if !validStringID(&o.Version) {
		return ErrInvalidVersion
	}
	if !validStringID(&o.KeyID) {
		return ErrInvalidKeyID
	}
	return nil
}
