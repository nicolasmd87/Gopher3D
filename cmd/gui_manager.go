package main

import (
	"Gopher3D/renderer"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

var modelChan = make(chan *renderer.Model)

func main() {
	app := app.New()
	window := app.NewWindow("Gopher 3D")
	gopher := NewGopher()
	// Set window size to 1024x768
	window.Resize(fyne.NewSize(1024, 768))

	loadItem := fyne.NewMenuItem("Load Object", func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader == nil {
				// User canceled the dialog
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
		}, window)

		fd.SetFilter(storage.NewExtensionFileFilter([]string{".obj"})) // For example, filter for ".obj" files
		fd.Show()
	})

	// Create a main menu with a "File" drop-down containing the load item
	mainMenu := fyne.NewMainMenu(
		// A File menu
		fyne.NewMenu("File", loadItem),
	)

	window.SetMainMenu(mainMenu)

	// Create a box container (vbox) to place the label.
	box := container.NewVBox(
		widget.NewLabel("Welcome to the Gopher 3D Engine!"),
	)

	gap := 50 // gap in pixels

	go gopher.Render(1024+gap, gap, modelChan)

	window.SetContent(box)

	window.ShowAndRun()
}
