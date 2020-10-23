//
// Copyright (c) 2019 Markku Rossi
//
// All rights reserved.
//

package query

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/markkurossi/iql/types"
	"github.com/markkurossi/tabulate"
)

type BuiltInTest struct {
	q string
	v string
}

var builtInData = `1970,100,100.5
1971,200,200.5
1972,300,300.5
1973,400,400.5
1974,500,500.5`

var builtInTests = []BuiltInTest{
	{
		q: `
select AVG(Year)
from (
      select "0" as Year,
             "1" as IVal,
             "2" as FVal
      from data
     );`,
		v: "1972",
	},
	{
		q: `
select COUNT(Year)
from (
      select "0" as Year,
             "1" as IVal,
             "2" as FVal
      from data
     );`,
		v: "5",
	},
	{
		q: `
select MAX(Year)
from (
      select "0" as Year,
             "1" as IVal,
             "2" as FVal
      from data
     );`,
		v: "1974",
	},
	{
		q: `
select MIN(Year)
from (
      select "0" as Year,
             "1" as IVal,
             "2" as FVal
      from data
     );`,
		v: "1970",
	},
	{
		q: `
select SUM(Year)
from (
      select "0" as Year,
             "1" as IVal,
             "2" as FVal
      from data
     );`,
		v: "9860",
	},
	{
		q: `
SELECT NULLIF(4, 4);`,
		v: "NULL",
	},
	{
		q: `
SELECT NULLIF(5, 4);`,
		v: "5",
	},
	{
		q: `
SELECT 5 / NULLIF(0.0, 0.0);`,
		v: "NULL",
	},
	{
		q: `
SELECT 5 / NULLIF(5.0, 0.0);`,
		v: "1.00",
	},
}

func TestBuiltIn(t *testing.T) {
	data := fmt.Sprintf("data:text/csv;base64,%s",
		base64.StdEncoding.EncodeToString([]byte(builtInData)))

	for idx, input := range builtInTests {
		name := fmt.Sprintf("Test %d", idx)
		parser := NewParser(bytes.NewReader([]byte(input.q)), name)

		parser.SetString("data", data)

		for {
			q, err := parser.Parse()
			if err != nil {
				if err == io.EOF {
					break
				}
				t.Fatalf("%s: parse failed: %v", name, err)
			}

			rows, err := q.Get()
			if err != nil {
				t.Errorf("%s: q.Get failed: %v", name, err)
				continue
			}
			if len(rows) != 1 {
				t.Errorf("%s: unexpected number of result rows", name)
				printResult(q, rows)
				continue
			}
			if len(rows[0]) != 1 {
				t.Fatalf("%s: unexpected number of result columns", name)
				printResult(q, rows)
				continue
			}
			result := rows[0][0].String()
			if result != input.v {
				t.Errorf("%s: failed: got %s, expected %s\n",
					name, result, input.v)
				printResult(q, rows)

			}
		}
	}
}

func printResult(q types.Source, rows []types.Row) {

	tab, err := types.Tabulate(q, tabulate.Unicode)
	if err != nil {
		return
	}
	tab.Print(os.Stdout)
}
