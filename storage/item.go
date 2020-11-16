package storage

import (
	"regexp"
	"strings"
)

type Rss struct {
	// XMLName xml.Name 	`xml:"rss"`
	Channels []Channel `xml:"channel"`
}

type Channel struct {
	// XMLName xml.Name 				`xml:"channel"`
	TotalResults int    `xml:"totalResults"`
	StartIndex   int    `xml:"startIndex"`
	ItemsPerPage int    `xml:"itemsPerPage"`
	Items        []Item `xml:"item"`
}

type Item struct {
	Download             string
	LastUpdateDate       string
	Title                []string `xml:"title"`
	Link                 string   `xml:"link"`
	Guid                 string   `xml:"guid"`
	Contributor          string   `xml:"contributor"`
	Subject              string   `xml:"subject"`
	Publisher            string   `xml:"publisher"`
	Description          string   `xml:"description"`
	Date                 string   `xml:"date"`
	Type                 []string `xml:"type"`
	Format               string   `xml:"format"`
	Source               string   `xml:"source"`
	Language             string   `xml:"language"`
	Rights               string   `xml:"rights"`
	DataProvider         string   `xml:"dataProvider"`
	DataProviderMetaJSON string
}

func (item *Item) GetScale() string {
	re := regexp.MustCompile(`1 ?: ?(?P<scale>[0-9 ]+)`)
	matches := re.FindStringSubmatch(item.Description)
	if len(matches) <= 0 {
		matches = re.FindStringSubmatch(strings.Join(item.Title, " "))
	}
	if len(matches) > 0 {
		return strings.ReplaceAll(matches[1], " ", "")
	}

	return "unknown"
}
