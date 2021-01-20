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
)

// Client implements the IQL client.
type Client struct {
	global *lang.Scope
	parser *lang.Parser
}

// NewClient creates a new IQL client.
func NewClient() *Client {
	global := lang.NewScope(nil)
	lang.InitSystemVariables(global)

	return &Client{
		global: global,
	}
}

// SetInput sets the IQL input to parse.
func (c *Client) SetInput(input io.Reader, source string) error {
	c.parser = lang.NewParser(c.global, input, source)
	return nil
}

// Parse parses the IQL file.
func (c *Client) Parse() (*lang.Query, error) {
	return c.parser.Parse()
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
