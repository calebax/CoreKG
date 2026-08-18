package utils

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/insmtx/corekg/pkgs/utils/httptools"
)

type FileType string

const (
	FileTypeUnknown FileType = "unknown"
	FileTypeXlsx    FileType = "xlsx"
	FileTypeXls     FileType = "xls"
	FileTypeCsv     FileType = "csv"
	FileTypePDF     FileType = "pdf"
	FileTypeText    FileType = "text"
	FileTypeImage   FileType = "image"
	FileTypeVideo   FileType = "video"
	FileTypeBinary  FileType = "binary"
)

// OpenFromPathOrURL opens a local path or URL using a background context.
func OpenFromPathOrURL(pathOrURL string) (io.ReadCloser, string, FileType, error) {
	return OpenFile(context.Background(), pathOrURL)
}

// OpenFile opens a local path or URL and returns a replayable stream with its detected file metadata.
func OpenFile(ctx context.Context, source string) (io.ReadCloser, string, FileType, error) {
	if source == "" {
		return nil, "", FileTypeUnknown, fmt.Errorf("路径或URL不能为空")
	}

	u, err := url.Parse(source)
	if err == nil && u.Scheme != "" {
		switch strings.ToLower(u.Scheme) {
		case "http", "https":
			resp, err := httptools.Get(ctx, source)
			if err != nil {
				return nil, "", FileTypeUnknown, fmt.Errorf("下载远程文件失败: %w", err)
			}

			fileName := ""
			if cd := resp.Header.Get("Content-Disposition"); cd != "" {
				if _, params, err := mime.ParseMediaType(cd); err == nil {
					if fn, ok := params["filename"]; ok && fn != "" {
						fileName = fn
					}
				}
			}
			if fileName == "" {
				fileName = path.Base(u.Path)
			}
			return inspectFile(resp.Body, fileName)

		case "file":
			localPath := filepath.FromSlash(u.Path)
			if localPath == "" {
				return nil, "", FileTypeUnknown, fmt.Errorf("无效的 file URL: %s", source)
			}

			f, err := os.Open(localPath)
			if err != nil {
				return nil, "", FileTypeUnknown, fmt.Errorf("打开本地文件失败: %v", err)
			}
			return inspectFile(f, filepath.Base(localPath))

		default:
			return nil, "", FileTypeUnknown, fmt.Errorf("不支持的URL协议: %s", u.Scheme)
		}
	}

	f, err := os.Open(source)
	if err != nil {
		return nil, "", FileTypeUnknown, fmt.Errorf("打开本地文件失败: %v", err)
	}
	return inspectFile(f, filepath.Base(source))
}

func inspectFile(reader io.ReadCloser, filename string) (io.ReadCloser, string, FileType, error) {
	buf := make([]byte, 512)
	n, err := reader.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		reader.Close()
		return nil, "", FileTypeUnknown, fmt.Errorf("读取文件头失败: %w", err)
	}
	fileType := detectFileType(filename, buf[:n])
	return &replayReadCloser{
		Reader: io.MultiReader(bytes.NewReader(buf[:n]), reader),
		Closer: reader,
	}, filename, fileType, nil
}

// replayReadCloser replays inspected header bytes while retaining ownership of the underlying stream.
type replayReadCloser struct {
	io.Reader
	io.Closer
}

// 针对文本文件，读取行内容
func ReadLines(r io.Reader) ([]string, error) {
	return ReadLinesRange(r, 1, -1)
}

func ReadLinesRange(r io.Reader, startLine, endLine int) ([]string, error) {
	if startLine < 0 {
		return nil, fmt.Errorf("startLine must be >= 0")
	} else if startLine == 0 {
		startLine = 1
	}
	if endLine != -1 && endLine < startLine {
		return nil, fmt.Errorf("endLine must be -1 or >= startLine")
	}

	scanner := bufio.NewScanner(r)
	lines := []string{}
	current := 0

	for scanner.Scan() {
		current++

		if current < startLine {
			continue
		}
		if endLine != -1 && current > endLine {
			break
		}

		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func detectFileType(filename string, peekBytes []byte) FileType {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".txt", ".md", ".json", ".xml", ".yaml", ".yml", ".ini", ".cfg":
		return FileTypeText
	case ".xlsx":
		return FileTypeXlsx
	case ".xls":
		return FileTypeXls
	case ".csv":
		return FileTypeCsv
	case ".pdf":
		return FileTypePDF
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".bmp":
		return FileTypeImage
	case ".mp4", ".mov", ".avi", ".mkv", ".webm":
		return FileTypeVideo
	}

	// 扩展名判断不出 → MIME sniffing
	contentType := http.DetectContentType(peekBytes)

	if strings.HasPrefix(contentType, "text/") {
		return FileTypeText
	}
	if strings.HasPrefix(contentType, "image/") {
		return FileTypeImage
	}
	if strings.HasPrefix(contentType, "video/") {
		return FileTypeVideo
	}
	if strings.Contains(contentType, "pdf") {
		return FileTypePDF
	}

	// Excel MIME 类型
	if strings.Contains(contentType, "spreadsheetml.sheet") {
		return FileTypeXlsx
	}
	if strings.Contains(contentType, "ms-excel") {
		return FileTypeXls
	}

	// 兜底：二进制
	if strings.Contains(contentType, "application/") {
		return FileTypeBinary
	}

	return FileTypeText
}
