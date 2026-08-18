package ziputil

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	ExtZip = ".zip"
	// maxFileCount 压缩文件最大文件数
	maxFileCount = 1500
	// singleFileMaxSize 单个文件解压和总文件解压不能超过1G
	singleFileMaxSize uint64 = 1073741824
	fileMaxSize       int64  = 1073741824
)

func UnCompress(src, dst string) error {
	isZip, err := IsZip(src)
	if err != nil {
		return err
	}
	if !isZip {
		return errors.New("src is not a zip file")
	}

	isDir, err := IsDir(dst)
	if err != nil {
		return err
	}
	if !isDir {
		return errors.New("dst is not a directory")
	}

	srcReader, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer srcReader.Close()

	if len(srcReader.File) > maxFileCount {
		return errors.New("zip file have too many files")
	}

	for _, f := range srcReader.File {
		// safety requirements
		if strings.Contains(f.Name, "./") ||
			strings.Contains(f.Name, ".\\") ||
			strings.Contains(f.Name, "..") {
		}
		if f.UncompressedSize64 > singleFileMaxSize {
			return errors.New("single file exceeds the maximum size")
		}
	}
	err = writeReader(srcReader, dst)
	if err != nil {
		return err
	}

	return nil
}

func writeReader(srcReader *zip.ReadCloser, dst string) error {
	var totalSize int64
	for _, f := range srcReader.File {
		if totalSize > fileMaxSize {
			return errors.New("the total size has exceeds the upper limit")
		}
		fileName := f.Name
		targetFilePath := filepath.Join(dst, fileName)

		if f.FileInfo().IsDir() {
			err := os.MkdirAll(targetFilePath, f.Mode())
			if err != nil {
				return err
			}
			continue
		}

		isExist, err := IsFileExist(targetFilePath)
		if err != nil {
			return err
		}
		if isExist {
			return errors.New("the target path has exist")
		}

		if err := os.MkdirAll(path.Dir(targetFilePath), 0700); err != nil {
			return err
		}

		writeSize, err := writeFile(f, targetFilePath)
		if err != nil {
			return err
		}
		totalSize += writeSize
	}

	return nil
}

func writeFile(f *zip.File, targetFilePath string) (int64, error) {
	targetFile, err := os.Create(targetFilePath)
	if err != nil {
		return 0, err
	}
	defer targetFile.Close()
	file, err := f.Open()
	if err != nil {
		return 0, err
	}
	defer file.Close()

	writeSize, err := io.Copy(targetFile, file)
	if err != nil {
		return 0, err
	}

	return writeSize, nil
}
