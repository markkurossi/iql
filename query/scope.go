//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package query

import (
	"github.com/markkurossi/iql/types"
)

// Scope implements name scope.
type Scope struct {
	Symbols map[string]types.Value
}
