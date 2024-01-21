package main

import (
	"Gopher3D/renderer"
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

var modelChan = make(chan *renderer.Model)
var objectNames []string // Global slice to store object names
var tree *widget.Tree

type GameObject struct {
	Name     string
	Children []*GameObject
}

func NewGameObject(name string) *GameObject {
	return &GameObject{Name: name, Children: []*GameObject{}}
}

func (g *GameObject) AddChild(child *GameObject) {
	g.Children = append(g.Children, child)
}

func (g *GameObject) FindChild(name string) *GameObject {
	if g.Name == name {
		return g
	}
	for _, child := range g.Children {
		found := child.FindChild(name)
		if found != nil {
			return found
		}
	}
	return nil
}

// Root node for your tree
var rootObject *GameObject = NewGameObject("Root")

// RefreshTree updates the tree view
func RefreshTree() {
	tree.Refresh()
}

func AddChildNode(parentName, childName string) {
	parentNode := rootObject.FindChild(parentName)
	if parentNode != nil {
		newNode := NewGameObject(childName)
		parentNode.AddChild(newNode)
		RefreshTree() // Make sure to implement this function to refresh the tree view
	}
}

// AddParentNode adds a new parent node to the tree
func AddParentNode(name string) {
	newNode := NewGameObject(name)
	rootObject.AddChild(newNode)
	RefreshTree()
}

func createTree() *widget.Tree {
	tree := widget.NewTree(
		func(id widget.TreeNodeID) (children []widget.TreeNodeID) {
			fmt.Println("ID: ", id)
			if id == "" {
				// If id is empty, it's the root node
				return []widget.TreeNodeID{"Root"} // Adjust as needed
			}

			node := rootObject.FindChild(id)
			if node == nil || len(node.Children) == 0 {
				return nil
			}
			var childrenIDs []widget.TreeNodeID
			for _, child := range node.Children {
				childrenIDs = append(childrenIDs, child.Name)
			}
			return childrenIDs
		},
		func(id widget.TreeNodeID) bool {
			node := rootObject.FindChild(id)
			return node != nil && len(node.Children) > 0
		},
		func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TreeNodeID, branch bool, o fyne.CanvasObject) {
			if o == nil {
				return // Safeguard against nil canvas object
			}
			label, ok := o.(*widget.Label)
			if !ok || label == nil {
				return // Safeguard against type assertion failure or nil label
			}
			label.SetText(id) // Set the label text to the node ID
		},
	)
	return tree
}

func main() {
	app := app.New()
	window := app.NewWindow("Gopher 3D")
	gopher := NewGopher()
	window.Resize(fyne.NewSize(1024, 768))

	tree = createTree() // Initialize the tree

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

			objectName := filepath.Base(filePath)
			fmt.Println("TREE: ", tree.Root)
			AddParentNode(objectName)

		}, window)

		fd.SetFilter(storage.NewExtensionFileFilter([]string{".obj"}))
		fd.Show()
	})
	view := fyne.NewMenuItem("Debug", func() {
		fmt.Println("Debug: ", renderer.Debug)
		if !renderer.Debug {
			renderer.Debug = true
			return
		}
		renderer.Debug = false
	})

	mainMenu := fyne.NewMainMenu(
		fyne.NewMenu("File", loadItem),
		fyne.NewMenu("View", view),
	)

	window.SetMainMenu(mainMenu)
	header := container.NewVBox(widget.NewLabel("Loaded Objects"), tree)
	content := fyne.NewContainerWithLayout(layout.NewBorderLayout(header, nil, nil, nil), header)

	gap := 50
	go gopher.Render(1024+gap, gap, modelChan)

	window.SetContent(content)
	window.ShowAndRun()
}
