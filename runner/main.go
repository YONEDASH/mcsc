package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"shadercompat/files"
	"shadercompat/groupedmapping"
	"shadercompat/properties"
	"shadercompat/shader"
	"strings"
)

var (
	shaderMappingsPath = flag.String("s", "", "directory containing shaders")
	modMappingsPath    = flag.String("m", "", "directory containing grouped mappings of mods")
	categoriesPath     = flag.String("c", "", "file containing categories")
	shaderName         = flag.String("shader", "", "shader name based on name set in shaders")
	sourcePath         = flag.String("source", "", "shader source directory path OR zip file path")
)

func handleFatal(err error) {
	if err != nil {
		log.Fatalf("an error occurred during execution: %v", err)
	}
}

func main() {
	flag.Parse()

	if len(*shaderMappingsPath) == 0 || len(*modMappingsPath) == 0 || len(*categoriesPath) == 0 || len(*shaderName) == 0 || len(*sourcePath) == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Unzip
	isExtracted := path.Ext(*sourcePath) == ".zip"
	if isExtracted {
		newPath := strings.TrimSuffix(*sourcePath, path.Ext(*sourcePath))
		err := files.Unzip(*sourcePath, newPath)
		if err != nil {
			handleFatal(fmt.Errorf("failed to unzip %s: %v", *sourcePath, err))
		}
		*sourcePath = newPath

		defer func() {
			err = os.RemoveAll(*sourcePath)
			handleFatal(err)
		}()
	} else {
		stat, err := os.Stat(*sourcePath)
		if err != nil {
			handleFatal(fmt.Errorf("source path %s does not exist", *sourcePath))
		}
		if stat.IsDir() {
			handleFatal(fmt.Errorf("source path %s is not a directory", *sourcePath))
		}
	}

	categories, err := loadCategories()
	handleFatal(err)

	modMappings, err := loadModMappings()
	handleFatal(err)
	err = validateSectionCategories(modMappings, categories)
	handleFatal(err)

	log.Println("Loaded", len(modMappings), "mods with", groupedmapping.CountEntries(modMappings), "entries.")

	transformers := shader.InitTransformers()

	shaders, err := loadShaders()
	handleFatal(err)

	instance, ok := shaders[*shaderName]
	if !ok {
		handleFatal(fmt.Errorf("shader %s not found", *shaderName))
	}
	err = shader.Validate(instance, transformers, categories)
	if err != nil {
		handleFatal(fmt.Errorf("failed to validate shader mapping %s: %v", *shaderName, err))
	}
	log.Println("Loaded shader", *shaderName, "with", len(instance.Mappings), "shaders.")

	mappings, err := instance.Map(modMappings, transformers)
	handleFatal(err)

	// TODO Copy shader source to a new file
	outputPath := path.Join(".", fmt.Sprintf("%s_generated", *shaderName))
	if _, err := os.Stat(outputPath); err == nil {
		err = os.RemoveAll(outputPath)
		handleFatal(err)
	}

	err = files.Copy(*sourcePath, outputPath)
	handleFatal(err)

	err = writeMappings(outputPath, mappings, instance)
	handleFatal(err)

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

func loadCategories() (shader.Categories, error) {
	categories := shader.Categories{}
	err := categories.Load(*categoriesPath)
	return categories, err
}

func loadShaders() (map[string]shader.Instance, error) {
	dirPath := *shaderMappingsPath
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	m := make(map[string]shader.Instance)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := path.Join(dirPath, file.Name())
		if path.Ext(filePath) != ".json" {
			return nil, fmt.Errorf("file %s is not a json file", filePath)
		}

		i := shader.Instance{}
		err = i.Load(filePath)
		if err != nil {
			return nil, err
		}

		m[i.Name] = i
	}

	return m, nil
}

func loadModMappings() (map[string][]string, error) {
	dirPath := *modMappingsPath
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	m := make(map[string][]string)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := path.Join(dirPath, file.Name())
		if path.Ext(filePath) != ".gm" {
			return nil, fmt.Errorf("file %s is not a grouped mapping (.gm) file", filePath)
		}

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

func validateSectionCategories(mappings map[string][]string, categories shader.Categories) error {
	for section := range mappings {
		if !categories.Contains(section) {
			return errors.New("undefined category: " + section)
		}
	}
	return nil
}
