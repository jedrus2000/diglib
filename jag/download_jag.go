package jag

import (
	strg "diglib/storage"
	"errors"
	"fmt"
	"github.com/gocolly/colly/v2"
	"strings"
)

func DownloadJag(item *strg.Item, outputFolder string, onlyMetadata bool) error {
	if onlyMetadata == true && item.DataProviderMetaJSON != "" {
		return errors.New("metadata already set")
	}

	var resourceId string
	// var polonaObject PolonaObject
	// item.LastUpdateDate = time.Now().Format("2006-01-02 15:04:05")

	page := colly.NewCollector()
	jsonResource := colly.NewCollector()

	page.OnResponse(func(res *colly.Response) {
		fmt.Printf("%s\n", res.Request.URL.Path)
		resourceId = strings.Split(res.Request.URL.Path, "?id=")[1]
		fmt.Printf("%s\n", resourceId)
		jsonResource.Visit("https://jbc.bj.uj.edu.pl/dlibra/oai-pmh-repository.xml?verb=GetRecord&metadataPrefix=mets&identifier=oai:jbc.bj.uj.edu.pl:" + resourceId)
	})

	// var scanUrls []string
	// var prettyJSON bytes.Buffer
	jsonResource.OnResponse(func(res *colly.Response) {
		fmt.Printf("%+v\n", string(res.Body)) // res.Body[:])
	})
	// fmt.Printf("%s", "https://polona.pl/api/entities/" + resourceId)

	println("Jag")
	fmt.Printf("%s\n", item.Link)
	err := page.Visit(item.Link)
	fmt.Printf("%+v\n", err)
	return nil
}
