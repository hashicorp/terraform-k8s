// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// +build appengine

package isatty

// IsTerminal returns true if the file descriptor is terminal which
// is always false on on appengine classic which is a sandboxed PaaS.
func IsTerminal(fd uintptr) bool {
	return false
}
