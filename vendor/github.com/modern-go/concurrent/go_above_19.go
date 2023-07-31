// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//+build go1.9

package concurrent

import "sync"

// Map is a wrapper for sync.Map introduced in go1.9
type Map struct {
	sync.Map
}

// NewMap creates a thread safe Map
func NewMap() *Map {
	return &Map{}
}
