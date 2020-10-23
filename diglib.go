package main

import (
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/uitools"
	"github.com/therecipe/qt/widgets"
)

func main() {
	// widgets.NewQApplication(len(os.Args), os.Args)

	// MainWindow().Show()

	// widgets.QApplication_Exec()

	Search()
}

func MainWindow() *widgets.QWidget {
	file := core.NewQFile2("./resources/main_window.ui")
	file.Open(core.QIODevice__ReadOnly)
	formWidget := uitools.NewQUiLoader(nil).Load(file, nil)

	return formWidget
}
