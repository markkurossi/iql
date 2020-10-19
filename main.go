//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/markkurossi/iql/data"
	"github.com/markkurossi/iql/query"
	"github.com/markkurossi/tabulate"
)

func main() {
	flag.Parse()
	log.SetFlags(0)

	for _, arg := range flag.Args() {
		f, err := os.Open(arg)
		if err != nil {
			fmt.Printf("failed to open '%s': %s\n", arg, err)
			os.Exit(1)
		}
		parser := query.NewParser(f, arg)
		q, err := parser.Parse()
		if err != nil {
			fmt.Printf("%s\n", err)
			os.Exit(1)
		}
		rows, err := q.Get()
		if err != nil {
			fmt.Printf("Query failed: %v\n", err)
			os.Exit(1)
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
