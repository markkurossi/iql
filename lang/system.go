//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package lang

import (
	"github.com/markkurossi/iql/types"
)

const (
	sysRealFmt = "REALFMT"
	sysTermOut = "TERMOUT"
)

// InitSystemVariables initializes the global system variables for the
// scope.
func InitSystemVariables(scope *Scope) {
	scope.Declare(sysRealFmt, types.String)
	scope.Declare(sysTermOut, types.Bool)
	scope.Set(sysTermOut, types.BoolValue(true))
}

// Format gets the value formatting options from the scope.
func Format(scope *Scope) *types.Format {
	real := scope.Get(sysRealFmt)
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
