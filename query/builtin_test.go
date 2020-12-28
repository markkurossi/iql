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

type IQLTest struct {
	q    string
	v    [][]string
	rest [][][]string
}

var builtInData = `1970,100,100.5
1971,200,200.5
1972,300,300.5
1973,400,400.5
1974,500,500.5`

var builtInTests = []IQLTest{
	{
		q: `
select AVG(Year)
from (
      select "0" as Year,
             "1" as IVal,
             "2" as FVal
      from data
     );`,
		v: [][]string{{"1972"}},
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
		v: [][]string{{"5"}},
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
		v: [][]string{{"1974"}},
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
		v: [][]string{{"1970"}},
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
		v: [][]string{{"9860"}},
	},
	{
		q: `
SELECT NULLIF(4, 4);`,
		v: [][]string{{"NULL"}},
	},
	{
		q: `
SELECT NULLIF(5, 4);`,
		v: [][]string{{"5"}},
	},
	{
		q: `
SELECT 5 / NULLIF(0.0, 0.0);`,
		v: [][]string{{"NULL"}},
	},
	{
		q: `
SELECT 5 / NULLIF(5.0, 0.0);`,
		v: [][]string{{"1"}},
	},

	// CAST tests.
	{
		q: `SELECT CAST(false AS BOOLEAN);`,
		v: [][]string{{"false"}},
	},
	{
		q: `SELECT CAST(false AS VARCHAR);`,
		v: [][]string{{"false"}},
	},
	{
		q: `SELECT CAST(5 AS INTEGER);`,
		v: [][]string{{"5"}},
	},
	{
		q: `SELECT CAST(5 AS REAL);`,
		v: [][]string{{"5"}},
	},
	{
		q: `SELECT CAST(5 AS VARCHAR);`,
		v: [][]string{{"5"}},
	},
	{
		q: `SELECT CAST(5.0 AS INTEGER);`,
		v: [][]string{{"5"}},
	},
	{
		q: `SELECT CAST(5.0 AS REAL);`,
		v: [][]string{{"5"}},
	},
	{
		q: `SELECT CAST(5.0 AS VARCHAR);`,
		v: [][]string{{"5"}},
	},
	{
		q: `SELECT CAST('5' AS INTEGER);`,
		v: [][]string{{"5"}},
	},
	{
		q: `SELECT CAST('5' AS REAL);`,
		v: [][]string{{"5"}},
	},
	{
		q: `SELECT CAST('5' AS VARCHAR);`,
		v: [][]string{{"5"}},
	},

	// String functions.
	{
		q: `SELECT CHARINDEX('Reflectors are vital safety' +
                             ' components of your bicycle.',
                             'bicycle');`,
		v: [][]string{{"48"}},
	},
	{
		q: `SELECT CHARINDEX('Reflectors are vital safety' +
                             ' components of your bicycle.',
                             'vital', 5);`,
		v: [][]string{{"16"}},
	},
	{
		q: `SELECT CHARINDEX('Reflectors are vital safety' +
                             ' components of your bicycle.',
                             'bike');`,
		v: [][]string{{"0"}},
	},
	{
		q: `SELECT BASE64ENC('foo');`,
		v: [][]string{{"Zm9v"}},
	},
	{
		q: `SELECT BASE64DEC('Zm9v');`,
		v: [][]string{{"foo"}},
	},
	{
		q: `SELECT LASTCHARINDEX('}abcd}def', '}');`,
		v: [][]string{{"6"}},
	},
	{
		q: `SELECT LASTCHARINDEX('}abcd}def', ',');`,
		v: [][]string{{"0"}},
	},
	{
		q: `SELECT LEFT('Hello, world!', 6);`,
		v: [][]string{{"Hello,"}},
	},
	{
		q: `SELECT LEFT('Hello', 6);`,
		v: [][]string{{"Hello"}},
	},
	{
		q: `SELECT LEN('Hello, world!');`,
		v: [][]string{{"13"}},
	},
	{
		q: `SELECT LOWER('Hello, world!');`,
		v: [][]string{{"hello, world!"}},
	},
	{
		q: `SELECT LTRIM('  Hello, World!  ');`,
		v: [][]string{{"Hello, World!  "}},
	},
	{
		q: `SELECT NCHAR(64);`,
		v: [][]string{{"@"}},
	},
	{
		q: `SELECT REVERSE('Ken');`,
		v: [][]string{{"neK"}},
	},
	{
		q: `SELECT REVERSE('Rob');`,
		v: [][]string{{"boR"}},
	},
	{
		q: `SELECT REVERSE(1234);`,
		v: [][]string{{"4321"}},
	},
	{
		q: `SELECT RIGHT('abcdefg', 0);`,
		v: [][]string{{""}},
	},
	{
		q: `SELECT RIGHT('abcdefg', 2);`,
		v: [][]string{{"fg"}},
	},
	{
		q: `SELECT RIGHT('abcdefg', 7);`,
		v: [][]string{{"abcdefg"}},
	},
	{
		q: `SELECT RIGHT('abcdefg', 100000);`,
		v: [][]string{{"abcdefg"}},
	},
	{
		q: `SELECT RTRIM('  Hello, World!  ');`,
		v: [][]string{{"  Hello, World!"}},
	},
	{
		q: `SELECT TRIM('  Hello, World!  ');`,
		v: [][]string{{"Hello, World!"}},
	},
	{
		q: `DECLARE nstring VARCHAR;
SET nstring = 'Åkergatan 24';
SELECT UNICODE(nstring), NCHAR(UNICODE(nstring));`,
		v: [][]string{{"197", "Å"}},
	},
	{
		q: `SELECT UPPER('Hello, world!');`,
		v: [][]string{{"HELLO, WORLD!"}},
	},

	// Datetime literals.
	{
		q: `SELECT YEAR('2010-04-30T01:01:01.1234567-07:00');`,
		v: [][]string{{"2010"}},
	},
	{
		q: `SELECT YEAR('2007-04-30 13:10:02.0474381');`,
		v: [][]string{{"2007"}},
	},
	{
		q: `SELECT YEAR('2007-04-30 13:10:02.0474381 -07:00');`,
		v: [][]string{{"2007"}},
	},
	{
		q: `SELECT YEAR('2007-04-30');`,
		v: [][]string{{"2007"}},
	},

	// Datetime functions.
	{
		q: `SELECT DATEDIFF(year,
                            '2005-12-31 23:59:59.9999999',
                            '2006-01-01 00:00:00.0000000');`,
		v: [][]string{{"1"}},
	},
	{
		q: `SELECT DATEDIFF(day,
                            '2005-12-31 23:59:59.9999999',
                            '2006-01-01 00:00:00.0000000');`,
		v: [][]string{{"1"}},
	},
	{
		q: `SELECT DATEDIFF(hour,
                            '2005-12-31 23:59:59.9999999',
                            '2006-01-01 00:00:00.0000000');`,
		v: [][]string{{"1"}},
	},
	{
		q: `SELECT DATEDIFF(minute,
                            '2005-12-31 23:59:59.9999999',
                            '2006-01-01 00:00:00.0000000');`,
		v: [][]string{{"1"}},
	},
	{
		q: `SELECT DATEDIFF(second,
                            '2005-12-31 23:59:59.9999999',
                            '2006-01-01 00:00:00.0000000');`,
		v: [][]string{{"1"}},
	},
	{
		q: `SELECT DATEDIFF(millisecond,
                            '2005-12-31 23:59:59.9999999',
                            '2006-01-01 00:00:00.0000000');`,
		v: [][]string{{"1"}},
	},
	{
		q: `SELECT DATEDIFF(microsecond,
                            '2005-12-31 23:59:59.9999999',
                            '2006-01-01 00:00:00.0000000');`,
		v: [][]string{{"1"}},
	},
	{
		q: `SELECT DATEDIFF(nanosecond,
                            '2005-12-31 23:59:59.9999999',
                            '2006-01-01 00:00:00.0000000');`,
		v: [][]string{{"100"}},
	},
	{
		q: `DECLARE now DATETIME;
SET now = GETDATE();
SELECT DATEDIFF(year, now, now);`,
		v: [][]string{{"0"}},
	},
	{
		q: `SELECT YEAR('2005-12-31 23:59:59.9999999');`,
		v: [][]string{{"2005"}},
	},
}

func TestBuiltIn(t *testing.T) {
	data := fmt.Sprintf("data:text/csv;base64,%s",
		base64.StdEncoding.EncodeToString([]byte(builtInData)))

	for testID, input := range builtInTests {
		name := fmt.Sprintf("Test %d", testID)
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
			verifyResult(t, name, q, input.v)
		}
	}
}

func verifyResult(t *testing.T, name string, q types.Source, v [][]string) {
	rows, err := q.Get()
	if err != nil {
		t.Errorf("%s: q.Get failed: %v", name, err)
		return
	}
	if len(rows) != len(v) {
		t.Errorf("%s: got %d rows, expected %d", name, len(rows), len(v))
		printResult(q, rows)
		return
	}
	for rowID, row := range rows {
		if len(row) != len(v[rowID]) {
			t.Fatalf("%s: row %d: got %d columns, expected %d",
				name, rowID, len(row), len(v[rowID]))
			printResult(q, rows)
			continue
		}
		for colID, col := range row {
			result := col.String()
			if result != v[rowID][colID] {
				t.Errorf("%s: %d.%d: got '%s', expected '%s'",
					name, rowID, colID, result, v[rowID][colID])
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
