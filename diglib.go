package main

import (
	"diglib/polona"
	strg "diglib/storage"
	"fmt"
	"github.com/docopt/docopt-go"
)

func main() {
	storage := &strg.Storage{}
	storage.Open()
	defer storage.Close()

	usage := `
digilib
	
	Usage:
	  digilib search [--page-size=<ps>]
	  digilib dump <output_filename>
      digilib download [--single=<guid>]
	  digilib -h | --help
  	  digilib --version

	Options:
	  --page-size=<ps>  			Page size of results [default: 1000].
      --single=<guid>				ID (guid) of element.
	  -h --help                     Show this screen.
  	  --version                     Show version.	

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
		guid, err := arguments.String("--single")
		if err == nil {
			item := storage.Find(guid)
			fmt.Println(item)
			polona.DownloadPolona(&item)
		}
	}
}
