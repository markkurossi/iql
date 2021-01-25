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
	"runtime/pprof"
	"strings"

	"github.com/markkurossi/iql"
	"github.com/markkurossi/iql/data"
	"github.com/markkurossi/iql/lang"
	"github.com/markkurossi/tabulate"
)

func main() {
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to `file`")
	htmlFilter := flag.String("html", "", "HTML filter")
	jsonFilter := flag.String("json", "", "JSON filter")
	tableFmt := flag.String("t", "uc", "table formatting style")
	expr := flag.String("e", "", "code to execute")
	flag.Parse()
	log.SetFlags(0)

	program := os.Args[0]
	idx := strings.LastIndex(program, "/")
	if idx >= 0 {
		program = program[idx+1:]
	}

	if len(*cpuprofile) > 0 {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	if len(*expr) > 0 {
		client := newClient(program, *tableFmt)
		err := client.SetStringArray(lang.SysARGS, flag.Args())
		if err != nil {
			log.Fatalf("%s: %s\n", program, err)
		}
		err = client.Parse(strings.NewReader(*expr), "expr")
		if err != nil {
			log.Fatalf("%s: %s\n", program, err)
		}
		return
	}

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
		} else if len(*jsonFilter) > 0 {
			result, err := data.JSONFilter(f, *jsonFilter)
			if err != nil {
				log.Fatalf("JSON filter: %s\n", err)
			}
			for idx, r := range result {
				fmt.Printf("%s:%s: nth=%d:\n%v\n", arg, *htmlFilter, idx, r)
			}
		} else {
			client := newClient(program, *tableFmt)
			err = client.Parse(f, arg)
			if err != nil {
				log.Fatalf("%s: %s\n", arg, err)
			}
		}
	}
}

func newClient(program, tableFmt string) *iql.Client {
	client := iql.NewClient(os.Stdout)
	err := client.SetString(lang.SysTableFmt, tableFmt)
	if err != nil {
		log.Printf("%s: %s\n", program, err)
		log.Fatalf("Possible styles are: %s\n",
			strings.Join(tabulate.StyleNames(), ", "))
	}
	return client
}
