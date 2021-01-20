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
	SysRealFmt = "REALFMT"
	SysTermOut = "TERMOUT"
)

// InitSystemVariables initializes the global system variables for the
// scope.
func InitSystemVariables(scope *Scope) {
	scope.Declare(SysRealFmt, types.String)
	scope.Declare(SysTermOut, types.Bool)
	scope.Set(SysTermOut, types.BoolValue(true))
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
