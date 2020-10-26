//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package query

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/markkurossi/iql/data"
	"github.com/markkurossi/iql/types"
)

// Parser implements IQL parser.
type Parser struct {
	lexer   *lexer
	nesting int
	global  *Scope
}

// NewParser creates a new IQL parser.
func NewParser(input io.Reader, source string) *Parser {
	return &Parser{
		lexer:  newLexer(input, source),
		global: NewScope(nil),
	}
}

// SetString defines the global string variable with value.
func (p *Parser) SetString(name, value string) error {
	b := p.global.Get(name)
	if b == nil {
		p.global.Declare(name, types.String)
	}
	return p.global.Set(name, types.StringValue(value))
}

func (p *Parser) get() (*Token, error) {
	t, err := p.lexer.get()
	if err != nil {
		return nil, p.err(err)
	}
	return t, nil
}

// Parse parses the next query from the parser's input.
func (p *Parser) Parse() (*Query, error) {
	p.nesting++
	defer func() {
		p.nesting--
	}()

	for {
		t, err := p.lexer.get()
		if err != nil {
			return nil, err
		}
		switch t.Type {
		case TSymDeclare:
			err = p.parseDeclare()
			if err != nil {
				return nil, err
			}

		case TSymSet:
			err = p.parseSet()
			if err != nil {
				return nil, err
			}

		case TSymPrint:
			err = p.parsePrint()
			if err != nil {
				return nil, err
			}

		case TSymSelect:
			return p.parseSelect()

		default:
			return nil, p.errUnexpected(t)
		}
	}
}

func (p *Parser) parseDeclare() error {
	t, err := p.get()
	if err != nil {
		return err
	}
	if t.Type != TIdentifier {
		return p.errUnexpected(t)
	}
	name := t.StrVal
	binding := p.global.Get(name)
	if binding != nil {
		return p.errf(t.From, "identifier '%s' already declared", name)
	}

	typ, err := p.parseType()
	if err != nil {
		return err
	}

	t, err = p.get()
	if err != nil {
		return err
	}
	if t.Type != ';' {
		return p.errUnexpected(t)
	}

	p.global.Declare(name, typ)

	return nil
}

func (p *Parser) parseType() (types.Type, error) {
	t, err := p.get()
	if err != nil {
		return 0, err
	}
	switch t.Type {
	case TSymBoolean:
		return types.Bool, nil
	case TSymInteger:
		return types.Int, nil
	case TSymReal:
		return types.Float, nil
	case TSymVarchar:
		return types.String, nil
	default:
		return 0, p.errUnexpected(t)
	}
}

func (p *Parser) parseSet() error {
	t, err := p.get()
	if err != nil {
		return err
	}
	if t.Type != TIdentifier {
		return p.errUnexpected(t)
	}
	name := t.StrVal
	t, err = p.get()
	if err != nil {
		return err
	}
	if t.Type != '=' {
		return p.errUnexpected(t)
	}
	expr, err := p.parseExpr()
	if err != nil {
		return err
	}
	t, err = p.get()
	if err != nil {
		return err
	}
	if t.Type != ';' {
		return p.errUnexpected(t)
	}

	q := NewQuery(p.global)
	err = expr.Bind(q)
	if err != nil {
		return err
	}
	v, err := expr.Eval(nil, nil, nil)
	if err != nil {
		return err
	}

	return p.global.Set(name, v)
}

func (p *Parser) parsePrint() error {
	expr, err := p.parseExpr()
	if err != nil {
		return err
	}
	t, err := p.get()
	if err != nil {
		return err
	}
	if t.Type != ';' {
		return p.errUnexpected(t)
	}
	v, err := expr.Eval(nil, nil, nil)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", v)
	return nil
}

func (p *Parser) parseSelect() (*Query, error) {
	q := NewQuery(p.global)

	// Columns
	for {
		col, err := p.parseColumn()
		if err != nil {
			return nil, err
		}
		q.Select = append(q.Select, *col)

		t, err := p.get()
		if err != nil {
			return nil, err
		}
		if t.Type != ',' {
			p.lexer.unget(t)
			break
		}
	}

	// Into
	t, err := p.get()
	if err != nil {
		return nil, err
	}
	if t.Type == TSymInto {
		t, err = p.get()
		if err != nil {
			return nil, err
		}
		if t.Type != TIdentifier {
			return nil, p.errUnexpected(t)
		}
		err = q.Global.Declare(t.StrVal, types.Table)
		if err != nil {
			return nil, err
		}
		err = q.Global.Set(t.StrVal, types.TableValue{
			Source: q,
		})
		if err != nil {
			return nil, err
		}
	} else {
		p.lexer.unget(t)
	}

	// From
	t, err = p.get()
	if err != nil {
		return nil, err
	}
	if t.Type == TSymFrom {
		for {
			source, err := p.parseSource(q)
			if err != nil {
				return nil, err
			}
			q.From = append(q.From, *source)

			t, err := p.get()
			if err != nil {
				return nil, err
			}
			if t.Type != ',' {
				p.lexer.unget(t)
				break
			}
		}
	} else {
		p.lexer.unget(t)
	}

	// Where
	t, err = p.get()
	if err != nil {
		return nil, err
	}
	if t.Type == TSymWhere {
		q.Where, err = p.parseExpr()
		if err != nil {
			return nil, err
		}
	} else {
		p.lexer.unget(t)
	}

	// Group by
	t, err = p.get()
	if err != nil {
		return nil, err
	}
	if t.Type == TSymGroup {
		q.GroupBy, err = p.parseGroupBy()
		if err != nil {
			return nil, err
		}
	} else {
		p.lexer.unget(t)
	}

	// Terminator.
	var expected TokenType
	if p.nesting == 1 {
		expected = ';'
	} else {
		expected = ')'
	}
	t, err = p.get()
	if err != nil {
		return nil, err
	}
	if t.Type != expected {
		return nil, p.errUnexpected(t)
	}
	return q, nil
}

func (p *Parser) parseColumn() (*ColumnSelector, error) {
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	t, err := p.get()
	if err != nil {
		return nil, err
	}
	var as string
	if t.Type == TSymAs {
		t, err = p.get()
		if err != nil {
			return nil, err
		}
		if t.Type != TIdentifier {
			return nil, p.errUnexpected(t)
		}
		as = t.StrVal
	} else {
		p.lexer.unget(t)
	}

	return &ColumnSelector{
		Expr: expr,
		As:   as,
	}, nil
}

func (p *Parser) parseSource(q *Query) (*SourceSelector, error) {
	var source types.Source
	var as string

	t, err := p.get()
	if err != nil {
		return nil, err
	}
	if t.Type == '(' {
		source, err = p.Parse()
		if err != nil {
			return nil, err
		}
		as, err = p.parseKeyword(TSymAs)
		if err != nil {
			return nil, err
		}
	} else {
		var url string

		switch t.Type {
		case TIdentifier:
			b := q.Global.Get(t.StrVal)
			if b == nil {
				return nil, p.errf(t.From, "unknown identifier '%s'", t.StrVal)
			}
			if b.Value == types.Null {
				return nil, p.errf(t.From, "identifier '%s' unset", t.StrVal)
			}
			if b.Type == types.String {
				url = b.Value.String()
			} else if b.Type == types.Table {
				table, ok := b.Value.(types.TableValue)
				if !ok {
					return nil, p.errf(t.From,
						"invalid table value for identifier '%s'", t.StrVal)
				}
				source = table.Source
				// Use the symbol name as the default alias. The 'AS'
				// below can override this.
				as = t.StrVal
			} else {
				return nil, p.errf(t.From, "invalid source type: %s", b.Type)
			}

		case TString:
			url = t.StrVal
		default:
			return nil, p.errUnexpected(t)
		}

		filter, err := p.parseKeyword(TSymFilter)
		if err != nil {
			return nil, err
		}
		alias, err := p.parseKeyword(TSymAs)
		if err != nil {
			return nil, err
		}
		if len(alias) > 0 {
			as = alias
		}

		if source == nil {
			source, err = data.New(url, filter, columnsFor(q.Select, as))
			if err != nil {
				return nil, err
			}
		}
	}

	return &SourceSelector{
		Source: source,
		As:     as,
	}, nil
}

func columnsFor(columns []ColumnSelector,
	source string) []types.ColumnSelector {

	var result []types.ColumnSelector

	for _, col := range columns {
		switch c := col.Expr.(type) {
		case *Reference:
			if c.Source == source {
				result = append(result, types.ColumnSelector{
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

func (p *Parser) parseKeyword(keyword TokenType) (string, error) {
	t, err := p.get()
	if err != nil {
		return "", err
	}
	if t.Type != keyword {
		p.lexer.unget(t)
		return "", nil
	}
	t, err = p.get()
	if err != nil {
		return "", err
	}
	switch t.Type {
	case TIdentifier, TString:
	default:
		return "", p.errUnexpected(t)
	}
	return t.StrVal, nil
}

func (p *Parser) parseGroupBy() ([]Expr, error) {
	t, err := p.get()
	if err != nil {
		return nil, err
	}
	if t.Type != TSymBy {
		return nil, p.errUnexpected(t)
	}
	var result []Expr
	for {
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		result = append(result, expr)

		t, err = p.get()
		if t.Type != ',' {
			p.lexer.unget(t)
			return result, nil
		}
	}
}

func (p *Parser) parseExpr() (Expr, error) {
	return p.parseExprLogicalOr()
}

func (p *Parser) parseExprLogicalOr() (Expr, error) {
	return p.parseExprLogicalAnd()
}

func (p *Parser) parseExprLogicalAnd() (Expr, error) {
	left, err := p.parseExprLogicalNot()
	if err != nil {
		return nil, err
	}
	for {
		t, err := p.get()
		if err != nil {
			return nil, err
		}
		if t.Type != TAnd {
			p.lexer.unget(t)
			return left, nil
		}
		right, err := p.parseExprLogicalNot()
		if err != nil {
			return nil, err
		}
		left = &And{
			Left:  left,
			Right: right,
		}
	}
}

func (p *Parser) parseExprLogicalNot() (Expr, error) {
	return p.parseExprComparative()
}

func (p *Parser) parseExprComparative() (Expr, error) {
	left, err := p.parseExprAdditive()
	if err != nil {
		return nil, err
	}
	for {
		t, err := p.get()
		if err != nil {
			return nil, err
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

func (p *Parser) parseExprAdditive() (Expr, error) {
	left, err := p.parseExprMultiplicative()
	if err != nil {
		return nil, err
	}
	for {
		t, err := p.get()
		if err != nil {
			return nil, err
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

func (p *Parser) parseExprMultiplicative() (Expr, error) {
	left, err := p.parseExprUnary()
	if err != nil {
		return nil, err
	}
	for {
		t, err := p.get()
		if err != nil {
			return nil, err
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

func (p *Parser) parseExprUnary() (Expr, error) {
	return p.parseExprPostfix()
}

func (p *Parser) parseExprPostfix() (Expr, error) {
	t, err := p.get()
	if err != nil {
		return nil, err
	}

	var val types.Value

	switch t.Type {
	case '(':
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		t, err = p.get()
		if err != nil {
			return nil, err
		}
		if t.Type != ')' {
			return nil, p.errUnexpected(t)
		}
		return expr, nil

	case TIdentifier:
		var source, column string

		n, err := p.get()
		if err != nil {
			return nil, err
		} else if n.Type == '(' {
			return p.parseFunc(t)
		} else if n.Type == '.' {
			n, err := p.get()
			if err != nil {
				return nil, err
			}
			switch n.Type {
			case TIdentifier, TString:
				source = t.StrVal
				column = n.StrVal

			case TInt:
				source = t.StrVal
				column = fmt.Sprintf("%d", n.IntVal)

			default:
				return nil, p.errUnexpected(n)
			}
		} else {
			p.lexer.unget(n)
			column = t.StrVal
		}
		return &Reference{
			Reference: types.Reference{
				Source: source,
				Column: column,
			},
		}, nil

	case TString:
		val = types.StringValue(t.StrVal)
	case TInt:
		val = types.IntValue(t.IntVal)
	case TFloat:
		val = types.FloatValue(t.FloatVal)
	case TNull:
		val = types.Null
	default:
		p.lexer.unget(t)
		return nil, p.errUnexpected(t)
	}

	return &Constant{
		Value: val,
	}, nil
}

func (p *Parser) parseFunc(name *Token) (Expr, error) {
	var args []Expr

	for {
		t, err := p.get()
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

		t, err = p.get()
		if err != nil {
			return nil, err
		}
		if t.Type != ',' {
			p.lexer.unget(t)
		}
	}
	f := builtIn(strings.ToUpper(name.StrVal))
	if f == nil {
		return nil, p.errf(name.From, "unknown function: %s", name.StrVal)
	}
	return &Call{
		Function:  f,
		Arguments: args,
	}, nil
}

func (p *Parser) errUnexpected(t *Token) error {
	return p.errf(t.From, "unexpected token: %s", t)
}

func (p *Parser) errf(loc Point, format string, a ...interface{}) error {
	return p.error(loc, fmt.Errorf(format, a...))
}

func (p *Parser) err(err error) error {
	return p.error(p.lexer.point, err)
}

func (p *Parser) error(loc Point, err error) error {
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
			loc, err, string(line), string(indicator))
	}
	return fmt.Errorf("%s: %s", loc, err)
}
