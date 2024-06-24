package grouped_mapping

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
)

func DecodeFiles(root string, paths ...string) (map[string][]string, error) {
	m := make(map[string][]string)
	for _, p := range paths {
		data, err := os.ReadFile(path.Join(root, p))
		if err != nil {
			return nil, errors.Join(errors.New("failed to decode file "+p+":"), err)
		}
		err = Decode(data, m)
		if err != nil {
			return nil, errors.Join(errors.New("failed to decode file "+p+":"), err)
		}
	}
	return m, nil
}

func Decode(data []byte, m map[string][]string) error {
	section := ""
	namespace := ""

	lines := strings.Split(string(data), "\n")

	for i, line := range lines {
		lines[i] = strings.ReplaceAll(line, "\r", "")
	}

	for i, line := range lines {
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		if line[0] == '[' {
			if !strings.HasSuffix(line, "]") {
				return errors.New(fmt.Sprintf("missing ] in section header in line %d", i+1))
			}
			section = line[1 : len(line)-1]
			if len(section) == 0 {
				return errors.New(fmt.Sprintf("empty section header in line %d", i+1))
			}
			continue
		}
		if line[0] == '$' {
			namespace = line[1:]
			if len(namespace) == 0 {
				return errors.New(fmt.Sprintf("empty namespace in line %d", i+1))
			}
			continue
		}
		if len(section) == 0 {
			return errors.New(fmt.Sprintf("missing section header in line %d", i+1))
		}
		m[section] = append(m[section], namespace+line)
	}
	return nil
}

func CountEntries(mappings map[string][]string) int {
	count := 0
	for _, entries := range mappings {
		count += len(entries)
	}
	return count
}
