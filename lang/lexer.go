//
// Copyright (c) 2020-2021 Markku Rossi
//
// All rights reserved.
//

package lang

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/markkurossi/iql/types"
)

// Point specifies a position in the parser input data.
type Point struct {
	Source string
	Line   int
	Col    int
}

func (p Point) String() string {
	return fmt.Sprintf("%s:%d:%d", p.Source, p.Line, p.Col)
}

// TokenType specifies input token types.
type TokenType int

// Know token types.
const (
	TIdentifier TokenType = iota + 256
	TString
	TInt
	TFloat
	TBool
	TNull
	TSymSelect
	TSymInto
	TSymNot
	TSymIn
	TSymFrom
	TSymWhere
	TSymGroup
	TSymOrder
	TSymAs
	TSymBy
	TSymAsc
	TSymDesc
	TSymFilter
	TSymDeclare
	TSymPrint
	TSymSet
	TSymBoolean
	TSymInteger
	TSymReal
	TSymDatetime
	TSymVarchar
	TSymCast
	TSymCase
	TSymWhen
	TSymThen
	TSymElse
	TSymBegin
	TSymEnd
	TSymCreate
	TSymFunction
	TSymReturns
	TSymReturn
	TSymDrop
	TSymIf
	TSymExists
	TSymLimit
	TAnd
	TOr
	TNEq
	TNMatch
	TLe
	TGe
)

var tokenTypes = map[TokenType]string{
	TIdentifier:  "identifier",
	TString:      "string",
	TInt:         "int",
	TFloat:       "float",
	TNull:        "NULL",
	TSymSelect:   "SELECT",
	TSymInto:     "INTO",
	TSymNot:      "NOT",
	TSymIn:       "IN",
	TSymFrom:     "FROM",
	TSymWhere:    "WHERE",
	TSymGroup:    "GROUP",
	TSymOrder:    "ORDER",
	TSymAs:       "AS",
	TSymBy:       "BY",
	TSymAsc:      "ASC",
	TSymDesc:     "DESC",
	TSymFilter:   "FILTER",
	TSymDeclare:  "DECLARE",
	TSymPrint:    "PRINT",
	TSymSet:      "SET",
	TSymBoolean:  "BOOLEAN",
	TSymInteger:  "INTEGER",
	TSymReal:     "REAL",
	TSymDatetime: "DATETIME",
	TSymVarchar:  "VARCHAR",
	TSymCast:     "CAST",
	TSymCase:     "CASE",
	TSymWhen:     "WHEN",
	TSymThen:     "THEN",
	TSymElse:     "ELSE",
	TSymBegin:    "BEGIN",
	TSymEnd:      "END",
	TSymCreate:   "CREATE",
	TSymFunction: "FUNCTION",
	TSymReturns:  "RETURNS",
	TSymReturn:   "RETURN",
	TSymDrop:     "DROP",
	TSymIf:       "IF",
	TSymExists:   "EXISTS",
	TSymLimit:    "LIMIT",
	TAnd:         "AND",
	TOr:          "OR",
	TNEq:         "<>",
	TNMatch:      "!~",
	TLe:          "<=",
	TGe:          ">=",
}

func (t TokenType) String() string {
	name, ok := tokenTypes[t]
	if ok {
		return name
	}
	if t < TIdentifier {
		return fmt.Sprintf("%c", rune(t))
	}
	return fmt.Sprintf("{TokenType %d}", t)
}

var symbols = map[string]TokenType{
	"NULL":     TNull,
	"SELECT":   TSymSelect,
	"INTO":     TSymInto,
	"NOT":      TSymNot,
	"IN":       TSymIn,
	"FROM":     TSymFrom,
	"WHERE":    TSymWhere,
	"GROUP":    TSymGroup,
	"ORDER":    TSymOrder,
	"AS":       TSymAs,
	"BY":       TSymBy,
	"ASC":      TSymAsc,
	"DESC":     TSymDesc,
	"FILTER":   TSymFilter,
	"DECLARE":  TSymDeclare,
	"PRINT":    TSymPrint,
	"SET":      TSymSet,
	"BOOLEAN":  TSymBoolean,
	"INTEGER":  TSymInteger,
	"REAL":     TSymReal,
	"DATETIME": TSymDatetime,
	"VARCHAR":  TSymVarchar,
	"CAST":     TSymCast,
	"CASE":     TSymCase,
	"WHEN":     TSymWhen,
	"THEN":     TSymThen,
	"ELSE":     TSymElse,
	"BEGIN":    TSymBegin,
	"END":      TSymEnd,
	"CREATE":   TSymCreate,
	"FUNCTION": TSymFunction,
	"RETURNS":  TSymReturns,
	"RETURN":   TSymReturn,
	"DROP":     TSymDrop,
	"IF":       TSymIf,
	"EXISTS":   TSymExists,
	"LIMIT":    TSymLimit,
	"AND":      TAnd,
	"OR":       TOr,
}

// Token implements an input token.
type Token struct {
	Type     TokenType
	From     Point
	To       Point
	StrVal   string
	IntVal   int64
	FloatVal float64
	BoolVal  bool
}

func (t *Token) String() string {
	switch t.Type {
	case TIdentifier:
		return t.StrVal
	case TString:
		return fmt.Sprintf("'%s'", t.StrVal)
	case TInt:
		return fmt.Sprintf("%d", t.IntVal)
	case TFloat:
		return fmt.Sprintf("%f", t.FloatVal)
	default:
		return t.Type.String()
	}
}

type lexer struct {
	in               *bufio.Reader
	trailingInjected bool
	point            Point
	tokenStart       Point
	ungot            *Token
	unread           bool
	unreadRune       rune
	unreadSize       int
	unreadPoint      Point
	history          map[int][]rune
}

func newLexer(input io.Reader, source string) *lexer {
	return &lexer{
		in: bufio.NewReader(input),
		point: Point{
			Source: source,
			Line:   1,
			Col:    0,
		},
		history: make(map[int][]rune),
	}
}

// ReadRune reads the next input rune.
func (l *lexer) ReadRune() (rune, int, error) {
	if l.unread {
		l.point, l.unreadPoint = l.unreadPoint, l.point
		l.unread = false
		return l.unreadRune, l.unreadSize, nil
	}
	r, size, err := l.in.ReadRune()
	if err != nil {
		if err == io.EOF && !l.trailingInjected {
			r = ';'
			size = 1
			l.trailingInjected = true
		} else {
			return r, size, err
		}
	}

	l.unreadRune = r
	l.unreadSize = size
	l.unreadPoint = l.point
	if r == '\n' {
		l.point.Line++
		l.point.Col = 0
	} else {
		l.point.Col++
		l.history[l.point.Line] = append(l.history[l.point.Line], r)
	}

	return r, size, nil
}

// UnreadRune unreads the last rune.
func (l *lexer) UnreadRune() error {
	l.point, l.unreadPoint = l.unreadPoint, l.point
	l.unread = true
	return nil
}

// FlushEOL discards all remaining input from the current source code
// line.
func (l *lexer) FlushEOL() error {
	for {
		r, _, err := l.ReadRune()
		if err != nil {
			if err != io.EOF {
				return err
			}
			return nil
		}
		if r == '\n' {
			return nil
		}
	}
}

func (l *lexer) get() (*Token, error) {
	if l.ungot != nil {
		token := l.ungot
		l.ungot = nil
		return token, nil
	}

lexer:
	for {
		l.tokenStart = l.point
		r, _, err := l.ReadRune()
		if err != nil {
			return nil, err
		}
		if unicode.IsSpace(r) {
			continue
		}

		switch r {
		case '+', '*', '~', '%', '=', '.', ',', '(', ')', ';':
			return l.token(TokenType(r)), nil

		case '<':
			r, _, err := l.ReadRune()
			if err != nil {
				if err != io.EOF {
					return nil, err
				}
				return l.token(TokenType('<')), nil
			}
			switch r {
			case '>':
				return l.token(TNEq), nil
			case '=':
				return l.token(TLe), nil
			default:
				l.UnreadRune()
				return l.token(TokenType('<')), nil
			}

		case '>':
			r, _, err := l.ReadRune()
			if err != nil {
				if err != io.EOF {
					return nil, err
				}
				return l.token(TokenType('>')), nil
			}
			switch r {
			case '=':
				return l.token(TGe), nil
			default:
				l.UnreadRune()
				return l.token(TokenType('>')), nil
			}

		case '!':
			r, _, err := l.ReadRune()
			if err != nil {
				return nil, err
			}
			switch r {
			case '~':
				return l.token(TNMatch), nil
			default:
				l.UnreadRune()
				return nil, fmt.Errorf("unexpected character '%s'",
					string(r))
			}

		case '-':
			r, _, err := l.ReadRune()
			if err != nil {
				if err != io.EOF {
					return nil, err
				}
				return l.token(TokenType('-')), nil
			}
			if r == '-' {
				// Single line comment: -- discard to EOL.
				l.FlushEOL()
				continue lexer
			}
			l.UnreadRune()
			return l.token(TokenType('-')), nil

		case '/':
			r, _, err := l.ReadRune()
			if err != nil {
				if err != io.EOF {
					return nil, err
				}
				return l.token(TokenType('/')), nil
			}
			if r == '*' {
				// C-style comment: discard until */
				for {
					r, _, err := l.ReadRune()
					if err != nil {
						return nil, err
					}
					if r == '*' {
						r, _, err := l.ReadRune()
						if err != nil {
							return nil, err
						}
						if r == '/' {
							continue lexer
						}
					}
				}
			}
			l.UnreadRune()
			return l.token(TokenType('/')), nil

		case '\'':
			var sb strings.Builder
			for {
				r, _, err := l.ReadRune()
				if err != nil {
					return nil, err
				}
				if r == '\'' {
					r, _, err := l.ReadRune()
					if err != nil {
						if err != io.EOF {
							return nil, err
						}
						break
					}
					if r != '\'' {
						l.UnreadRune()
						break
					}
				}
				sb.WriteRune(r)
			}
			token := l.token(TString)
			token.StrVal = sb.String()
			return token, nil

		case '`':
			for i := 0; i < 2; i++ {
				r, _, err = l.ReadRune()
				if err != nil {
					return nil, err
				}
				if r != '`' {
					l.UnreadRune()
					return nil, fmt.Errorf("unexpected character '%s'",
						string(r))
				}
			}
			return l.readHereString()

		case '"', '[':
			end := r
			if r == '[' {
				end = ']'
			}
			var runes []rune
			for {
				r, _, err := l.ReadRune()
				if err != nil {
					return nil, err
				}
				if r == end {
					r, _, err := l.ReadRune()
					if err != nil {
						if err != io.EOF {
							return nil, err
						}
						break
					}
					if r != end {
						l.UnreadRune()
						break
					}
				}
				runes = append(runes, r)
			}
			token := l.token(TIdentifier)
			token.StrVal = string(runes)
			return token, nil

		case '0':
			var i64 int64

			r, _, err = l.ReadRune()
			if err != nil {
				if err != io.EOF {
					return nil, err
				}
			} else {
				switch r {
				case 'b', 'B':
					i64, err = l.readBinaryLiteral([]rune{'0', r})
				case 'o', 'O':
					i64, err = l.readOctalLiteral([]rune{'0', r})
				case 'x', 'X':
					i64, err = l.readHexLiteral([]rune{'0', r})
				case '0', '1', '2', '3', '4', '5', '6', '7':
					i64, err = l.readOctalLiteral([]rune{'0', r})
				case '.':
					f64, err := l.readFloatLiteral([]rune{'0', r})
					if err != nil {
						return nil, err
					}
					token := l.token(TFloat)
					token.FloatVal = f64
					return token, nil
				default:
					l.UnreadRune()
				}
				if err != nil {
					return nil, err
				}
			}
			token := l.token(TInt)
			token.IntVal = i64
			return token, nil

		default:
			if unicode.IsLetter(r) {
				identifier := string(r)
				for {
					r, _, err := l.ReadRune()
					if err != nil {
						if err != io.EOF {
							return nil, err
						}
						break
					}
					if !unicode.IsLetter(r) && !unicode.IsDigit(r) &&
						r != '_' && r != '$' {
						l.UnreadRune()
						break
					}
					identifier += string(r)
				}
				sym, ok := symbols[strings.ToUpper(identifier)]
				if ok {
					return l.token(sym), nil
				}
				bval, ok := types.ParseBoolean(identifier)
				if ok {
					token := l.token(TBool)
					token.BoolVal = bval
					return token, nil
				}
				token := l.token(TIdentifier)
				token.StrVal = identifier
				return token, nil
			}
			if unicode.IsDigit(r) {
				val := []rune{r}
				for {
					r, _, err := l.ReadRune()
					if err != nil {
						if err != io.EOF {
							return nil, err
						}
						break
					}
					if unicode.IsDigit(r) {
						val = append(val, r)
					} else if r == '.' {
						f64, err := l.readFloatLiteral(append(val, r))
						if err != nil {
							return nil, err
						}
						token := l.token(TFloat)
						token.FloatVal = f64
						return token, nil
					} else {
						l.UnreadRune()
						break
					}
				}
				i64, err := strconv.ParseInt(string(val), 10, 64)
				if err != nil {
					return nil, err
				}
				token := l.token(TInt)
				token.IntVal = i64
				return token, nil
			}
			l.UnreadRune()
			return nil, fmt.Errorf("unexpected character '%s'", string(r))
		}
	}
}

func (l *lexer) readHereString() (*Token, error) {
	var sb strings.Builder

	// Header line until the first newline.
	for {
		r, _, err := l.ReadRune()
		if err != nil {
			return nil, err
		}
		if r == '\n' {
			break
		}
		sb.WriteRune(r)
	}
	options := strings.Split(sb.String(), " ")

	sb.Reset()
	for {
		r, _, err := l.ReadRune()
		if err != nil {
			return nil, err
		}
		if r == '`' {
			count := 1
			for ; count < 3; count++ {
				r, _, err = l.ReadRune()
				if err != nil {
					return nil, err
				}
				if r != '`' {
					break
				}
			}
			if count >= 3 {
				break
			}
			for ; count > 0; count-- {
				sb.WriteRune('`')
			}
		}
		sb.WriteRune(r)
	}

	val := sb.String()

	for _, option := range options {
		if len(option) == 0 {
			continue
		}
		parts := strings.Split(option, ":")
		switch len(parts) {
		case 2:
			switch parts[0] {
			case "datauri":
				val = fmt.Sprintf("data:%s;base64,%s", parts[1],
					base64.StdEncoding.EncodeToString([]byte(val)))
			default:
				return nil, fmt.Errorf("Unknown here option: %s", option)
			}

		default:
			return nil, fmt.Errorf("unknown here option: %s", option)
		}
	}

	token := l.token(TString)
	token.StrVal = val
	return token, nil
}

func (l *lexer) readBinaryLiteral(val []rune) (int64, error) {
loop:
	for {
		r, _, err := l.ReadRune()
		if err != nil {
			if err != io.EOF {
				return 0, err
			}
			break
		}
		switch r {
		case '0', '1':
			val = append(val, r)
		default:
			l.UnreadRune()
			break loop
		}
	}
	return strconv.ParseInt(string(val), 0, 64)
}

func (l *lexer) readOctalLiteral(val []rune) (int64, error) {
loop:
	for {
		r, _, err := l.ReadRune()
		if err != nil {
			if err != io.EOF {
				return 0, err
			}
			break
		}
		switch r {
		case '0', '1', '2', '3', '4', '5', '6', '7':
			val = append(val, r)
		default:
			l.UnreadRune()
			break loop
		}
	}
	return strconv.ParseInt(string(val), 0, 64)
}

func (l *lexer) readHexLiteral(val []rune) (int64, error) {
	for {
		r, _, err := l.ReadRune()
		if err != nil {
			if err != io.EOF {
				return 0, err
			}
			break
		}
		if unicode.Is(unicode.Hex_Digit, r) {
			val = append(val, r)
		} else {
			l.UnreadRune()
			break
		}
	}
	return strconv.ParseInt(string(val), 0, 64)
}

func (l *lexer) readFloatLiteral(val []rune) (float64, error) {
	for {
		r, _, err := l.ReadRune()
		if err != nil {
			if err != io.EOF {
				return 0, err
			}
			break
		}
		if unicode.IsDigit(r) {
			val = append(val, r)
		} else {
			l.UnreadRune()
			break
		}
	}
	return strconv.ParseFloat(string(val), 64)
}

func (l *lexer) unget(t *Token) {
	l.ungot = t
}

func (l *lexer) token(t TokenType) *Token {
	return &Token{
		Type: t,
		From: l.tokenStart,
		To:   l.point,
	}
}
