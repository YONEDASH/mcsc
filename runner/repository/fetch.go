package repository

/*
func FetchCategories() (shader.Categories, error) {
	var empty shader.Categories

	contents, err := github.getGitHubContents(owner, repository, "global")
	if err != nil {
		return empty, fmt.Errorf("failed to fetch global categories: %v", err)
	}

	if len(contents) != 1 {
		return empty, fmt.Errorf("expected exactly one file in global, got %d", len(contents))
	}

	categoriesFile := contents[0]

	data, err := github.readBytesFromUrl(categoriesFile.DownloadUrl)
	if err != nil {
		return empty, fmt.Errorf("failed to read global categories: %v", err)
	}

	var categories shader.Categories
	err = json.Unmarshal(data, &categories)
	return categories, err
}

func FetchShaders() (map[string]shader.Instance, error) {
	contents, err := github.getGitHubContents(owner, repository, "shaders")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch shaders: %v", err)
	}

	shaders := make(map[string]shader.Instance)
	for _, file := range contents {
		data, err := github.readBytesFromUrl(file.DownloadUrl)
		if err != nil {
			return nil, fmt.Errorf("failed to read shader %s: %v", file.Name, err)
		}

		var instance shader.Instance
		err = json.Unmarshal(data, &instance)
		if err != nil {
			return nil, fmt.Errorf("failed to parse shader %s: %v", file.Name, err)
		}

		shaders[instance.Name] = instance
	}

	return shaders, nil
}

func LoadModFiles() ([]github.file, error) {
	contents, err := github.getGitHubContents(owner, repository, "mods")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch mods: %v", err)
	}
	return contents, nil
}

*/
