package repository

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
)

func Clone(localPath string) error {
	modsResp, err := getGitHubContents(owner, repository, "mods")
	if err != nil {
		return err
	}
	err = writeFiles(localPath+"/mods", modsResp)
	if err != nil {
		return err
	}

	shadersResp, err := getGitHubContents(owner, repository, "shaders")
	if err != nil {
		return err
	}
	err = writeFiles(localPath+"/shaders", shadersResp)
	if err != nil {
		return err
	}

	globalResp, err := getGitHubContents(owner, repository, "global")
	if err != nil {
		return err
	}
	err = writeFiles(localPath+"/global", globalResp)
	if err != nil {
		return err
	}

	return nil
}

func Update(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to remove old repository: %v", err)
	}

	return nil
}

func readBytesFromUrl(url string) ([]byte, error) {
	get, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(get.Body)
}

func writeBytes(p string, data []byte) error {
	err := os.MkdirAll(path.Dir(p), 0755)
	if err != nil {
		return err
	}
	file, err := os.Create(p)
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	return err
}

func writeFiles(dir string, files []file) error {
	for _, file := range files {
		if file.Type != "file" {
			continue
		}
		data, err := readBytesFromUrl(file.DownloadUrl)
		if err != nil {
			return err
		}
		err = writeBytes(dir+"/"+file.Name, data)
		if err != nil {
			return err
		}
	}
	return nil
}
