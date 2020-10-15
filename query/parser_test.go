//
// Copyright (c) 2019 Markku Rossi
//
// All rights reserved.
//

package query

import (
	"bytes"
	"os"
	"testing"

	"github.com/markkurossi/iql/data"
	"github.com/markkurossi/tabulate"
)

var parserTests = []string{
	`select 1 + 0x01 + 0b10 + 077 + 0o70 as Sum, 100-42 as Diff`,
}

func TestParser(t *testing.T) {
	for _, input := range parserTests {
		q, err := Parse(bytes.NewReader([]byte(input)), "{data}")
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		rows, err := q.Get()
		if err != nil {
			t.Fatalf("q.Get failed: %v", err)
		}
		tab := data.Table(q, tabulate.Unicode)
		for _, columns := range rows {
			row := tab.Row()
			for _, col := range columns {
				row.Column(col.String())
			}
		}

		tab.Print(os.Stdout)
	}
}
