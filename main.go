//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/markkurossi/iql/data"
	"github.com/markkurossi/iql/query"
	"github.com/markkurossi/iql/types"
	"github.com/markkurossi/tabulate"
)

func main() {
	htmlFilter := flag.String("html", "", "HTML filter")
	flag.Parse()
	log.SetFlags(0)

	for _, arg := range flag.Args() {
		f, err := os.Open(arg)
		if err != nil {
			log.Fatalf("failed to open '%s': %s\n", arg, err)
		}

		if len(*htmlFilter) > 0 {
			result, err := data.HTMLFilter(f, *htmlFilter)
			if err != nil {
				log.Fatalf("HTML filter: %s\n", err)
			}
			for idx, r := range result {
				fmt.Printf("%s:%s: nth=%d:\n%s\n", arg, *htmlFilter, idx+1, r)
			}
		} else {
			parser := query.NewParser(f, arg)
			for {
				q, err := parser.Parse()
				if err != nil {
					if err == io.EOF {
						break
					}
					log.Fatalf("%s\n", err)
				}
				tab, err := types.Tabulate(q, tabulate.Unicode)
				if err != nil {
					log.Fatalf("Query failed: %v\n", err)
				}
				tab.Print(os.Stdout)
			}
		}
	}
}
