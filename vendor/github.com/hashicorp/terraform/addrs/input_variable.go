// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package addrs

import (
	"fmt"
)

// InputVariable is the address of an input variable.
type InputVariable struct {
	referenceable
	Name string
}

func (v InputVariable) String() string {
	return "var." + v.Name
}

// Absolute converts the receiver into an absolute address within the given
// module instance.
func (v InputVariable) Absolute(m ModuleInstance) AbsInputVariableInstance {
	return AbsInputVariableInstance{
		Module:   m,
		Variable: v,
	}
}

// AbsInputVariableInstance is the address of an input variable within a
// particular module instance.
type AbsInputVariableInstance struct {
	Module   ModuleInstance
	Variable InputVariable
}

// InputVariable returns the absolute address of the input variable of the
// given name inside the receiving module instance.
func (m ModuleInstance) InputVariable(name string) AbsInputVariableInstance {
	return AbsInputVariableInstance{
		Module: m,
		Variable: InputVariable{
			Name: name,
		},
	}
}

func (v AbsInputVariableInstance) String() string {
	if len(v.Module) == 0 {
		return v.Variable.String()
	}

	return fmt.Sprintf("%s.%s", v.Module.String(), v.Variable.String())
}
