//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/PuerkitoBio/goquery"
	"github.com/markkurossi/htmlq/data"
	"github.com/markkurossi/htmlq/query"
	"github.com/markkurossi/tabulate"
)

func main() {
	if true {
		test()
		return
	}

	doc, err := goquery.NewDocumentFromReader(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	tab := tabulate.New(tabulate.Unicode)
	tab.Header("Name")
	tab.Header("Price").SetAlign(tabulate.MR)
	tab.Header("Weight").SetAlign(tabulate.MR)

	doc.Find("tbody > tr").Each(func(i int, s *goquery.Selection) {
		name := s.Find(".name")
		avg := s.Find(".avgprice")
		share := s.Find(".share")
		link := name.Find("a")
		if link.Length() > 0 {
			row := tab.Row()
			row.Column(link.Text())
			row.Column(avg.Text())
			row.Column(share.Text())

			if false {
				fmt.Printf("%d:\t%s\t%s\t%s\n", i, link.Text(), avg.Text(),
					share.Text())
			}
		}
	})
	tab.Print(os.Stdout)
}

func test() {
	ref, err := data.NewHTMLFromReader(os.Stdin, "tbody > tr",
		[]data.ColumnSelector{
			{
				Name: data.Reference{
					Column: ".name",
				},
				As: "Name",
			},
			{
				Name: data.Reference{
					Column: ".avgprice",
				},
				As:    "Price",
				Align: tabulate.MR,
			},
			{
				Name: data.Reference{
					Column: ".share",
				},
				As:    "Weight",
				Align: tabulate.MR,
			},
			{
				Name: data.Reference{
					Column: "a",
				},
				As: "link",
			},
		})
	q := &query.Query{
		Select: []data.ColumnSelector{
			{
				Name: data.Reference{
					Column: "Name",
				},
			},
			{
				Name: data.Reference{
					Column: "Price",
				},
				Align: tabulate.MR,
			},
			{
				Name: data.Reference{
					Column: "Weight",
				},
				Align: tabulate.MR,
			},
			{
				Name: data.Reference{
					Column: "link",
				},
			},
		},
		From: []query.SourceSelector{
			{
				Source: ref,
				As:     "ref",
			},
		},
		Where: &query.Binary{
			Type: query.BinNEQ,
			Left: &query.Reference{
				Reference: data.Reference{
					Column: "link",
				},
			},
			Right: &query.Constant{
				Value: query.StringValue(""),
			},
		},
	}

	rows, err := q.Get()
	if err != nil {
		log.Fatal(err)
	}
	tab := tabulate.New(tabulate.Unicode)
	for _, col := range q.Columns() {
		if col.IsPublic() {
			tab.Header(col.String()).SetAlign(col.Align)
		}
	}
	for _, columns := range rows {
		row := tab.Row()
		for _, col := range columns {
			row.Column(col.String())
		}
	}

	tab.Print(os.Stdout)
}
