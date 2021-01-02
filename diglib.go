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
	"strings"
	"syscall"
)

func main() {
	usage := `
diglib
	
	Usage:
	  diglib search [--page-size=<value>]
	  diglib show [--item-selector=<item_guid>] [--download-selector=<value>] [--scale-selector=<value>] 
			[--library-selector=<value>] [--find-everywhere-selector=<value>] [--output-excel-filename=<excel_filename>]
	  diglib download [--item-selector=<item_guid>] [--download-selector=<value>] [--scale-selector=<value>] 
			[--library-selector=<value>] [--find-everywhere-selector=<value>]
			[--dry-run] [--output-folder=<value>] [--only-metadata]
	  diglib set-property [--single=<guid>] [--download-selector=<value>]		
	  diglib -h | --help
  	  diglib --version

	Options:
	  --output-folder=<value>         Output folder [default: ./downloads].
	  --page-size=<value>             Page size of results [default: 1000].
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

	if arguments["show"] == true {
		outputExcelFilename, _ := arguments.String("--output-excel-filename")
		if outputExcelFilename != "" {
			ss := &strg.Spreadsheet{}
			ss.Open(outputExcelFilename)
			ss.SetDefaultHeader()
			defer ss.Close()
			rowId := 2
			selectItemsBySelectors(&arguments, storage, func(item *strg.Item) {
				ss.SetDefaultData(rowId, item)
				rowId++
			})
		} else {
			selectItemsBySelectors(&arguments, storage, func(item *strg.Item) {
				jsonItem, _ := json.MarshalIndent(item, "", "  ")
				fmt.Println(string(jsonItem))
			})
		}
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

		selectItemsBySelectors(&arguments, storage, func(item *strg.Item) {
			if dryRun == true {
				fmt.Printf("%s, %s %s \n", item.Guid, item.Title, item.DataProvider)
			} else if err := downloadItem(item, outputFolder, onlyMetadata); err == nil {
				storage.SaveItem(item, true)
			}
		})

	}
}

func selectItemsBySelectors(arguments *docopt.Opts, storage *strg.Storage, processItem func(item *strg.Item)) {
	var downloadSelectorRe, librarySelectorRe, itemSelectorRe, findEverywhereSelectorRe *regexp.Regexp

	findEverywhereSelector, _ := arguments.String("--find-everywhere-selector")
	findEverywhereSelectorRe = regexp.MustCompile(findEverywhereSelector)

	itemSelector, _ := arguments.String("--item-selector")
	itemSelectorRe = regexp.MustCompile(itemSelector)

	downloadSelector, _ := arguments.String("--download-selector")
	downloadSelectorRe = regexp.MustCompile(downloadSelector)

	scaleSelector, _ := arguments.String("--scale-selector")

	librarySelector, _ := arguments.String("--library-selector")
	librarySelectorRe = regexp.MustCompile(librarySelector)

	foundCounter := 0
	missingCnt := 0
	missingProviderMetadataCounter := 0

	onItem := func(item *strg.Item) {
		if downloadSelectorRe.MatchString(item.Download) &&
			librarySelectorRe.MatchString(item.DataProvider) &&
			itemSelectorRe.MatchString(item.Guid) &&
			matchAnyField(item, findEverywhereSelectorRe) &&
			matchScale(item, scaleSelector, &missingProviderMetadataCounter) {
			fmt.Fprint(os.Stdout, "\r \r")
			foundCounter++
			processItem(item)
		} else {
			fmt.Fprintf(os.Stdout, "\r%s\r", string(`-\|/`[missingCnt%4]))
			missingCnt++
		}
	}

	singleItem, err := storage.Read(itemSelector)
	if err == nil {
		onItem(&singleItem)
	} else {
		storage.ForEach(onItem)
	}

	fmt.Printf("Found %d items. %d have missing data provider metadata that can be useful with provided selectors.\n", foundCounter, missingProviderMetadataCounter)
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

func matchAnyField(item *strg.Item, findEverywhereSelectorRe *regexp.Regexp) bool {
	if findEverywhereSelectorRe.String() == "" {
		return true
	}
	return findEverywhereSelectorRe.MatchString(item.DataProviderMetaJSON) ||
		findEverywhereSelectorRe.MatchString(item.Description) ||
		findEverywhereSelectorRe.MatchString(item.Contributor) ||
		findEverywhereSelectorRe.MatchString(item.Subject) ||
		findEverywhereSelectorRe.MatchString(strings.Join(item.Title, "")) ||
		findEverywhereSelectorRe.MatchString(item.Format) ||
		findEverywhereSelectorRe.MatchString(item.Publisher) ||
		findEverywhereSelectorRe.MatchString(item.Rights) ||
		findEverywhereSelectorRe.MatchString(item.Source) ||
		findEverywhereSelectorRe.MatchString(strings.Join(item.Type, ""))
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
