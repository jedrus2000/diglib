package polona

import (
	strg "diglib/storage"
	"encoding/json"
	"github.com/gocolly/colly/v2"
	"strings"
)

type Resource struct {
	Url  string `json:"url"`  // "https://polona.pl/archive?uid=68472740&cid=68611965&name=download_fullJPG",
	Mime string `json:"mime"` // "image/jpeg"
}

type MainScan struct {
	Resources []Resource `json:"resources"`
}

type PolonaObject struct {
	Id                   string   `json:"id"`
	Slug                 string   `json:"slug"`
	Title                string   `json:"title"`
	Alternative          string   `json:"alternative"`
	Contributors         []string `json:"contributor_corp"`
	Date                 string   `json:"date"`
	DateDescriptive      string   `json:"date_descriptive"`
	Country              string   `json:"country"`
	Publisher            string   `json:"publisher"`
	Imprint              string   `json:"imprint"`
	PhysicalDescription  string   `json:"physical_description"`   // "1 mapa : wielobarwna ; 58x72 cm, arkusz 71x84 cm",
	Categories           []string `json:"categories"`             // ["maps"],
	Metatypes            []string `json:"metatypes"`              // ["mapa ogólnogeograficzna"],
	Udc                  []string `json:"udc"`                    // ["912(4)(084.3)","912(438)(084.3)"],
	CallNo               []string `json:"call_no"`                // ["ZZK S-31 539 A"],
	CartographicMathData string   `json:"cartographic_math_data"` // "Skala 1:1 000 000 (E 20°40'-E 30°50'/ N 57°02'-N 52°00').",
	MainScan             MainScan `json:"main_scan"`
}

func DownloadPolona(item *strg.Item) {
	var resourceId string
	var polonaObject PolonaObject
	// var polonaObjectString string
	page := colly.NewCollector()

	page.OnResponse(func(res *colly.Response) {
		resourceId = strings.Split(res.Request.URL.Path, ",")[1]
	})
	page.Visit(item.Link)

	jsonResource := colly.NewCollector()
	jsonResource.OnResponse(func(res *colly.Response) {
		err := json.Unmarshal(res.Body, &polonaObject)
		// polonaObjectString = fmt.Sprint(res.Body[:])
		if err != nil {
			panic(err)
		}
	})
	jsonResource.Visit("https://polona.pl/api/entities/" + resourceId)

	imageResource := colly.NewCollector()
	imageResource.OnResponse(func(res *colly.Response) {
		err := res.Save("something.jpg")
		if err != nil {
			panic(err)
		}
	})
	jsonResource.Visit(polonaObject.MainScan.Resources[0].Url)

}
