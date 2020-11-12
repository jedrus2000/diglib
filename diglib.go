package main

import (
	"diglib/polona"
	strg "diglib/storage"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docopt/docopt-go"
	"os"
)

func main() {
	storage := &strg.Storage{}
	storage.Open()
	defer storage.Close()

	usage := `
digilib
	
	Usage:
	  digilib search [--page-size=<value>]
	  digilib dump <output_filename>
	  digilib download [--single=<guid>] [--download-indicator=<value>] [--output-folder=<value>]
	  digilib show <item_guid>
	  digilib -h | --help
  	  digilib --version

	Options:
	  --output-folder=<value>		  Output folder [default: ./]
	  --page-size=<value>  			  Page size of results [default: 1000].
      --single=<guid>				  ID (guid) of element.
	  --download-indicator=<value>    Download indicator [default: 1].
	  -h --help                       Show this screen.
  	  --version                       Show version.	

`
	arguments, _ := docopt.ParseDoc(usage)

	// fmt.Println(arguments)

	if arguments["dump"] == true {
		filename, _ := arguments.String("<output_filename>")
		ss := &strg.Spreadsheet{}
		ss.Open(filename)
		defer ss.Close()
		Dump(storage, ss)
	} else if arguments["search"] == true {
		pageSize, _ := arguments.Int("--page-size")
		Search(storage, pageSize)
	} else if arguments["download"] == true {
		downloadIndicator, err := arguments.String("--download-indicator")
		outputFolder, err := arguments.String("--output-folder")
		os.MkdirAll(outputFolder, os.ModePerm)
		if err != nil {
			panic(err)
		}
		if guid, err := arguments.String("--single"); err == nil {
			if item, err := storage.Find(guid); err == nil {
				if err := downloadItem(&item, downloadIndicator, outputFolder); err == nil {
					storage.SaveItem(&item)
				}
			} else {
				fmt.Println(err)
			}
		} else {
			storage.ForEach(func(item *strg.Item) {
				if err := downloadItem(item, downloadIndicator, outputFolder); err == nil {
					storage.SaveItem(item)
				}
			})
		}
	} else if arguments["show"] == true {
		guid, err := arguments.String("<item_guid>")
		if err != nil {
			panic(err)
		}
		if item, err := storage.Find(guid); err == nil {
			jsonItem, _ := json.MarshalIndent(item, "", "  ")
			fmt.Println(string(jsonItem))
		} else {
			fmt.Println(err)
		}
	}
}

func downloadItem(item *strg.Item, downloadIndicator string, outputFolder string) error {
	if item.Download != downloadIndicator {
		fmt.Printf("Item %s download skipped. Download indictator: %s differs.\n", item.Guid, item.Download)
		return errors.New("download indicator differs")
	}
	switch item.DataProvider {
	case "CBN Polona":
		polona.DownloadPolona(item, outputFolder)
	default:
		fmt.Printf("Data provider %s not supported.\n", item.DataProvider)
		return errors.New("data provider not supported")
	}
	return nil
}
