package polona

import (
	"bytes"
	"diglib/generic_library"
	strg "diglib/storage"
	"diglib/tools"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gocolly/colly/v2"
	"os"
	"path/filepath"
	"regexp"
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
	Imprint              multiString `json:"imprint"`
	Series               []string    `json:"series"`                 // "series": [ "Town plans of Poland (Geographical Section. General Staff)" ],
	PhysicalDescription  multiString `json:"physical_description"`   // "1 mapa : wielobarwna ; 58x72 cm, arkusz 71x84 cm",
	Categories           []string    `json:"categories"`             // ["maps"],
	Metatypes            []string    `json:"metatypes"`              // ["mapa ogólnogeograficzna"],
	Udc                  []string    `json:"udc"`                    // ["912(4)(084.3)","912(438)(084.3)"],
	CallNo               []string    `json:"call_no"`                // ["ZZK S-31 539 A"],
	CartographicMathData multiString `json:"cartographic_math_data"` // "Skala 1:1 000 000 (E 20°40'-E 30°50'/ N 57°02'-N 52°00').",
	Scans                []Scan      `json:"scans"`
}

func (pol *PolonaObject) GetSygn() string {
	return strings.ReplaceAll(strings.Join(pol.CallNo, " "), "/", "|")
}

func (pol *PolonaObject) GetScale() string {
	re := regexp.MustCompile(`1 ?: ?(?P<scale>[0-9 ]+)`)
	matches := re.FindStringSubmatch(string(pol.CartographicMathData))
	if len(matches) > 0 {
		return strings.ReplaceAll(matches[1], " ", "")
	}

	return "unknown"
}

func (pol *PolonaObject) GetSerie() string {
	if pol.Series != nil {
		return pol.Series[0]
	}

	return pol.GetScale()
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

func DownloadPolona(item *strg.Item, outputFolder string, onlyMetadata bool) error {
	if onlyMetadata == true && item.DataProviderMetaJSON != "" {
		return errors.New("metadata already set")
	}

	var resourceId string
	var polonaObject PolonaObject
	item.LastUpdateDate = time.Now().Format("2006-01-02 15:04:05")

	page := colly.NewCollector()
	generic_library.SetDefaultDownloadSettings(page)
	jsonResource := colly.NewCollector()
	generic_library.SetDefaultDownloadSettings(jsonResource)

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
			if len(scan.Resources) > 0 {
				scanUrls = append(scanUrls, scan.Resources[0].Url)
			}
		}
	})
	// fmt.Printf("%s", "https://polona.pl/api/entities/" + resourceId)

	page.Visit(item.Link)

	if onlyMetadata == true {
		fmt.Printf("Metadata downloaded for %s, %s\n", item.Guid, polonaObject.Slug)
		return nil
	}

	serieForPath := tools.ClearPathStringForWindows(polonaObject.GetSerie())
	dstPath := filepath.Join(outputFolder, serieForPath)
	jsonFileNamePrefix := polonaObject.Slug + "_" + item.Guid
	err := os.MkdirAll(dstPath, os.ModePerm)
	if err != nil {
		fmt.Printf("Error while creating destination folder: %v\n", dstPath)
		panic(err)
	}

	jsonItemBytes, _ := json.MarshalIndent(item, "", "  ")
	f, _ := os.Create(filepath.Join(dstPath, jsonFileNamePrefix+".json"))
	f.WriteString(string(jsonItemBytes))
	f.Close()
	f, _ = os.Create(filepath.Join(dstPath, jsonFileNamePrefix+".polona.json"))
	f.WriteString(string(prettyJSON.Bytes()))
	f.Close()

	for i, url := range scanUrls {
		retry := 0
		imageResource := colly.NewCollector()
		generic_library.SetDefaultDownloadSettings(imageResource)
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
			fileName := tools.ClearPathStringForWindows(
				fmt.Sprintf("%s_%s_Sygn.%s_%04d%s", polonaObject.Slug, item.Guid, polonaObject.GetSygn(),
					i, filepath.Ext(res.FileName())))
			fileNameWithPath := filepath.Join(dstPath, fileName)
			err := res.Save(fileNameWithPath)
			if err != nil {
				fmt.Printf("Error while saving file: %v\n", fileNameWithPath)
				panic(err)
			}
			item.Download = fileNameWithPath
			fmt.Printf("Downloaded %s\n", res.Request.URL)
		})

		fmt.Printf("Downloading file %d of %d, %s, %s from %s.\n", i+1, len(scanUrls), item.Guid,
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

	return nil
}
