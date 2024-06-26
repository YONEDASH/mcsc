package runtime

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"log"
	"os"
	"path"
	"shadercompat/shader"
)

const UnknownVersion = "?"

type Context struct {
	LocalVersion            string
	LocalPath               string
	RemoteVersion           string
	Window                  fyne.Window
	RepositoryMods          []string
	Shaders                 map[string]shader.Instance
	ShaderNames             []string
	categories              shader.Categories
	VersionView, ShaderView View
	SelectedFilePath        string
}

func (c *Context) loadCategories() error {
	categories := shader.Categories{}
	err := categories.Load(c.LocalPath + "/global/category.json")
	c.categories = categories
	return err
}

func (c *Context) loadShaders() error {
	dirPath := c.LocalPath + "/shaders"
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	m := make(map[string]shader.Instance)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := path.Join(dirPath, file.Name())
		if path.Ext(filePath) != ".json" {
			return fmt.Errorf("file %s is not a json file", filePath)
		}

		i := shader.Instance{}
		err = i.Load(filePath)
		if err != nil {
			return err
		}

		m[i.Name] = i
	}

	c.Shaders = m

	return nil
}

func (c *Context) LoadMappings() error {
	if err := c.loadCategories(); err != nil {
		return err
	}

	if err := c.loadShaders(); err != nil {
		return err
	}

	return nil
}

func wrapText(text string) string {
	txt := ""
	n := 64
	for i, r := range text {
		txt += string(r)
		if i >= n && i%n == 0 {
			txt += "\n"
		}
	}
	return txt
}

func wrapError(err error) error {
	return fmt.Errorf("%s", wrapText(fmt.Sprintf("%v", err)))
}

func (c *Context) ShowError(err error) {
	log.Printf("Error: %v", err)
	dialog.ShowError(wrapError(err), c.Window)
}

func (c *Context) ShowFatal(err error) {
	log.Printf("Fatal: %v", err)
	dialog.ShowCustomWithoutButtons("Fatal Error", container.NewVBox(
		widget.NewLabel(fmt.Sprintf("%v", wrapError(err))), widget.NewButton("Exit", func() {
			os.Exit(1)
		})), c.Window)
}
