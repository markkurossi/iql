//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package query

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
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
	TInteger
	TNull
	TSymSelect
	TSymFrom
	TSymWhere
	TSymAs
	TSymFilter
	TAnd
	TOr
	TNeq
	TLe
	TGe
)

var tokenTypes = map[TokenType]string{
	TIdentifier: "identifier",
	TString:     "string",
	TInteger:    "integer",
	TNull:       "NULL",
	TSymSelect:  "SELECT",
	TSymFrom:    "FROM",
	TSymWhere:   "WHERE",
	TSymAs:      "AS",
	TSymFilter:  "FILTER",
	TAnd:        "AND",
	TOr:         "OR",
	TNeq:        "<>",
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
	"NULL":   TNull,
	"SELECT": TSymSelect,
	"FROM":   TSymFrom,
	"WHERE":  TSymWhere,
	"AS":     TSymAs,
	"FILTER": TSymFilter,
	"AND":    TAnd,
	"OR":     TOr,
}

// Token implements an input token.
type Token struct {
	Type   TokenType
	From   Point
	To     Point
	StrVal string
	IntVal int64
}

func (t *Token) String() string {
	switch t.Type {
	case TIdentifier:
		return t.StrVal
	case TString:
		return fmt.Sprintf("'%s'", t.StrVal)
	case TInteger:
		return fmt.Sprintf("%d", t.IntVal)
	default:
		return t.Type.String()
	}
}

type lexer struct {
	in          *bufio.Reader
	point       Point
	tokenStart  Point
	ungot       *Token
	unread      bool
	unreadRune  rune
	unreadSize  int
	unreadPoint Point
	history     map[int][]rune
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
		return r, size, err
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
				return l.token(TNeq), nil
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
				return l.token(TokenType('<')), nil
			}
			switch r {
			case '=':
				return l.token(TGe), nil
			default:
				l.UnreadRune()
				return l.token(TokenType('>')), nil
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
			var runes []rune
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
				runes = append(runes, r)
			}
			token := l.token(TString)
			token.StrVal = string(runes)
			return token, nil

		case '"':
			var runes []rune
			for {
				r, _, err := l.ReadRune()
				if err != nil {
					return nil, err
				}
				if r == '"' {
					r, _, err := l.ReadRune()
					if err != nil {
						if err != io.EOF {
							return nil, err
						}
						break
					}
					if r != '"' {
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
				default:
					l.UnreadRune()
				}
				if err != nil {
					return nil, err
				}
			}
			token := l.token(TInteger)
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
					} else {
						l.UnreadRune()
						break
					}
				}
				i64, err := strconv.ParseInt(string(val), 10, 64)
				if err != nil {
					return nil, err
				}
				token := l.token(TInteger)
				token.IntVal = i64
				return token, nil
			}
			return nil, fmt.Errorf("%s: unexpected character '%s'",
				l.point, string(r))
		}
	}
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
