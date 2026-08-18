package file

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	tool "github.com/cloudwego/eino/components/tool"
	toolutils "github.com/cloudwego/eino/components/tool/utils"
	agutils "github.com/insmtx/corekg/pkgs/einotools/utils"
)

// ViewSheetConfig defines configuration for viewing a sheet region.
type FileToolConfig struct {
	ToolName string `json:"tool_name"`
	ToolDesc string `json:"tool_desc"`
}

// ExecRequest defines the input for executing a file operation.
type ExecRequest struct {
	// FilePath string `json:"file_path" jsonschema:"description=Local file system path. Mutually exclusive with 'fileUrl'; exactly one of 'filePath' or 'fileUrl' must be provided."`
	FileUrl string `json:"file_url"  jsonschema:"description=Remote URL. Mutually exclusive with 'filePath'; exactly one of 'filePath' or 'fileUrl' must be provided."`

	ViewRange []int `json:"view_range" jsonschema:"description=Optional parameter when the 'view' command is used on a file. This value must be an array of two integers: [startLine, endLine]. The values are 1-based line numbers. The endLine is inclusive, and setting endLine to -1 means all lines from startLine to the end of the file will be shown. If omitted, the entire file is displayed. Examples: [1, 200], [1000, -1]."`

	SheetIndex int `json:"sheet_index" jsonschema:"description=Zero-based index of the sheet when viewing Excel files (.xlsx or .xls). Ignored for non-Excel files. Default is 0."`
}

type ExecResponse struct {
	Rows string `json:"rows"`
}

type fileTool struct {
	conf *FileToolConfig
}

// NewFileTool creates an InvokableTool that returns a row region of a sheet.
func NewFileTool(ctx context.Context, conf *FileToolConfig) (tool.InvokableTool, error) {
	if conf == nil {
		conf = &FileToolConfig{}
	}
	toolName := conf.ToolName
	toolDesc := conf.ToolDesc
	if toolName == "" {
		toolName = "file_read_tool"
	}
	if toolDesc == "" {
		toolDesc = `在文件结构已明确的前提下，用于读取并展示文件内容。

本工具用于：
- 读取指定的 Sheet、字段或数据范围
- 查看具体数据行，用于验证或后续处理
- 在结构确认后访问文件内容

重要说明（表格文件）：
- 对于 Excel / 表格类文件，日期字段可能以原始数值形式返回
  （例如 44986），表示 Excel 的内部日期序列值。
- 如需将该数值解释为具体日期，应在后续推理或处理阶段显式转换，
  不得在读取阶段擅自假定其格式。

使用规则：
- 仅在已调用 file_inspect_tool 或结构已明确时使用
- 若输出内容过长，将被截断并标记为“<response clipped>”
- 不负责推断或修正文件结构，必须依赖已有的检查结果

查询建议：
- 为减少工具调用次数，优先一次性读取较大范围的数据
- 默认建议单次读取不少于 200 行
- 仅在确有必要时，再缩小到局部范围读取
- 对于需要定位问题的场景，可先大范围读取，再逐步缩小范围
`
	}

	ft := &fileTool{conf: conf}
	tl, err := toolutils.InferTool(toolName, toolDesc, ft.invoke)
	if err != nil {
		return nil, err
	}
	return tl, nil
}

func (t *fileTool) invoke(ctx context.Context, req *ExecRequest) (*ExecResponse, error) {
	if req == nil || (req.FileUrl == "") {
		return nil, fmt.Errorf("file path or file url is required")
	}

	viewRange := req.ViewRange
	if viewRange == nil {
		viewRange = []int{1, -1}
	} else if len(viewRange) != 2 {
		return nil, fmt.Errorf("view range must be a 1-based line number range formatted as [startLine, endLine]")
	}

	source := req.FileUrl
	// if source == "" {
	// 	source = req.FilePath
	// }

	f, _, fileType, err := agutils.OpenFromPathOrURL(source)
	if err != nil {
		return nil, fmt.Errorf("open file failed: %w", err)
	}
	defer f.Close()

	var lines []string
	switch fileType {
	case agutils.FileTypeXlsx:
		sheetRows, err := agutils.ViewSheetRowDataFromReader(f, req.SheetIndex, viewRange[0]-1, viewRange[1]-1)
		if err != nil {
			return nil, fmt.Errorf("view sheet row data failed: %w", err)
		}
		for _, row := range sheetRows {
			lines = append(lines, strings.Join(row, "\t"))
		}
	case agutils.FileTypeText:
		textRows, err := agutils.ReadLinesRange(f, viewRange[0], viewRange[1])
		if err != nil {
			return nil, fmt.Errorf("read lines range failed: %w", err)
		}
		lines = append(lines, textRows...)
	case agutils.FileTypePDF:
		pdfRows, err := agutils.ReadPDFLinesRange(f, viewRange[0], viewRange[1])
		if err != nil {
			return nil, fmt.Errorf("read pdf lines range failed: %w", err)
		}
		lines = append(lines, pdfRows...)
	default:
		return nil, fmt.Errorf("unsupported file type: %s", fileType)
	}

	return &ExecResponse{
		Rows: formatResponse(lines, 64*1024, "\n<response clipped>"),
	}, nil
}

// formatResponse joins lines with newlines and trims to a maximum size, marking clipped output.
func formatResponse(lines []string, maxBytes int, clipMarker string) string {
	joined := strings.Join(lines, "\n")
	if maxBytes <= 0 {
		return joined
	}

	joinedBytes := []byte(joined)
	if len(joinedBytes) <= maxBytes {
		return joined
	}

	markerBytes := []byte(clipMarker)
	available := maxBytes - len(markerBytes)
	if available < 0 {
		available = 0
	}

	truncated := joinedBytes[:available]
	for len(truncated) > 0 && !utf8.Valid(truncated) {
		truncated = truncated[:len(truncated)-1]
	}

	return string(truncated) + clipMarker
}
