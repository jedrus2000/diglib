package polona

import (
	"bytes"
	strg "diglib/storage"
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly/v2"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type multiString string

type Resource struct {
	Url  string `json:"url"`  // "https://polona.pl/archive?uid=68472740&cid=68611965&name=download_fullJPG",
	Mime string `json:"mime"` // "image/jpeg"
}

type Scan struct {
	Resources []Resource `json:"resources"`
}

type PolonaObject struct {
	Id                   string      `json:"id"`
	Slug                 string      `json:"slug"`
	Title                string      `json:"title"`
	Alternative          string      `json:"alternative"`
	Contributors         []string    `json:"contributor_corp"`
	Date                 string      `json:"date"`
	DateDescriptive      string      `json:"date_descriptive"`
	Country              string      `json:"country"`
	Publisher            string      `json:"publisher"`
	Imprint              string      `json:"imprint"`
	PhysicalDescription  multiString `json:"physical_description"`   // "1 mapa : wielobarwna ; 58x72 cm, arkusz 71x84 cm",
	Categories           []string    `json:"categories"`             // ["maps"],
	Metatypes            []string    `json:"metatypes"`              // ["mapa ogólnogeograficzna"],
	Udc                  []string    `json:"udc"`                    // ["912(4)(084.3)","912(438)(084.3)"],
	CallNo               []string    `json:"call_no"`                // ["ZZK S-31 539 A"],
	CartographicMathData multiString `json:"cartographic_math_data"` // "Skala 1:1 000 000 (E 20°40'-E 30°50'/ N 57°02'-N 52°00').",
	Scans                []Scan      `json:"scans"`
}

func (ms *multiString) UnmarshalJSON(data []byte) error {
	if len(data) > 0 {
		switch data[0] {
		case '"':
			var s string
			if err := json.Unmarshal(data, &s); err != nil {
				return err
			}
			*ms = multiString(s)
		case '[':
			var s []string
			if err := json.Unmarshal(data, &s); err != nil {
				return err
			}
			*ms = multiString(strings.Join(s, "&"))
		}
	}
	return nil
}

func DownloadPolona(item *strg.Item, outputFolder string) {
	var resourceId string
	var polonaObject PolonaObject
	item.LastUpdateDate = time.Now().Format("2006-01-02 15:04:05")

	page := colly.NewCollector()
	jsonResource := colly.NewCollector()

	page.OnResponse(func(res *colly.Response) {
		resourceId = strings.Split(res.Request.URL.Path, ",")[1]
		jsonResource.Visit("https://polona.pl/api/entities/" + resourceId)
	})

	var scanUrls []string
	var prettyJSON bytes.Buffer
	jsonResource.OnResponse(func(res *colly.Response) {
		err := json.Unmarshal(res.Body, &polonaObject)
		if err != nil {
			panic(err)
		}
		// fmt.Printf("%+v\n", polonaObject) // res.Body[:])
		json.Indent(&prettyJSON, res.Body, "", "  ")
		item.DataProviderMetaJSON = string(prettyJSON.Bytes())
		for _, scan := range polonaObject.Scans {
			scanUrls = append(scanUrls, scan.Resources[0].Url)
		}
	})
	// fmt.Printf("%s", "https://polona.pl/api/entities/" + resourceId)

	page.Visit(item.Link)

	scansFolder := ""
	if len(scanUrls) > 1 {
		scansFolder = fmt.Sprintf("%s_%s", polonaObject.Slug, item.Guid)
		os.MkdirAll(filepath.Join(outputFolder, scansFolder), os.ModePerm)
	}
	jsonItemBytes, _ := json.MarshalIndent(item, "", "  ")
	f, _ := os.Create(filepath.Join(outputFolder, scansFolder, item.Guid+".json"))
	f.WriteString(string(jsonItemBytes))
	f.Close()
	f, _ = os.Create(filepath.Join(outputFolder, scansFolder, item.Guid+".polona.json"))
	f.WriteString(string(prettyJSON.Bytes()))
	f.Close()

	for i, url := range scanUrls {
		retry := 0
		imageResource := colly.NewCollector()
		imageResource.MaxBodySize = 10 * 1024 * 1024 * 1024
		imageResource.SetRequestTimeout(60 * time.Second)
		imageResource.AllowURLRevisit = true

		imageResource.OnError(func(res *colly.Response, err error) {
			if (retry < 3) && ((res.StatusCode < 400) || (res.StatusCode > 499)) {
				retry++
				var errorStr = fmt.Sprintf("Error %d %s while downloading %s\nSleep 3 secs and retry %d",
					res.StatusCode, err.Error(), res.Request.URL, retry)
				fmt.Println(errorStr)
				time.Sleep(3 * time.Second)
			} else {
				retry = 4
				var errorStr = fmt.Sprintf("Error %d %s while downloading %s",
					res.StatusCode, err.Error(), res.Request.URL)
				item.Download = errorStr
				fmt.Println(errorStr)
			}
		})

		imageResource.OnResponse(func(res *colly.Response) {
			var fileName string
			if scansFolder != "" {
				fileName = fmt.Sprintf("%05d%s", i, filepath.Ext(res.FileName()))
			} else {
				fileName = fmt.Sprintf("%s_%s_%s", polonaObject.Slug, item.Guid, res.FileName())
			}
			fileNameWithPath := filepath.Join(outputFolder, scansFolder, fileName)
			err := res.Save(fileNameWithPath)
			if err != nil {
				panic(err)
			}
			item.Download = fileNameWithPath
			fmt.Printf("Downloaded %s\n", res.Request.URL)
		})

		fmt.Printf("Downloading %s, %s from %s.\n", item.Guid,
			polonaObject.Slug, url)
		// fmt.Printf("Async: %v\n", imageResource.Async)
		err := imageResource.Visit(url)
		if retry == 4 {
			break
		}
		for (err != nil) && (retry < 3) {
			err = imageResource.Visit(url)
		}
	}

}
