package main

import (
	"encoding/xml"
	"fmt"
	"github.com/gocolly/colly/v2"
)

type Rss struct {
	// XMLName xml.Name 	`xml:"rss"`
	Channels []Channel `xml:"channel"`
}

type Channel struct {
	// XMLName xml.Name 				`xml:"channel"`
	TotalResults string `xml:"totalResults"`
}

func Search() {
	c := colly.NewCollector()

	/*
		c.OnXML("//channel", func(e *colly.XMLElement) {
			fmt.Println(e.Attr("item"))
		})
	*/

	c.OnResponse(func(res *colly.Response) {
		// xml.
		// fmt.Println(string(res.Body))
		v := Rss{}
		err := xml.Unmarshal(res.Body, &v)
		if err != nil {
			fmt.Printf("error: %v", err)
			return
		}
		fmt.Println(v)
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.Visit("http://fbc.pionier.net.pl/opensearch/search?searchTerms=dc_type%3A(mapa)")
}
