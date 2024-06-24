package files

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

func Unzip(src string, dst string) error {
	err := os.MkdirAll(dst, 0755)
	if err != nil {
		return err
	}

	reader, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer reader.Close()

	extract := func(file *zip.File) error {
		path := filepath.Join(dst, file.Name)
		if file.FileInfo().IsDir() {
			err = os.MkdirAll(path, file.Mode())
			if err != nil {
				return err
			}
			return nil
		}

		readCloser, err := file.Open()
		if err != nil {
			return err
		}
		defer readCloser.Close()

		err = os.MkdirAll(filepath.Dir(path), file.Mode())
		if err != nil {
			return err
		}

		dstFile, err := os.Create(path)
		if err != nil {
			return err
		}

		_, err = io.Copy(dstFile, readCloser)
		if err != nil {
			return err
		}

		return dstFile.Close()
	}

	for _, file := range reader.File {
		err = extract(file)
		if err != nil {
			return err
		}
	}

	return nil
}
