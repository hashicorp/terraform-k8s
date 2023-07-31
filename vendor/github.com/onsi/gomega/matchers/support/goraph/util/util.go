// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package util

import "math"

func Odd(n int) bool {
	return math.Mod(float64(n), 2.0) == 1.0
}
