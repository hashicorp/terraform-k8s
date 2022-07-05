package tfe

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// Compile-time proof of interface implementation.
var _ UserTokens = (*userTokens)(nil)

// UserTokens describes all the user token related methods that the
// Terraform Cloud/Enterprise API supports.
//
// TFE API docs:
// https://www.terraform.io/docs/cloud/api/user-tokens.html
type UserTokens interface {
	// List all the tokens of the given user ID.
	List(ctx context.Context, userID string) (*UserTokenList, error)

	// Create a new user token
	Create(ctx context.Context, userID string, options UserTokenCreateOptions) (*UserToken, error)

	// Read a user token by its ID.
	Read(ctx context.Context, tokenID string) (*UserToken, error)

	// Delete a user token by its ID.
	Delete(ctx context.Context, tokenID string) error
}

// userTokens implements UserTokens.
type userTokens struct {
	client *Client
}

// UserTokenList is a list of tokens for the given user ID.
type UserTokenList struct {
	*Pagination
	Items []*UserToken
}

// UserToken represents a Terraform Enterprise user token.
type UserToken struct {
	ID          string    `jsonapi:"primary,authentication-tokens"`
	CreatedAt   time.Time `jsonapi:"attr,created-at,iso8601"`
	Description string    `jsonapi:"attr,description"`
	LastUsedAt  time.Time `jsonapi:"attr,last-used-at,iso8601"`
	Token       string    `jsonapi:"attr,token"`
}

// UserTokenCreateOptions the options for creating a user token.
type UserTokenCreateOptions struct {
	// Optional: Description of the token
	Description string `jsonapi:"attr,description,omitempty"`
}

// Create a new user token
func (s *userTokens) Create(ctx context.Context, userID string, options UserTokenCreateOptions) (*UserToken, error) {
	if !validStringID(&userID) {
		return nil, ErrInvalidUserID
	}

	u := fmt.Sprintf("users/%s/authentication-tokens", url.QueryEscape(userID))
	req, err := s.client.newRequest("POST", u, &options)
	if err != nil {
		return nil, err
	}

	ut := &UserToken{}
	err = s.client.do(ctx, req, ut)
	if err != nil {
		return nil, err
	}

	return ut, err
}

// List shows existing user tokens
func (s *userTokens) List(ctx context.Context, userID string) (*UserTokenList, error) {
	if !validStringID(&userID) {
		return nil, ErrInvalidUserID
	}

	u := fmt.Sprintf("users/%s/authentication-tokens", url.QueryEscape(userID))
	req, err := s.client.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	tl := &UserTokenList{}
	err = s.client.do(ctx, req, tl)
	if err != nil {
		return nil, err
	}

	return tl, err
}

// Read a user token by its ID.
func (s *userTokens) Read(ctx context.Context, tokenID string) (*UserToken, error) {
	if !validStringID(&tokenID) {
		return nil, ErrInvalidTokenID
	}

	u := fmt.Sprintf("authentication-tokens/%s", url.QueryEscape(tokenID))
	req, err := s.client.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	tt := &UserToken{}
	err = s.client.do(ctx, req, tt)
	if err != nil {
		return nil, err
	}

	return tt, err
}

// Delete a user token by its ID.
func (s *userTokens) Delete(ctx context.Context, tokenID string) error {
	if !validStringID(&tokenID) {
		return ErrInvalidTokenID
	}

	u := fmt.Sprintf("authentication-tokens/%s", url.QueryEscape(tokenID))
	req, err := s.client.newRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}
