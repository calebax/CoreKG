package ziputil

import (
	"archive/zip"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func IsZip(src string) (bool, error) {
	f, err := os.Stat(src)
	if err != nil {
		return false, nil
	}

	if f.IsDir() {
		return false, nil
	}

	return ExtZip == filepath.Ext(f.Name()), nil

}

func IsDir(src string) (bool, error) {
	f, err := os.Stat(src)
	if err != nil {
		return false, err
	}

	return f.IsDir(), nil
}

func IsFileExist(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

func GetSrcZips(src string) ([]string, error) {
	files, err := ioutil.ReadDir(src)
	if err != nil {
		return nil, err
	}
	fileNames := []string{}
	for _, f := range files {
		ext := filepath.Ext(f.Name())
		if ext != ExtZip {
			continue
		}

		fileNames = append(fileNames, strings.TrimSuffix(f.Name(), ext))
	}

	return fileNames, nil
}

func Compress(srcDir, dstZipPath string) error {
	if filepath.Ext(filepath.Base(dstZipPath)) != ExtZip {
		return errors.New("not a zip file")
	}

	dstDir := filepath.Dir(dstZipPath)
	isDstDir, err := IsDir(dstDir)
	if err != nil {
		return err
	}
	if !isDstDir {
		return errors.New("dstDir is not a directory")
	}

	isSrcDir, err := IsDir(srcDir)
	if err != nil {
		return err
	}
	if !isSrcDir {
		return errors.New("srcDirs is not a directory")
	}

	f, err := os.Create(dstZipPath)
	if err != nil {
		return err
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	files, err := ioutil.ReadDir(srcDir)
	if err != nil {
		return err
	}

	for _, fi := range files {
		if err := compress(fi, srcDir, "", zw); err != nil {
			return err
		}
	}

	return nil
}

func compress(fi os.FileInfo, fileDir string, subName string, zw *zip.Writer) error {
	if fi.IsDir() {
		fileDir := filepath.Join(subName, fi.Name())
		if subName != "" {
			subName = filepath.Join(subName, fi.Name())
		} else {
			subName = fi.Name()
		}

		header, err := zip.FileInfoHeader(fi)
		if err != nil {
			return err
		}
		header.Name = subName + "/"

		_, err = zw.CreateHeader(header)
		if err != nil {
			return err
		}

		files, err := ioutil.ReadDir(fileDir)
		if err != nil {
			return err
		}
		for _, fi := range files {
			if err := compress(fi, fileDir, subName, zw); err != nil {
				return err
			}
		}
	} else {
		filePath := filepath.Join(fileDir, fi.Name())
		f, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer f.Close()

		header, err := zip.FileInfoHeader(fi)
		if err != nil {
			return err
		}
		if subName != "" {
			header.Name = filepath.Join(subName, fi.Name())
		}
		writer, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, f)
		if err != nil {
			return err
		}
	}

	return nil
}
