package shader

import (
	"errors"
	"fmt"
)

func Validate(instance Instance, transformers map[string]Transformer, categories Categories) error {
	for typeName, mapping := range instance.Mappings {
		if _, ok := instance.Types[typeName]; !ok {
			return fmt.Errorf("undefined type '%s' in shaders", typeName)
		}
		for category, typeMappings := range mapping {
			if !categories.Contains(category) {
				return errors.New("undefined category: " + category)
			}
			for _, mapping := range typeMappings {
				if err := validateMapping(mapping, transformers, categories); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func validateMapping(mapping Mapping, transformers map[string]Transformer, categories Categories) error {
	if _, ok := transformers[mapping.Transformer]; !ok {
		return fmt.Errorf("unknown transformer '%s' for mapping '%s'", mapping.Transformer, mapping.To)
	}
	return nil
}
