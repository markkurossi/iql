//
// Copyright (c) 2019 Markku Rossi
//
// All rights reserved.
//

package query

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

var inputs = []string{
	`
SELECT ref.'.name'                      		AS Name,
       ref.':nth-child(5)'              		AS Price,
       ref.'.share'                     		AS Weigth,
       ref.a                            		AS link,
       portfolio.0                      		AS name
       portfolio.1                      		AS Count
       Count * Price                    		AS Invested
       Count * Price / SUM(Count * Price) * 100 AS 'My Weight'
FROM ',reference.html' FILTER 'tbody > tr' AS ref,
     ',portfolio.csv' AS portfolio
WHERE ref.link <> '' AND ref.Name = portfolio.name
`,
	`select 1 + 0x01 + 0b10 + 077 + 0o70`,
}

func TestLexer(t *testing.T) {
	for _, input := range inputs {
		lexer := newLexer(bytes.NewReader([]byte(input)), "{data}")
		for {
			token, err := lexer.get()
			if err != nil {
				if err == io.EOF {
					break
				}
				t.Fatalf("get failed: %v", err)
			}
			if false {
				fmt.Printf("%v ", token)
			}
		}
		fmt.Println()
	}
}
