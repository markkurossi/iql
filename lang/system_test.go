//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package lang

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"testing"
)

var systemTests = []struct {
	q    string
	v    [][]string
	rest [][][]string
}{
	{
		q: `
SET REALFMT = '%.2f';
SELECT 3.1415;`,
		v: [][]string{
			{"3.14"},
		},
	},
	{
		q: `
SET TERMOUT OFF
SELECT 'Hello, world!';`,
		v: [][]string{
			{"Hello, world!"},
		},
	},
}

func TestSystem(t *testing.T) {
	data := fmt.Sprintf("data:text/csv;base64,%s",
		base64.StdEncoding.EncodeToString([]byte(builtInData)))

	for testID, input := range systemTests {
		name := fmt.Sprintf("Test %d", testID)
		global := NewScope(nil)
		InitSystemVariables(global)
		parser := NewParser(global, bytes.NewReader([]byte(input.q)), name,
			os.Stdout)

		parser.SetString("data", data)

		for {
			q, err := parser.Parse()
			if err != nil {
				if err == io.EOF {
					break
				}
				t.Fatalf("%s: parse failed: %v", name, err)
			}
			verifyResult(t, name, input.q, q, input.v)
		}
	}
}
