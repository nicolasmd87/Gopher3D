package main

import (
	"Gopher3D/renderer"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

var modelChan = make(chan *renderer.Model)
var objectNames []string // Global slice to store object names
var list *widget.List    // List widget to display object names

func main() {
	app := app.New()
	window := app.NewWindow("Gopher 3D")
	gopher := NewGopher()
	window.Resize(fyne.NewSize(1024, 768))

	list = widget.NewList(
		func() int {
			return len(objectNames)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(id widget.ListItemID, co fyne.CanvasObject) {
			co.(*widget.Label).SetText(objectNames[id])
		},
	)

	loadItem := fyne.NewMenuItem("Load Object", func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader == nil {
				return
			}
			if err != nil {
				dialog.ShowError(err, window)
				return
			}

			defer reader.Close()

			filePath := reader.URI().Path()
			model, err := renderer.LoadObjectWithPath(filePath)
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
			modelChan <- model

			objectName := filepath.Base(filePath) // Extract the file name as the object name
			objectNames = append(objectNames, objectName)
			list.Refresh() // Refresh the list to update the UI with the new object name
		}, window)

		fd.SetFilter(storage.NewExtensionFileFilter([]string{".obj"}))
		fd.Show()
	})

	mainMenu := fyne.NewMainMenu(
		fyne.NewMenu("File", loadItem),
	)
	window.SetMainMenu(mainMenu)

	content := container.NewVBox(
		widget.NewLabel("Loaded Objects:"),
		list,
	)

	gap := 50
	go gopher.Render(1024+gap, gap, modelChan)

	window.SetContent(content)
	window.ShowAndRun()
}
