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
	q := &query.Query{
		Columns: []query.ColumnSelector{
			{
				Find: ".name",
				As:   "Name",
			},
			{
				Find:  ".avgprice",
				As:    "Price",
				Align: tabulate.MR,
			},
			{
				Find:  ".share",
				As:    "Weight",
				Align: tabulate.MR,
			},
			{
				Find: "a",
				As:   "link",
			},
		},
		Document: "tbody > tr",
		Where: &query.Binary{
			Type: query.BinGT,
			Left: &query.Function{
				Type:      query.FuncLength,
				Arguments: []string{"link"},
			},
			Right: &query.Constant{
				Value: query.IntValue(0),
			},
		},
	}

	doc, err := goquery.NewDocumentFromReader(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	tab := tabulate.New(tabulate.Unicode)
	for _, col := range q.Columns {
		if col.IsPublic() {
			tab.Header(col.As).SetAlign(col.Align)
		}
	}

	q.Execute(doc, func(columns []string) error {
		row := tab.Row()
		for _, col := range columns {
			row.Column(col)
		}
		return nil
	})

	tab.Print(os.Stdout)
}
