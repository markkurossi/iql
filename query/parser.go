//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package query

import (
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/markkurossi/iql/data"
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
	t, err := p.lexer.get()
	if err != nil {
		return nil, err
	}
	switch t.Type {
	case TSymSelect:
		return p.parseSelect()

	default:
		return nil, p.errf(t.From, "unexpected token: %s", t)
	}
}

func (p *parser) parseSelect() (*Query, error) {
	var q Query

	// Columns.
	for {
		col, err := p.parseColumn()
		if err != nil {
			return nil, err
		}
		if col == nil {
			break
		}
		q.Select = append(q.Select, *col)

		t, err := p.lexer.get()
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			break
		}
		if t.Type != ',' {
			p.lexer.unget(t)
			break
		}
	}
	return &q, nil
}

func (p *parser) parseColumn() (*ColumnSelector, error) {
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	t, err := p.lexer.get()
	if err != nil {
		if err != io.EOF {
			return nil, err
		}
		return &ColumnSelector{
			Expr: expr,
		}, nil
	}
	if t.Type != TSymAs {
		p.lexer.unget(t)
		return &ColumnSelector{
			Expr: expr,
		}, nil
	}
	t, err = p.lexer.get()
	if err != nil {
		return nil, err
	}
	if t.Type != TIdentifier {
		return nil, p.errf(t.From, "unexpected token: %s", t)
	}

	return &ColumnSelector{
		Expr: expr,
		As:   t.StrVal,
	}, nil
}

func (p *parser) parseExpr() (Expr, error) {
	return p.parseAdditive()
}

func (p *parser) parseAdditive() (Expr, error) {
	left, err := p.parseMultiplicative()
	if err != nil {
		return nil, err
	}
	for {
		t, err := p.lexer.get()
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			return left, nil
		}
		var bt BinaryType

		switch t.Type {
		case '+':
			bt = BinAdd

		case '-':
			bt = BinSub

		default:
			p.lexer.unget(t)
			return left, nil
		}
		right, err := p.parseMultiplicative()
		if err != nil {
			return nil, err
		}
		left = &Binary{
			Type:  bt,
			Left:  left,
			Right: right,
		}
	}
}

func (p *parser) parseMultiplicative() (Expr, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for {
		t, err := p.lexer.get()
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			return left, nil
		}
		var bt BinaryType

		switch t.Type {
		case '*':
			bt = BinMult

		case '/':
			bt = BinDiv

		default:
			p.lexer.unget(t)
			return left, nil
		}
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = &Binary{
			Type:  bt,
			Left:  left,
			Right: right,
		}
	}
}

func (p *parser) parseUnary() (Expr, error) {
	return p.parsePostfix()
}

func (p *parser) parsePostfix() (Expr, error) {
	t, err := p.lexer.get()
	if err != nil {
		return nil, err
	}

	var val data.Value

	switch t.Type {
	case TString:
		val = data.StringValue(t.StrVal)
	case TInteger:
		val = data.IntValue(t.IntVal)
	default:
		p.lexer.unget(t)
		return nil, p.errf(t.From, "unexpected token: %s", t)
	}

	return &Constant{
		Value: val,
	}, nil
}

func (p *parser) errf(loc Point, format string, a ...interface{}) error {
	msg := fmt.Sprintf(format, a...)

	p.lexer.FlushEOL()

	line, ok := p.lexer.history[loc.Line]
	if ok {
		var indicator []rune
		for i := 0; i < loc.Col; i++ {
			var r rune
			if line[i] == '\t' {
				r = '\t'
			} else {
				r = ' '
			}
			indicator = append(indicator, r)
		}
		indicator = append(indicator, '^')
		log.Printf("%s: %s\n%s\n%s\n",
			loc, msg, string(line), string(indicator))

		return errors.New(msg)
	}
	return errors.New(msg)
}
