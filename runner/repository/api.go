package repository

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type contentsResponse []file

type file struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Sha         string `json:"sha"`
	Size        int    `json:"size"`
	Url         string `json:"url"`
	HtmlUrl     string `json:"html_url"`
	GitUrl      string `json:"git_url"`
	DownloadUrl string `json:"download_url"`
	Type        string `json:"type"`
	Links       struct {
		Self string `json:"self"`
		Git  string `json:"git"`
		Html string `json:"html"`
	} `json:"_links"`
}

func getGitHubContents(owner, repo, path string) (contentsResponse, error) {
	get, err := http.Get("https://api.github.com/repos/" + owner + "/" + repo + "/contents/" + path)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(get.Body)
	if err != nil {
		return nil, err
	}
	var response contentsResponse
	err = json.Unmarshal(data, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v\n%s\n", err, string(data))
	}
	return response, nil
}
