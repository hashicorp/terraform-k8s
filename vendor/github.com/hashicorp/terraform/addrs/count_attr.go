// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package addrs

// CountAttr is the address of an attribute of the "count" object in
// the interpolation scope, like "count.index".
type CountAttr struct {
	referenceable
	Name string
}

func (ca CountAttr) String() string {
	return "count." + ca.Name
}
