package shader

import (
	"encoding/json"
	"fmt"
	"os"
)

type Transformer interface {
	Generate(entries []string) []string
}

type Mapping struct {
	To string `json:"to"`
	// Transformer is optional. Name of the generator to use.
	Transformer string `json:"transformer,omitempty"`
}

type Type struct {
	FilePath string `json:"file_path"`
}

type Instance struct {
	Name      string                          `json:"name"`
	Separator string                          `json:"separator"`
	Types     map[string]Type                 `json:"types"`
	Mappings  map[string]map[string][]Mapping `json:"shaders"`
}

func (s *Instance) Load(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	return nil
}

func (s *Instance) Map(entries map[string][]string, transformers map[string]Transformer) (map[string]map[string][]string, error) {
	m := make(map[string]map[string][]string)
	mappings := s.Mappings

	for typeName := range s.Types {
		m[typeName] = make(map[string][]string)
	}

	for typeName := range mappings {
		if _, ok := s.Types[typeName]; !ok {
			return nil, fmt.Errorf("unknown type '%s' in shaders", typeName)
		}

		for category, mappings := range mappings[typeName] {
			for _, mapping := range mappings {
				if _, ok := entries[mapping.To]; ok {
					return nil, fmt.Errorf("mapping '%s' was already mapped to before", mapping.To)
				}
				transformer, ok := transformers[mapping.Transformer]
				if !ok {
					// This should never happen, since we validate the shaders before.
					return nil, fmt.Errorf("unknown transformer '%s' for mapping '%s'", mapping.Transformer, mapping.To)
				}

				transformedEntries := transformer.Generate(entries[category])
				m[typeName][mapping.To] = transformedEntries
			}
		}
	}

	return m, nil
}

type Categories struct {
	List []string `json:"categories"`
}

func (c *Categories) Load(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &c)
	if err != nil {
		return err
	}
	return nil
}

func (c *Categories) Contains(name string) bool {
	for _, v := range c.List {
		if v == name {
			return true
		}
	}
	return false
}
