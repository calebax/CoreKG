package utils

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

// 测试文件路径（统一管理）
const testExcelFilePath = "./1-yGrnoy2po.xlsx"

// getTestFilePath 返回测试文件路径，如果文件不存在则跳过测试
func getTestFilePath(t *testing.T) string {
	if _, err := os.Stat(testExcelFilePath); os.IsNotExist(err) {
		t.Skipf("测试文件不存在，跳过测试: %s", testExcelFilePath)
	}
	return testExcelFilePath
}

// openTestFile 打开测试文件，如果文件不存在则跳过测试
func openTestFile(t *testing.T) *os.File {
	filePath := getTestFilePath(t)
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("打开文件失败: %v", err)
	}
	return file
}

func TestGetXlsFileInfo(t *testing.T) {
	filePath := getTestFilePath(t)

	startTime := time.Now()
	fileInfo, err := GetXlsFileInfo(filePath)
	elapsed := time.Since(startTime)

	if err != nil {
		t.Fatalf("GetXlsFileInfo failed: %v", err)
	}

	// 将结果转为 JSON 格式输出，便于阅读
	jsonData, err := json.MarshalIndent(fileInfo, "", "  ")
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}

	t.Logf("=== Excel 文件信息 ===\n%s", string(jsonData))
	t.Logf("=== 执行耗时: %v ===", elapsed)
}

func TestViewSheetRowDataFromReader(t *testing.T) {
	file := openTestFile(t)
	defer file.Close()

	startTime := time.Now()
	// sheetIndex=0 表示第一个工作表
	// rowStartIdx=0 表示从首行开始
	// rowEndIdx=-1 表示读取到最后一行
	rows, err := ViewSheetRowDataFromReader(file, 0, 0, -1)
	elapsed := time.Since(startTime)

	if err != nil {
		t.Fatalf("ViewSheetRowDataFromReader failed: %v", err)
	}

	t.Logf("=== 读取行数: %d ===", len(rows))
	if len(rows) > 0 {
		t.Logf("=== 首行列数: %d ===", len(rows[0]))
		if len(rows) > 1 {
			t.Logf("=== 第二行数据预览: %v ===", rows[1])
		}
	}
	t.Logf("=== 执行耗时: %v ===", elapsed)
}

// TestViewSheetRowDataFromReader_PartialRows 测试读取部分行的性能
func TestViewSheetRowDataFromReader_PartialRows(t *testing.T) {
	file := openTestFile(t)
	defer file.Close()

	startTime := time.Now()
	// 只读取前 1000 行
	rows, err := ViewSheetRowDataFromReader(file, 0, 0, 999)
	elapsed := time.Since(startTime)

	if err != nil {
		t.Fatalf("ViewSheetRowDataFromReader failed: %v", err)
	}

	t.Logf("=== 读取行数: %d ===", len(rows))
	t.Logf("=== 执行耗时: %v (只读取前1000行) ===", elapsed)
}
