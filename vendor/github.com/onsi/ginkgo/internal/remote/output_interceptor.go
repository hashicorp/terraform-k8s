// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package remote

import "os"

/*
The OutputInterceptor is used by the ForwardingReporter to
intercept and capture all stdin and stderr output during a test run.
*/
type OutputInterceptor interface {
	StartInterceptingOutput() error
	StopInterceptingAndReturnOutput() (string, error)
	StreamTo(*os.File)
}
