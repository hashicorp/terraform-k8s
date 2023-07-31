// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ratelimiter

type Storage interface {
	GetBucketFor(string) (*LeakyBucket, error)
	SetBucketFor(string, LeakyBucket) error
}
