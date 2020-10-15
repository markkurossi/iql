//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package query

import (
	"fmt"
	"io"
)

// Parse parses the query input and returns a query object.
func Parse(input io.Reader, source string) (*Query, error) {
	p := &parser{
		lexer: newLexer(input, source),
	}

	return p.parse()
}

type parser struct {
	lexer *lexer
}

func (p *parser) parse() (*Query, error) {
	return nil, fmt.Errorf("parser.Parse not implemented yet")
}
