// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package node

type Node struct {
	ID    int
	Value interface{}
}

type NodeOrderedSet []Node
