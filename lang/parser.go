//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package lang

import (
	"fmt"
	"io"
	"log"
	"math"
	"strings"

	"github.com/markkurossi/iql/data"
	"github.com/markkurossi/iql/types"
)

// Parser implements IQL parser.
type Parser struct {
	lexer   *lexer
	nesting int
	global  *Scope
	output  io.Writer
}

// NewParser creates a new IQL parser.
func NewParser(global *Scope, input io.Reader, source string,
	output io.Writer) *Parser {

	return &Parser{
		lexer:  newLexer(input, source),
		global: global,
		output: output,
	}
}

// SetString defines the global string variable with value.
func (p *Parser) SetString(name, value string) error {
	b := p.global.Get(name)
	if b == nil {
		p.global.Declare(name, types.String, nil)
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

func (p *Parser) need(tt TokenType) (*Token, error) {
	t, err := p.get()
	if err != nil {
		return nil, err
	}
	if t.Type != tt {
		return nil, p.errUnexpected(t)
	}
	return t, nil
}

func (p *Parser) optional(tt TokenType) (*Token, error) {
	t, err := p.get()
	if err != nil {
		return nil, err
	}
	if t.Type != tt {
		p.lexer.unget(t)
		return nil, nil
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
		case ';':
			// Empty statement.

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

		case TSymCreate:
			err = p.parseCreate()
			if err != nil {
				return nil, err
			}

		case TSymDrop:
			err = p.parseDrop()
			if err != nil {
				return nil, err
			}

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

	p.global.Declare(name, typ, nil)

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
	case TSymDatetime:
		return types.Date, nil
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

	_, err = p.optional('=')
	if err != nil {
		return err
	}

	// Value to set.
	expr, err := p.parseExpr()
	if err != nil {
		return err
	}

	_, err = p.optional(';')
	if err != nil {
		return err
	}

	q := NewQuery(p.global)
	err = expr.Bind(q)
	if err != nil {
		return err
	}
	v, err := expr.Eval(nil, nil)
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
	v, err := expr.Eval(nil, nil)
	if err != nil {
		return err
	}
	fmt.Fprintf(p.output, "%s\n", v)
	return nil
}

func (p *Parser) parseSelect() (*Query, error) {
	q := NewQuery(p.global)

	// Columns. The columns list is empty for "SELECT *" queries.
	t, err := p.get()
	if err != nil {
		return nil, err
	}
	if t.Type != '*' {
		p.lexer.unget(t)
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
	}

	// INTO
	t, err = p.get()
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
		err = q.Global.Declare(t.StrVal, types.Table, nil)
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

	// FROM
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

	// WHERE
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

	// GROUP BY
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

	// ORDER BY
	t, err = p.get()
	if err != nil {
		return nil, err
	}
	if t.Type == TSymOrder {
		q.OrderBy, err = p.parseOrderBy()
		if err != nil {
			return nil, err
		}
	} else {
		p.lexer.unget(t)
	}

	// LIMIT
	t, err = p.get()
	if err != nil {
		return nil, err
	}
	if t.Type == TSymLimit {
		q.LimitFrom, q.Limit, err = p.parseLimit()
		if err != nil {
			return nil, err
		}
	} else {
		p.lexer.unget(t)
	}

	// Terminator.
	if p.nesting == 1 {
		_, err = p.optional(';')
	} else {
		_, err = p.need(')')
	}
	if err != nil {
		return nil, err
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
		var url []string

		switch t.Type {
		case TIdentifier:
			b := q.Global.Get(t.StrVal)
			if b == nil {
				return nil, p.errf(t.From, "unknown identifier '%s'", t.StrVal)
			}
			if b.Value == types.Null {
				return nil, p.errf(t.From, "identifier '%s' unset", t.StrVal)
			}
			switch b.Type {
			case types.String:
				url = append(url, b.Value.String())

			case types.Table:
				table, ok := b.Value.(types.TableValue)
				if !ok {
					return nil, p.errf(t.From,
						"invalid table value for identifier '%s'", t.StrVal)
				}
				source = table.Source
				// Use the symbol name as the default alias. The 'AS'
				// below can override this.
				as = t.StrVal

			case types.Array:
				av, ok := b.Value.(types.ArrayValue)
				if !ok {
					return nil, p.errf(t.From, "invalid array: %s", b.Value)
				}
				for _, a := range av.Data {
					url = append(url, a.String())
				}

			default:
				return nil, p.errf(t.From, "invalid source type: %s", b.Type)
			}

		case TString:
			url = append(url, t.StrVal)
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

	// Collect all referenced columns for the source.
	seen := make(map[string]bool)
	for _, col := range columns {
		var filtered []types.Reference

		for _, ref := range col.Expr.References() {
			if ref.Source == source {
				if !seen[ref.Column] {
					filtered = append(filtered, ref)
					seen[ref.Column] = true
				}
			}
		}

		for _, ref := range filtered {
			result = append(result, types.ColumnSelector{
				Name: ref,
			})
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
	_, err := p.need(TSymBy)
	if err != nil {
		return nil, err
	}
	var result []Expr
	for {
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		result = append(result, expr)

		t, err := p.get()
		if err != nil {
			return nil, err
		}
		if t.Type != ',' {
			p.lexer.unget(t)
			return result, nil
		}
	}
}

func (p *Parser) parseOrderBy() ([]Order, error) {
	_, err := p.need(TSymBy)
	if err != nil {
		return nil, err
	}
	var result []Order
	for {
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		t, err := p.get()
		if err != nil {
			return nil, err
		}
		var desc bool
		if t.Type == TSymAsc {
			desc = false
		} else if t.Type == TSymDesc {
			desc = true
		} else {
			p.lexer.unget(t)
		}
		result = append(result, Order{
			Expr: expr,
			Desc: desc,
		})

		t, err = p.get()
		if err != nil {
			return nil, err
		}
		if t.Type != ',' {
			p.lexer.unget(t)
			return result, nil
		}
	}
}

func (p *Parser) parseLimit() (uint32, uint32, error) {
	// LIMIT from [, to]
	lim1, err := p.need(TInt)
	if err != nil {
		return 0, 0, err
	}
	if lim1.IntVal < 0 || lim1.IntVal > math.MaxUint32 {
		return 0, 0, fmt.Errorf("invalid limit: %d", lim1.IntVal)
	}
	v64 := lim1.IntVal
	var i1 uint32
	if v64 > math.MaxUint32 {
		i1 = math.MaxUint32
	} else {
		i1 = uint32(v64)
	}
	t, err := p.get()
	if err != nil {
		return 0, 0, err
	}
	if t.Type != ',' {
		p.lexer.unget(t)
		return 0, i1, nil
	}
	lim2, err := p.need(TInt)
	if err != nil {
		return 0, 0, err
	}
	if lim2.IntVal < 0 || lim2.IntVal > math.MaxUint32 {
		return 0, 0, fmt.Errorf("invalid limit: %d", lim2.IntVal)
	}
	v64 = lim2.IntVal
	var i2 uint32
	if v64 > math.MaxUint32 {
		i2 = math.MaxUint32
	} else {
		i2 = uint32(v64)
	}
	return i1, i2, nil
}

func (p *Parser) parseCreate() error {
	t, err := p.get()
	if err != nil {
		return err
	}
	switch t.Type {
	case TSymFunction:
		return p.parseCreateFunction()

	default:
		return p.errUnexpected(t)
	}
}

func (p *Parser) parseCreateFunction() error {
	t, err := p.need(TIdentifier)
	if err != nil {
		return err
	}
	name := strings.ToUpper(t.StrVal)
	var args []FunctionArg

	t, err = p.need('(')
	if err != nil {
		return err
	}
	t, err = p.get()
	if err != nil {
		return err
	}
	if t.Type != ')' {
		// Function arguments.
		p.lexer.unget(t)
		for {
			t, err = p.need(TIdentifier)
			if err != nil {
				return err
			}
			argName := strings.ToUpper(t.StrVal)

			argType, err := p.parseType()
			if err != nil {
				return err
			}
			args = append(args, FunctionArg{
				Name: argName,
				Type: argType,
			})

			t, err = p.get()
			if err != nil {
				return err
			}
			if t.Type == ')' {
				break
			} else if t.Type != ',' {
				return p.errUnexpected(t)
			}
		}
	}
	_, err = p.need(TSymReturns)
	if err != nil {
		return err
	}
	retType, err := p.parseType()
	if err != nil {
		return err
	}
	_, err = p.optional(TSymAs)
	if err != nil {
		return err
	}
	_, err = p.need(TSymBegin)
	if err != nil {
		return err
	}

	var ret Expr
	for {
		t, err = p.get()
		if err != nil {
			return err
		}
		if t.Type == TSymReturn {
			ret, err = p.parseExpr()
			if err != nil {
				return err
			}
			_, err = p.optional(';')
			if err != nil {
				return err
			}
			break
		}

		p.lexer.unget(t)
		_, err = p.parseStmt()
		if err != nil {
			return err
		}
	}
	_, err = p.need(TSymEnd)
	if err != nil {
		return err
	}
	_, err = p.optional(';')
	if err != nil {
		return err
	}

	return createFunction(&Function{
		Name:         name,
		Args:         args,
		RetType:      retType,
		Ret:          ret,
		MinArgs:      len(args),
		MaxArgs:      len(args),
		IsIdempotent: idempotentFalse,
	})
}

func (p *Parser) parseDrop() error {
	t, err := p.get()
	if err != nil {
		return err
	}
	switch t.Type {
	case TSymFunction:
		return p.parseDropFunction()

	default:
		return p.errUnexpected(t)
	}
}

func (p *Parser) parseDropFunction() error {
	var ifExists bool

	t, err := p.get()
	if err != nil {
		return err
	}
	if t.Type == TSymIf {
		_, err = p.need(TSymExists)
		if err != nil {
			return err
		}
		ifExists = true
	} else {
		p.lexer.unget(t)
	}

	t, err = p.need(TIdentifier)
	if err != nil {
		return err
	}
	name := strings.ToUpper(t.StrVal)

	_, err = p.optional(';')
	if err != nil {
		return err
	}

	return dropFunction(name, ifExists)
}

func (p *Parser) parseStmt() (int, error) {
	return 0, fmt.Errorf("parseStmt not implemented yet")
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
	t, err := p.get()
	if err != nil {
		return nil, err
	}

	var ut UnaryType

	switch t.Type {
	case '-':
		ut = UnaryMinus

	default:
		p.lexer.unget(t)
		return p.parseExprPostfix()
	}

	expr, err := p.parseExprPostfix()
	if err != nil {
		return nil, err
	}
	return &Unary{
		Type: ut,
		Expr: expr,
	}, nil
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
		}
		if n.Type == '(' {
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

	case TSymCast:
		_, err = p.need('(')
		if err != nil {
			return nil, err
		}
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		_, err = p.need(TSymAs)
		if err != nil {
			return nil, err
		}
		castType, err := p.parseType()
		if err != nil {
			return nil, err
		}
		_, err = p.need(')')
		if err != nil {
			return nil, err
		}
		return &Cast{
			Expr: expr,
			Type: castType,
		}, nil

	case TSymCase:
		return p.parseCase()

	case TString:
		val = types.StringValue(t.StrVal)
	case TInt:
		val = types.IntValue(t.IntVal)
	case TFloat:
		val = types.FloatValue(t.FloatVal)
	case TBool:
		val = types.BoolValue(t.BoolVal)
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
	call := &Call{
		Name:      strings.ToUpper(name.StrVal),
		Arguments: args,
	}

	// Resolve function.
	call.Function = builtIn(call.Name)
	if call.Function == nil {
		return nil, fmt.Errorf("undefined function: %s", call.Name)
	}

	return call, nil
}

func (p *Parser) parseCase() (Expr, error) {
	caseExpr := new(Case)

	t, err := p.get()
	if err != nil {
		return nil, err
	}
	if t.Type != TSymWhen {
		p.lexer.unget(t)
		caseExpr.Input, err = p.parseExpr()
		if err != nil {
			return nil, err
		}
		_, err = p.need(TSymWhen)
		if err != nil {
			return nil, err
		}
	}

	// Parse when branches.
	for {
		when, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		_, err = p.need(TSymThen)
		if err != nil {
			return nil, err
		}
		then, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		caseExpr.Branches = append(caseExpr.Branches, Branch{
			When: when,
			Then: then,
		})
		t, err = p.get()
		if err != nil {
			return nil, err
		}
		if t.Type != TSymWhen {
			p.lexer.unget(t)
			break
		}
	}

	t, err = p.get()
	if err != nil {
		return nil, err
	}
	if t.Type == TSymElse {
		caseExpr.Else, err = p.parseExpr()
		if err != nil {
			return nil, err
		}
		t, err = p.get()
		if err != nil {
			return nil, err
		}
	}
	if t.Type != TSymEnd {
		return nil, p.errUnexpected(t)
	}

	return caseExpr, nil
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
