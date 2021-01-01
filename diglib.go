package main

import (
	"diglib/jag"
	"diglib/polona"
	strg "diglib/storage"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docopt/docopt-go"
	"os"
	"os/signal"
	"regexp"
	"syscall"
)

func main() {
	usage := `
diglib
	
	Usage:
	  diglib search [--page-size=<value>]
	  diglib dump <output_filename> 
	  diglib download [--item-selector=<item_guid>] [--download-selector=<value>] [--scale-selector=<value>] 
			[--library-selector=<value>] 
			[--dry-run] [--output-folder=<value>] [--only-metadata]
	  diglib set-property [--single=<guid>] [--download-selector=<value>]		
	  diglib show <item_guid>
	  diglib -h | --help
  	  diglib --version

	Options:
	  --output-folder=<value>         Output folder [default: ./downloads].
	  --page-size=<value>             Page size of results [default: 1000].
	  --single=<guid>                 ID (guid) of element.
	  --dry-run                       Prints selected items for download, without downloading content.
	  -h --help                       Show this screen.
  	  -v --version                    Show version.	

`
	arguments, _ := docopt.ParseDoc(usage)

	// fmt.Println(arguments)

	if version, _ := arguments.Bool("--version"); version == true {
		fmt.Println("0.3") // tag added
		syscall.Exit(0)
	}

	storage := &strg.Storage{}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			fmt.Printf("Got %v !\n", sig)
			storage.Close()
			os.Exit(0)
		}
	}()

	storage.Open()
	defer storage.Close()

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
		outputFolder, _ := arguments.String("--output-folder")
		err := os.MkdirAll(outputFolder, os.ModePerm)
		if err != nil {
			panic(err)
		}
		dryRun, _ := arguments.Bool("--dry-run")
		onlyMetadata, _ := arguments.Bool("--only-metadata")
		var downloadSelectorRe, librarySelectorRe, itemSelectorRe *regexp.Regexp

		itemSelector, err := arguments.String("--item-selector")
		if err != nil {
			itemSelector = ""
		}
		itemSelectorRe = regexp.MustCompile(itemSelector)

		downloadSelector, err := arguments.String("--download-selector")
		if err != nil {
			downloadSelector = ""
		}
		downloadSelectorRe = regexp.MustCompile(downloadSelector)

		scaleSelector, err := arguments.String("--scale-selector")
		if err != nil {
			scaleSelector = ""
		}

		librarySelector, err := arguments.String("--library-selector")
		if err != nil {
			librarySelector = ""
		}
		librarySelectorRe = regexp.MustCompile(librarySelector)

		counter := 0
		missingProviderMetadataCounter := 0
		dot := false
		storage.ForEach(func(item *strg.Item) {
			if downloadSelectorRe.MatchString(item.Download) &&
				librarySelectorRe.MatchString(item.DataProvider) &&
				itemSelectorRe.MatchString(item.Guid) &&
				matchScale(item, scaleSelector, &missingProviderMetadataCounter) {
				if dot == true && dryRun == false {
					println("")
					dot = false
				}
				if dryRun == true {
					fmt.Printf("%s, %s %s \n", item.Guid, item.Title, item.DataProvider)
				} else if err := downloadItem(item, outputFolder, onlyMetadata); err == nil {
					storage.SaveItem(item, true)
				}
				counter++
			} else if dryRun == false {
				print(".")
				dot = true
			}
		})
		fmt.Printf("%d items. %d have missing data provider metadata that can be useful with provided selectors.\n", counter, missingProviderMetadataCounter)
	} else if arguments["show"] == true {
		guid, err := arguments.String("<item_guid>")
		if err != nil {
			panic(err)
		}
		if item, err := storage.Read(guid); err == nil {
			jsonItem, _ := json.MarshalIndent(item, "", "  ")
			fmt.Println(string(jsonItem))
		} else {
			fmt.Println(err)
		}
	}
}

func matchScale(item *strg.Item, scaleSelector string, missingProviderMetadata *int) bool {
	scale := ""
	if scaleSelector == "" {
		return true
	}
	if item.DataProviderMetaJSON == "" {
		*missingProviderMetadata++
	} else {
		switch item.DataProvider {
		case "CBN Polona":
			var polonaObject polona.PolonaObject
			err := json.Unmarshal([]byte(item.DataProviderMetaJSON), &polonaObject)
			if err == nil {
				scale = polonaObject.GetScale()
			}
		}
	}
	if scale == "" || scale == "unknown" {
		scale = item.GetScale()
	}

	scaleSelectorRe := regexp.MustCompile(scaleSelector)
	return scaleSelectorRe.MatchString(scale)
}

func downloadItem(item *strg.Item, outputFolder string, onlyMetadata bool) error {
	switch item.DataProvider {
	case "CBN Polona":
		return polona.DownloadPolona(item, outputFolder, onlyMetadata)
	case "Jagiellońska Biblioteka Cyfrowa":
		return jag.DownloadJag(item, outputFolder, onlyMetadata)
	default:
		fmt.Printf("Data provider %s not supported.\n", item.DataProvider)
		return errors.New("data provider not supported")
	}
}
