package files

import (
	"io"
	"os"
	"path"
)

func Copy(srcPath, dstPath string) error {
	stat, err := os.Stat(srcPath)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		return CopyDir(srcPath, dstPath)
	}
	return CopyFile(srcPath, dstPath)
}

func CopyDir(srcDirPath, dstDirPath string) error {
	src, err := os.ReadDir(srcDirPath)
	if err != nil {
		return err
	}

	err = os.MkdirAll(dstDirPath, 0755)
	if err != nil {
		return err
	}

	for _, file := range src {
		srcFilePath := path.Join(srcDirPath, file.Name())
		dstFilePath := path.Join(dstDirPath, file.Name())

		if file.IsDir() {
			err = CopyDir(srcFilePath, dstFilePath)
			if err != nil {
				return err
			}
		} else {
			err = CopyFile(srcFilePath, dstFilePath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func CopyFile(srcFilePath, dstFilePath string) error {
	src, err := os.Open(srcFilePath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(dstFilePath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)

	return err
}
