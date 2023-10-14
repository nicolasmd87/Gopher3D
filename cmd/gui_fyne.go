package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("3D Engine")
	g := NewGopher()
	// Set window size to 1024x768
	w.Resize(fyne.NewSize(1024, 768))

	// Menu item for loading object
	loadItem := &fyne.MenuItem{Label: "Load Object", Action: func() {
		// Logic to load objects into your 3D engine.
		dialog.ShowInformation("Info", "Object loaded!", w)
	}}

	// Create a main menu with a "File" drop-down containing the load item
	mainMenu := fyne.NewMainMenu(
		// A File menu
		fyne.NewMenu("File", loadItem),
	)

	w.SetMainMenu(mainMenu)

	// Create a box container (vbox) to place the label.
	box := container.NewVBox(
		widget.NewLabel("Welcome to the Gopher 3D Engine!"),
	)
	go g.Render()
	w.SetContent(box)
	w.ShowAndRun()

	// Use a goroutine to run the renderer, so it doesn't block the main thread

}
