//
// Copyright (c) 2021 Markku Rossi
//
// All rights reserved.
//

package iql

import (
	"fmt"
	"io"

	"github.com/markkurossi/iql/lang"
	"github.com/markkurossi/iql/types"
	"github.com/markkurossi/tabulate"
)

// Client implements the IQL client.
type Client struct {
	global *lang.Scope
	out    io.Writer
}

// NewClient creates a new IQL client.
func NewClient(out io.Writer) *Client {
	global := lang.NewScope(nil)
	lang.InitSystemVariables(global)

	return &Client{
		global: global,
		out:    out,
	}
}

// SetString assigns the string value to the global variable. The
// global variable must have been declared and its type must be
// VARCHAR.
func (c *Client) SetString(name, value string) error {
	return c.global.Set(name, types.StringValue(value))
}

// SetStringArray assings the string array value to the global
// variable. The global variable must have been declared and its type
// must be []VARCHAR.
func (c *Client) SetStringArray(name string, value []string) error {
	var arr []types.Value
	for _, v := range value {
		arr = append(arr, types.StringValue(v))
	}
	return c.global.Set(name, types.NewArray(types.String, arr))
}

// Write implements io.Write().
func (c *Client) Write(p []byte) (n int, err error) {
	if c.SysTermOut() {
		return c.out.Write(p)
	}
	return len(p), nil
}

// Parse parses the IQL file.
func (c *Client) Parse(input io.Reader, source string) error {
	parser := lang.NewParser(c.global, input, source, c)
	for {
		q, err := parser.Parse()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		tab, err := types.Tabulate(q, c.SysTableFmt())
		if err != nil {
			return err
		}
		tab.Print(c)
	}
}

// SysTableFmt returns the table formatting style.
func (c *Client) SysTableFmt() (style tabulate.Style) {
	style = tabulate.Unicode
	b := c.global.Get(lang.SysTableFmt)
	if b == nil {
		return
	}
	s, ok := tabulate.Styles[b.Value.String()]
	if ok {
		style = s
	}
	return
}

// SysTermOut describes if terminal output is enabled.
func (c *Client) SysTermOut() bool {
	b := c.global.Get(lang.SysTermOut)
	if b == nil {
		panic("system variable TERMOUT not set")
	}
	v, err := b.Value.Bool()
	if err != nil {
		panic(fmt.Sprintf("invalid system variable value: %s", err))
	}
	return v
}
