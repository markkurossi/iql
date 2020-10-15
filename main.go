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
	"github.com/markkurossi/iql/data"
	"github.com/markkurossi/iql/query"
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
					Column: ":nth-child(5)",
				},
				As: "Price",
			},
			{
				Name: data.Reference{
					Column: ".share",
				},
				As: "Weight",
			},
			{
				Name: data.Reference{
					Column: "a",
				},
				As: "link",
			},
		})
	portfolio, err := data.NewCSV(",portfolio.csv", "", []data.ColumnSelector{
		{
			Name: data.Reference{
				Column: "0",
			},
			As: "name",
		},
		{
			Name: data.Reference{
				Column: "1",
			},
			As: "Count",
		},
	})

	q := &query.Query{
		Select: []query.ColumnSelector{
			{
				Expr: &query.Reference{
					Reference: data.Reference{
						Source: "ref",
						Column: "Name",
					},
				},
				As: "Name",
			},
			{
				Expr: &query.Reference{
					Reference: data.Reference{
						Column: "Price",
					},
				},
				As: "Price",
			},
			{
				Expr: &query.Reference{
					Reference: data.Reference{
						Column: "Weight",
					},
				},
				As: "Weight",
			},
			{
				Expr: &query.Reference{
					Reference: data.Reference{
						Column: "link",
					},
				},
				As: "link",
			},
			{
				Expr: &query.Reference{
					Reference: data.Reference{
						Source: "portfolio",
						Column: "name",
					},
				},
				As: "name",
			},
			{
				Expr: &query.Reference{
					Reference: data.Reference{
						Source: "portfolio",
						Column: "Count",
					},
				},
				As: "Count",
			},
			{
				Expr: &query.Binary{
					Type: query.BinMult,
					Left: &query.Reference{
						Reference: data.Reference{
							Source: "portfolio",
							Column: "Count",
						},
					},
					Right: &query.Reference{
						Reference: data.Reference{
							Source: "ref",
							Column: "Price",
						},
					},
				},
				As: "Invested",
			},

			{
				Expr: &query.Binary{
					Type: query.BinMult,
					Left: &query.Binary{
						Type: query.BinDiv,
						Left: &query.Binary{
							Type: query.BinMult,
							Left: &query.Reference{
								Reference: data.Reference{
									Source: "portfolio",
									Column: "Count",
								},
							},
							Right: &query.Reference{
								Reference: data.Reference{
									Source: "ref",
									Column: "Price",
								},
							},
						},
						Right: &query.Function{
							Type: query.FuncSum,
							Arguments: []query.Expr{
								&query.Binary{
									Type: query.BinMult,
									Left: &query.Reference{
										Reference: data.Reference{
											Source: "portfolio",
											Column: "Count",
										},
									},
									Right: &query.Reference{
										Reference: data.Reference{
											Source: "ref",
											Column: "Price",
										},
									},
								},
							},
						},
					},
					Right: &query.Constant{
						Value: data.IntValue(100),
					},
				},
				As: "My Weight",
			},
		},
		From: []query.SourceSelector{
			{
				Source: ref,
				As:     "ref",
			},
			{
				Source: portfolio,
				As:     "portfolio",
			},
		},
		Where: &query.Binary{
			Type: query.BinAnd,
			Left: &query.Binary{
				Type: query.BinNeq,
				Left: &query.Reference{
					Reference: data.Reference{
						Column: "link",
					},
				},
				Right: &query.Constant{
					Value: data.StringValue(""),
				},
			},
			Right: &query.Binary{
				Type: query.BinEq,
				Left: &query.Reference{
					Reference: data.Reference{
						Source: "ref",
						Column: "Name",
					},
				},
				Right: &query.Reference{
					Reference: data.Reference{
						Source: "portfolio",
						Column: "name",
					},
				},
			},
		},
	}

	rows, err := q.Get()
	if err != nil {
		log.Fatal(err)
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
