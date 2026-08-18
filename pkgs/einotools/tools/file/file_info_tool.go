package file

import (
	"context"
	"fmt"
	"mime"
	"net/http"
	neturl "net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	tool "github.com/cloudwego/eino/components/tool"
	toolutils "github.com/cloudwego/eino/components/tool/utils"
	agutils "github.com/insmtx/corekg/pkgs/einotools/utils"
)

// FileInfoConfig defines configuration for the file info tool.
type FileInfoConfig struct {
	ToolName string `json:"tool_name"` // default: file_info_tool
	ToolDesc string `json:"tool_desc"` // default: Get basic info of any file
}

// FileInfoRequest defines the input for retrieving file info.
type FileInfoRequest struct {
	// FilePath string `json:"file_path" jsonschema:"description=Local file system path. Mutually exclusive with 'fileUrl'. Exactly one of 'filePath' or 'fileUrl' must be provided. ,omitempty"`
	FileUrl string `json:"file_url"  jsonschema:"description=Remote URL of the file. Mutually exclusive with 'filePath'. Exactly one of 'filePath' or 'fileUrl' must be provided. ,omitempty"`
}

type FileInfo struct {
	Name     string   `json:"name"`
	Size     int64    `json:"size"`
	Ext      string   `json:"ext"`
	Path     string   `json:"path"`
	Type     string   `json:"type"` // "excel", "csv", "text", "image", "generic"...
	MimeType string   `json:"mime_type"`
	Extra    any      `json:"extra,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

type fileInfoTool struct {
	conf *FileInfoConfig
}

// NewFileInfoTool creates an InvokableTool that returns Excel file metadata using utils.GetFileInfo.
func NewFileInfoTool(ctx context.Context, conf *FileInfoConfig) (tool.InvokableTool, error) {
	if conf == nil {
		conf = &FileInfoConfig{}
	}
	toolName := conf.ToolName
	toolDesc := conf.ToolDesc
	if toolName == "" {
		toolName = "file_inspect_tool"
	}
	if toolDesc == "" {
		toolDesc = `在进行任何文件级推理之前，用于获取文件的元信息并推断其结构。

本工具仅用于“轻量级结构检查”，包括但不限于：
- 文件类型、大小、编码方式
- 可用的 Sheet / 顶层字段
- 列名、字段名及其推断的数据类型
- 少量数据样例（仅用于了解结构）

重要说明（表格文件）：
- 对于 Excel / 表格类文件，日期字段可能以原始数值形式返回
  （例如 44986），表示 Excel 的内部日期序列值。
- 如需将该数值解释为具体日期，应在后续推理或处理阶段显式转换，
  不得在读取阶段擅自假定其格式。

使用规则：
- 不允许读取完整数据或执行分析计算
- 不对字段含义做业务层面的推断
- 当文件结构尚不明确时，必须优先使用本工具

`
	}

	fit := &fileInfoTool{conf: conf}
	tl, err := toolutils.InferTool(toolName, toolDesc, fit.invoke)
	if err != nil {
		return nil, err
	}
	return tl, nil
}

// invoke is the core logic to get file info.
func (t *fileInfoTool) invoke(ctx context.Context, req *FileInfoRequest) (*FileInfo, error) {
	if req == nil || (req.FileUrl == "") {
		return nil, fmt.Errorf("input cannot be nil")
	}

	source := req.FileUrl
	// if source == "" {
	// 	source = req.FilePath
	// }

	info, err := getFileInfo(source)
	if err != nil {
		return nil, fmt.Errorf("failed to get base file info: %w", err)
	}

	switch info.Ext {
	case ".xlsx":
		extra, err := agutils.GetXlsFileInfo(source)
		if err != nil {
			return nil, fmt.Errorf("failed to get excel file info: %w", err)
		}
		info.Extra = extra
	case ".pdf":
		extra, err := agutils.GetPDFFileInfo(source)
		if err != nil {
			return nil, fmt.Errorf("failed to get pdf file info: %w", err)
		}
		info.Extra = extra
	}

	return info, nil
}

func getFileInfo(source string) (*FileInfo, error) {
	// 尝试解析为 URL
	if u, err := neturl.Parse(source); err == nil && u.Scheme != "" {
		switch strings.ToLower(u.Scheme) {
		case "http", "https":
			// 远程文件：尽量使用 HEAD 获取基础信息
			req, err := http.NewRequest("HEAD", source, nil)
			if err != nil {
				return nil, fmt.Errorf("create HEAD request failed: %v", err)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil || resp.StatusCode < 200 || resp.StatusCode >= 300 {
				// HEAD 不可用时，退回 GET（不读取主体）
				if resp != nil {
					resp.Body.Close()
				}
				resp, err = http.Get(source)
				if err != nil {
					return nil, fmt.Errorf("http request failed: %v", err)
				}
			}
			defer resp.Body.Close()

			// 文件名：优先 Content-Disposition，其次 URL 路径
			name := path.Base(u.Path)
			if cd := resp.Header.Get("Content-Disposition"); cd != "" {
				if _, params, err := mime.ParseMediaType(cd); err == nil {
					if fn, ok := params["filename"]; ok && fn != "" {
						name = fn
					}
				}
			}
			ext := strings.ToLower(filepath.Ext(name))

			mimeType := resp.Header.Get("Content-Type")
			if mimeType == "" {
				mimeType = mime.TypeByExtension(ext)
			}

			size := resp.ContentLength
			fileType := classifyFileType(mimeType, ext)

			fi := &FileInfo{
				Name:     name,
				Path:     source,
				Ext:      ext,
				Size:     size,
				MimeType: mimeType,
				Type:     fileType,
			}
			// 可能缺少长度信息
			if size < 0 {
				fi.Warnings = append(fi.Warnings, "content length unknown")
			}
			if mimeType == "" {
				fi.Warnings = append(fi.Warnings, "content type unknown")
			}
			return fi, nil

		case "file":
			// 本地文件的 file:// URL
			localPath := filepath.FromSlash(u.Path)
			return getFileInfoLocal(localPath)
		default:
			return nil, fmt.Errorf("不支持的URL协议: %s", u.Scheme)
		}
	}

	// 非 URL，当作本地路径
	return getFileInfoLocal(source)
}

func getFileInfoLocal(p string) (*FileInfo, error) {
	// 检查文件是否存在
	stat, err := os.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", p)
		}
		return nil, fmt.Errorf("cannot access file: %v", err)
	}
	if stat.IsDir() {
		return nil, fmt.Errorf("path is a directory, not a file: %s", p)
	}

	// 扩展名与 MIME
	ext := strings.ToLower(filepath.Ext(p))
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = detectMimeByContent(p)
	}
	fileType := classifyFileType(mimeType, ext)

	return &FileInfo{
		Name:     filepath.Base(p),
		Path:     p,
		Ext:      ext,
		Size:     stat.Size(),
		MimeType: mimeType,
		Type:     fileType,
	}, nil
}

func detectMimeByContent(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return "application/octet-stream"
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, _ := f.Read(buf)
	return http.DetectContentType(buf[:n])
}

func classifyFileType(mimeType, ext string) string {
	if strings.HasPrefix(mimeType, "text/") {
		return "text"
	}
	if strings.HasPrefix(mimeType, "image/") {
		return "image"
	}
	if strings.HasPrefix(mimeType, "application/vnd.openxmlformats-officedocument") ||
		strings.HasSuffix(ext, "xls") {
		return "excel"
	}
	if strings.Contains(mimeType, "pdf") {
		return "pdf"
	}
	if strings.Contains(mimeType, "zip") || strings.Contains(mimeType, "tar") {
		return "archive"
	}
	if mimeType == "application/octet-stream" {
		return "binary"
	}
	return "unknown"
}
