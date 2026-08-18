package filecontent

import (
	"context"
	"fmt"
	"io"
	"strings"

	fileutils "github.com/insmtx/corekg/pkgs/einotools/utils"
)

// maxSpreadsheetRows limits spreadsheet content to 80 rows to control prompt size.
const maxSpreadsheetRows = 80

// Read reads supported file content up to the requested size and reports whether the result was truncated.
func Read(ctx context.Context, source string, maxBytes int64) (content string, filename string, truncated bool, err error) {
	if maxBytes <= 0 {
		return "", "", false, fmt.Errorf("max bytes must be greater than zero")
	}

	reader, name, fileType, err := fileutils.OpenFile(ctx, source)
	if err != nil {
		return "", "", false, fmt.Errorf("open file failed: %w", err)
	}
	defer reader.Close()

	switch fileType {
	case fileutils.FileTypeText, fileutils.FileTypeCsv:
		data, readErr := io.ReadAll(io.LimitReader(reader, maxBytes+1))
		if readErr != nil {
			return "", "", false, fmt.Errorf("read file failed: %w", readErr)
		}
		if int64(len(data)) > maxBytes {
			truncated = true
			data = data[:maxBytes]
		}
		return string(data), name, truncated, nil
	case fileutils.FileTypeXlsx, fileutils.FileTypeXls:
		rows, readErr := fileutils.ViewSheetRowDataFromReader(reader, 0, 0, maxSpreadsheetRows)
		if readErr != nil {
			return "", "", false, fmt.Errorf("read excel content failed: %w", readErr)
		}
		if len(rows) > maxSpreadsheetRows {
			truncated = true
			rows = rows[:maxSpreadsheetRows]
		}
		var contentBuilder strings.Builder
		for index, row := range rows {
			if index > 0 {
				contentBuilder.WriteString("\n")
			}
			line, flattenErr := fileutils.FlattenRowToCSVLine(row)
			if flattenErr != nil {
				return "", "", false, fmt.Errorf("flatten row failed: %w", flattenErr)
			}
			contentBuilder.WriteString(line)
		}
		return contentBuilder.String(), name, truncated, nil
	default:
		return "", "", false, fmt.Errorf("unsupported file type: %s", fileType)
	}
}
