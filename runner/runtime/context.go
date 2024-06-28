package runtime

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"log"
	"os"
	"path"
	"shadercompat/files"
	"shadercompat/groupedmapping"
	"shadercompat/properties"
	"shadercompat/shader"
	"strings"
)

const UnknownVersion = "?"

type Context struct {
	LocalVersion                                 string
	LocalPath                                    string
	RemoteVersion                                string
	Window                                       fyne.Window
	RepositoryMods                               []string
	ModStates                                    map[string]bool
	VersionView, ShaderView, ModsView, PatchView View
	SelectedFilePath                             string
	OutputFilePath                               string
	SelectedShaderName                           string
	categories                                   shader.Categories
	Shaders                                      map[string]shader.Instance
}

func writeMappings(srcPath string, mappings map[string]map[string][]string, instance shader.Instance) error {
	for typeName := range mappings {
		shaderType := instance.Types[typeName]
		filePath := path.Join(srcPath, shaderType.FilePath)

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return err
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		model, err := properties.Load(data)
		if err != nil {
			return fmt.Errorf("failed to load model from file %s: %v", filePath, err)
		}

		for key, typeMappings := range mappings[typeName] {
			values := make(map[string]bool)
			before, ok := model.Get(key)
			if ok {
				vals := strings.Split(strings.TrimSpace(before), instance.Separator)
				for _, v := range vals {
					values[v] = true
				}
			}

			for _, v := range typeMappings {
				values[v] = true
			}

			value := ""
			i := 0
			for v := range values {
				if i < len(values)-1 {
					value += v + instance.Separator
				} else {
					value += v
				}
				i++
			}
			model.Set(key, value)
		}

		file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		err = model.Write(file)
		if err != nil {
			return fmt.Errorf("failed to write model to file %s: %v", filePath, err)
		}
		err = file.Close()
		if err != nil {
			return fmt.Errorf("failed to close file %s: %v", filePath, err)
		}
	}
	return nil
}

func validateSectionCategories(mappings map[string][]string, categories shader.Categories) error {
	for section := range mappings {
		if !categories.Contains(section) {
			return errors.New("undefined category: " + section)
		}
	}
	return nil
}

func (c *Context) Patch(outputPath string) error {
	sourcePath := c.SelectedFilePath

	// Unzip
	isExtracted := path.Ext(sourcePath) == ".zip"
	if isExtracted {
		newPath := strings.TrimSuffix(sourcePath, path.Ext(sourcePath))
		err := files.Unzip(sourcePath, newPath)
		if err != nil {
			return fmt.Errorf("failed to unzip %s: %v", sourcePath, err)
		}
		sourcePath = newPath

		defer func() {
			err = os.RemoveAll(sourcePath)
			log.Println(err)
		}()
	} else {
		return fmt.Errorf("path %s is not a .zip file", sourcePath)
		//stat, err := os.Stat(sourcePath)
		//if err != nil {
		//	return fmt.Errorf("source path %s does not exist", sourcePath)
		//}
		//if stat.IsDir() {
		//	return fmt.Errorf("source path %s is not a directory", sourcePath)
		//}
	}

	////////////////
	// sourcePath is now destination path
	////////////////

	modMappings, err := c.loadSelectedModMappings()
	if err != nil {
		return err
	}

	if validateSectionCategories(modMappings, c.categories) != nil {
		return err
	}

	transformers := shader.InitTransformers()

	shaders := c.Shaders
	shaderName := c.SelectedShaderName

	instance, ok := shaders[shaderName]
	if !ok {
		return fmt.Errorf("shader %s not found")
	}
	err = shader.Validate(instance, transformers, c.categories)
	if err != nil {
		return fmt.Errorf("failed to validate shader mapping %s: %v", shaderName, err)
	}
	log.Println("Loaded shader", shaderName, "with", len(instance.Mappings), " types.")

	mappings, err := instance.Map(modMappings, transformers)
	if err != nil {
		return err
	}

	// TODO Copy shader source to a new file
	if _, err := os.Stat(outputPath); err == nil {
		err = os.RemoveAll(outputPath)
		if err != nil {
			return err
		}
	}

	err = files.Copy(sourcePath, outputPath)
	if err != nil {
		return err
	}

	err = writeMappings(outputPath, mappings, instance)
	if err != nil {
		return err
	}

	return nil
}

func (c *Context) loadSelectedModMappings() (map[string][]string, error) {
	m := make(map[string][]string)

	for modName, state := range c.ModStates {
		if !state {
			continue
		}

		filePath := path.Join(c.LocalPath, "mods", modName+".gm")
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		err = groupedmapping.Decode(data, m)
		if err != nil {
			return nil, err
		}
	}

	return m, nil
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

func (c *Context) loadModFilePaths() error {
	modsPath := path.Join(c.LocalPath, "mods")
	entries, err := os.ReadDir(modsPath)
	if err != nil {
		return err
	}

	c.ModStates = make(map[string]bool)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fileName := entry.Name()
		if path.Ext(fileName) != ".gm" {
			continue
		}
		c.ModStates[path.Join(modsPath, fileName)] = false
	}

	return nil
}

func (c *Context) LoadMappings() error {
	if err := c.loadCategories(); err != nil {
		return err
	}

	if err := c.loadShaders(); err != nil {
		return err
	}

	if err := c.loadModFilePaths(); err != nil {
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
