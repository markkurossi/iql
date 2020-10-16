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
	"strings"

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
	q := new(Query)

	// Columns.
	for {
		col, err := p.parseColumn()
		if err != nil {
			return nil, err
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

	// From
	t, err := p.lexer.get()
	if err != nil {
		if err != io.EOF {
			return nil, err
		}
		return q, nil
	}
	if t.Type != TSymFrom {
		return nil, p.errf(t.From, "unexpected token: %s", t)
	}
	for {
		source, err := p.parseSource(q)
		if err != nil {
			return nil, err
		}
		q.From = append(q.From, *source)

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

	// Where
	t, err = p.lexer.get()
	if err != nil {
		if err != io.EOF {
			return nil, err
		}
		return q, nil
	}
	if t.Type != TSymWhere {
		return nil, p.errf(t.From, "unexpected token: %s", t)
	}
	q.Where, err = p.parseExpr()
	if err != nil {
		return nil, err
	}

	return q, nil
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

func (p *parser) parseSource(q *Query) (*SourceSelector, error) {
	t, err := p.lexer.get()
	if err != nil {
		return nil, err
	}
	if t.Type != TString {
		return nil, p.errf(t.From, "unexpected token: %s", t)
	}

	filter, err := p.parseKeyword(TSymFilter)
	if err != nil {
		return nil, err
	}
	as, err := p.parseKeyword(TSymAs)
	if err != nil {
		return nil, err
	}

	source, err := data.New(t.StrVal, filter, columnsFor(q.Select, as))
	if err != nil {
		return nil, err
	}

	return &SourceSelector{
		Source: source,
		As:     as,
	}, nil
}

func columnsFor(columns []ColumnSelector, source string) []data.ColumnSelector {
	var result []data.ColumnSelector

	for _, col := range columns {
		switch c := col.Expr.(type) {
		case *Reference:
			if c.Source == source {
				result = append(result, data.ColumnSelector{
					Name: c.Reference,
					As:   col.As,
				})
				// We passed the source specific selectors down to the
				// source and from now on, we are referencing the
				// fields with their alias names.
				c.Column = col.As
			}
		}
	}

	return result
}

func (p *parser) parseKeyword(keyword TokenType) (string, error) {
	t, err := p.lexer.get()
	if err != nil {
		if err != io.EOF {
			return "", err
		}
		return "", nil
	}
	if t.Type != keyword {
		p.lexer.unget(t)
		return "", nil
	}
	t, err = p.lexer.get()
	if err != nil {
		return "", err
	}
	switch t.Type {
	case TIdentifier, TString:
	default:
		return "", p.errf(t.From, "unexpected token: %s", t)
	}
	return t.StrVal, nil
}

func (p *parser) parseExpr() (Expr, error) {
	return p.parseExprLogicalOr()
}

func (p *parser) parseExprLogicalOr() (Expr, error) {
	return p.parseExprLogicalAnd()
}

func (p *parser) parseExprLogicalAnd() (Expr, error) {
	left, err := p.parseExprLogicalNot()
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
		if t.Type != TAnd {
			p.lexer.unget(t)
			return left, nil
		}
		right, err := p.parseExprLogicalNot()
		if err != nil {
			return nil, err
		}
		left = &Binary{
			Type:  BinAnd,
			Left:  left,
			Right: right,
		}
	}
}

func (p *parser) parseExprLogicalNot() (Expr, error) {
	return p.parseExprComparative()
}

func (p *parser) parseExprComparative() (Expr, error) {
	left, err := p.parseExprAdditive()
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
		case '=':
			bt = BinEq
		case TNeq:
			bt = BinNeq
		case '<':
			bt = BinLt
		case TLe:
			bt = BinLe
		case '>':
			bt = BinGt
		case TGe:
			bt = BinGe
		default:
			p.lexer.unget(t)
			return left, nil
		}
		right, err := p.parseExprAdditive()
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

func (p *parser) parseExprAdditive() (Expr, error) {
	left, err := p.parseExprMultiplicative()
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
		right, err := p.parseExprMultiplicative()
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

func (p *parser) parseExprMultiplicative() (Expr, error) {
	left, err := p.parseExprUnary()
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
		right, err := p.parseExprUnary()
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

func (p *parser) parseExprUnary() (Expr, error) {
	return p.parseExprPostfix()
}

func (p *parser) parseExprPostfix() (Expr, error) {
	t, err := p.lexer.get()
	if err != nil {
		return nil, err
	}

	var val data.Value

	switch t.Type {
	case TIdentifier:
		var source, column string

		n, err := p.lexer.get()
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			column = t.StrVal
		} else if n.Type == '(' {
			return p.parseFunc(t)
		} else if n.Type == '.' {
			n, err := p.lexer.get()
			if err != nil {
				return nil, err
			}
			switch n.Type {
			case TIdentifier, TString:
				source = t.StrVal
				column = n.StrVal

			case TInteger:
				source = t.StrVal
				column = fmt.Sprintf("%d", n.IntVal)

			default:
				return nil, p.errf(n.From, "unexpected token: %s", n)
			}
		} else {
			p.lexer.unget(n)
			column = t.StrVal
		}
		return &Reference{
			Reference: data.Reference{
				Source: source,
				Column: column,
			},
		}, nil

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

func (p *parser) parseFunc(name *Token) (Expr, error) {
	var args []Expr

	for {
		t, err := p.lexer.get()
		if err != nil {
			return nil, err
		}
		if t.Type == ')' {
			break
		}
		p.lexer.unget(t)

		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		args = append(args, expr)

		t, err = p.lexer.get()
		if err != nil {
			return nil, err
		}
		if t.Type != ',' {
			p.lexer.unget(t)
		}
	}
	f, ok := functions[strings.ToUpper(name.StrVal)]
	if !ok {
		return nil, p.errf(name.From, "unknown function: %s", name.StrVal)
	}
	return &Function{
		Type:      f,
		Arguments: args,
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
