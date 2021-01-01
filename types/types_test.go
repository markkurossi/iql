//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package types

import (
	"fmt"
	"testing"
)

var dateTests = []string{
	"2007-04-30T01:01:01.1234567 -07:00",
	"2005-12-31 23:59:59.9999999",
	"2005-12-31 23:59:59.999",
	"2005-12-31 23:59:59",
	"2005-12-31",
	"12/19/2020",
}

func TestDate(t *testing.T) {
	for _, test := range dateTests {
		date, err := ParseDate(test)
		if err != nil {
			t.Errorf("ParseDate(%s) failed: %s", test, err)
		} else if false {
			fmt.Printf("%s => %s\n", test, date)
		}
	}
}
