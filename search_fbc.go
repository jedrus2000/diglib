package main

import (
	strg "diglib/storage"
	"encoding/xml"
	"fmt"
	"github.com/gocolly/colly/v2"
	"strconv"
	"time"
)

func Search(storage *strg.Storage, pageSize int) {
	saved := 0
	currentDateTimeString := time.Now().Format("2006-01-02 15:04:05")
	startPage := 1
	// http://fbc.pionier.net.pl/opensearch/search?count=1000&startIndex=1&searchTerms=dc_type%3A(mapa)%20OR%20dc_description%3A(mapa)
	url := func() string {
		return "http://fbc.pionier.net.pl/opensearch/search?searchTerms=dc_type%3A(mapa)%20OR%20dc_description%3A(mapa)%20OR%20dc_description%3A(plan)&count=" +
			strconv.Itoa(pageSize) + "&startPage=" + strconv.Itoa(startPage)
	}

	c := colly.NewCollector()
	c.SetRequestTimeout(60 * time.Second)
	c.AllowURLRevisit = true

	c.OnResponse(func(res *colly.Response) {
		// xml.
		// fmt.Println(string(res.Body))
		v := strg.Rss{}
		err := xml.Unmarshal(res.Body, &v)
		if err != nil {
			fmt.Printf("Unmarshal error: %v\n", err)
			return
		}

		for idx, _ := range v.Channels[0].Items {
			item := &v.Channels[0].Items[idx]
			item.Download = ""
			item.LastUpdateDate = currentDateTimeString
			item.DataProviderMetaJSON = ""
			reduceDuplicateStrings(&item.Title)
			reduceDuplicateStrings(&item.Type)
		}
		saved += storage.SaveItems(&v.Channels[0].Items, false)

		if v.Channels[0].TotalResults > 0 {
			totalResults := v.Channels[0].TotalResults
			if startPage*pageSize < totalResults {
				var err error
				tryNext := 0
				for {
					startPage += 1
					if err = c.Visit(url()); err != nil {
						fmt.Printf("Visit error: %v\n", err)
						tryNext++
					} else {
						break
					}
					if tryNext > 2 {
						fmt.Printf("Too much errors. Quits...\n")
						break
					}
				}
			}
		}
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.Visit(url())

	fmt.Printf("Saved %v new items, with date-time: %v.\n", saved, currentDateTimeString)
}

func Dump(storage *strg.Storage, ss *strg.Spreadsheet) {
	ss.SetDefaultHeader()

	ssRowId := 2

	storage.ForEach(
		func(item *strg.Item) {
			ss.SetDefaultData(ssRowId, item)
			ssRowId++
		})
	// fmt.Printf("%+v \n", item)
}

func reduceDuplicateStrings(list *[]string) {
	if len(*list) == 1 {
		return
	}
	if (*list)[0] == (*list)[1] {
		(*list) = append((*list)[:1])
	}
}
