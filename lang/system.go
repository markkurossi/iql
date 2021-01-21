//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package lang

import (
	"github.com/markkurossi/iql/types"
)

// System variables.
const (
	SysRealFmt  = "REALFMT"
	SysTableFmt = "TABLEFMT"
	SysTermOut  = "TERMOUT"
)

var sysvars = []struct {
	name string
	typ  types.Type
	def  types.Value
}{
	{
		name: SysRealFmt,
		typ:  types.String,
		def:  types.StringValue("%g"),
	},
	{
		name: SysTableFmt,
		typ:  types.String,
		def:  types.StringValue("uc"),
	},
	{
		name: SysTermOut,
		typ:  types.Bool,
		def:  types.BoolValue(true),
	},
}

// InitSystemVariables initializes the global system variables for the
// scope.
func InitSystemVariables(scope *Scope) {
	for _, sysvar := range sysvars {
		scope.Declare(sysvar.name, sysvar.typ)
		scope.Set(sysvar.name, sysvar.def)
	}
}

// Format gets the value formatting options from the scope.
func Format(scope *Scope) *types.Format {
	real := scope.Get(SysRealFmt)
	if real == nil {
		return nil
	}
	_, ok := real.Value.(types.NullValue)
	if ok {
		return nil
	}
	return &types.Format{
		Float: real.Value.String(),
	}
}
