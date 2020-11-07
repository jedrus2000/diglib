package storage

import (
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
)

const (
	DEFAULT_SHEET = "Sheet1"
)

type Spreadsheet struct {
	fileName string
	file     *excelize.File
}

func (ss *Spreadsheet) Open(fileName string) {
	ss.fileName = fileName
	var err error
	ss.file, err = excelize.OpenFile(ss.fileName)
	if err != nil {
		ss.file = excelize.NewFile()
		ss.file.NewSheet(DEFAULT_SHEET)
	}
}

func (ss *Spreadsheet) SetHeader(header []string) {
	for i, headerValue := range header {
		ss.file.SetCellValue(DEFAULT_SHEET, string(rune(65+i))+"1", headerValue)
	}
}

func (ss *Spreadsheet) SetData(rowId int, rowData []interface{}) {
	for i, cellValue := range rowData {
		ss.file.SetCellValue(DEFAULT_SHEET, string(rune(65+i))+fmt.Sprintf("%d", rowId), cellValue)
	}
}

func (ss *Spreadsheet) Close() {
	ss.file.SaveAs(ss.fileName)
}
