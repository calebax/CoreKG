package utils

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/xuri/excelize/v2"
)

// helper: create a small excel file locally and return its path
func createLocalSampleExcel(t *testing.T) string {
	t.Helper()
	f := excelize.NewFile()
	// Default sheet is "Sheet1"
	// Fill 3 rows x 2 cols
	_ = f.SetCellValue("Sheet1", "A1", "r1c1")
	_ = f.SetCellValue("Sheet1", "B1", "r1c2")
	_ = f.SetCellValue("Sheet1", "A2", "r2c1")
	_ = f.SetCellValue("Sheet1", "B2", "r2c2")
	_ = f.SetCellValue("Sheet1", "A3", "r3c1")
	_ = f.SetCellValue("Sheet1", "B3", "r3c2")

	p := filepath.Join(t.TempDir(), "sample.xlsx")
	if err := f.SaveAs(p); err != nil {
		t.Fatalf("SaveAs failed: %v", err)
	}
	_ = f.Close()
	return p
}

// helper: create excel bytes for http server
func createSampleExcelBytes(t *testing.T) []byte {
    t.Helper()
    f := excelize.NewFile()
    _ = f.SetCellValue("Sheet1", "A1", "r1c1")
    _ = f.SetCellValue("Sheet1", "B1", "r1c2")
    _ = f.SetCellValue("Sheet1", "A2", "r2c1")
    _ = f.SetCellValue("Sheet1", "B2", "r2c2")
    _ = f.SetCellValue("Sheet1", "A3", "r3c1")
    _ = f.SetCellValue("Sheet1", "B3", "r3c2")

    var buf bytes.Buffer
    if _, err := f.WriteTo(&buf); err != nil {
        t.Fatalf("Write failed: %v", err)
    }
    _ = f.Close()
    return buf.Bytes()
}

func TestOpenFromPathOrURL_LocalPath(t *testing.T) {
    p := createLocalSampleExcel(t)
    rc, name, ft, err := OpenFromPathOrURL(p)
    if err != nil {
        t.Fatalf("OpenFromPathOrURL local failed: %v", err)
    }
    defer func() { _ = rc.Close() }()

    if name != filepath.Base(p) {
        t.Fatalf("unexpected file name: got %s want %s", name, filepath.Base(p))
    }
    if ft != FileTypeXlsx {
        t.Fatalf("unexpected file type: %s", ft)
    }
    f, err := excelize.OpenReader(rc)
    if err != nil {
        t.Fatalf("OpenReader failed: %v", err)
    }
    defer func() { _ = f.Close() }()
    sheets := f.GetSheetList()
    if len(sheets) != 1 || sheets[0] != "Sheet1" {
        t.Fatalf("unexpected sheet list: %v", sheets)
    }
}

func TestOpenFromPathOrURL_FileURL(t *testing.T) {
    p := createLocalSampleExcel(t)
    fileURL := "file://" + filepath.ToSlash(p)
    rc, name, ft, err := OpenFromPathOrURL(fileURL)
    if err != nil {
        t.Fatalf("OpenFromPathOrURL file URL failed: %v", err)
    }
    defer func() { _ = rc.Close() }()

    if name != filepath.Base(p) {
        t.Fatalf("unexpected file name: got %s want %s", name, filepath.Base(p))
    }
    if ft != FileTypeXlsx {
        t.Fatalf("unexpected file type: %s", ft)
    }
    f, err := excelize.OpenReader(rc)
    if err != nil {
        t.Fatalf("OpenReader failed: %v", err)
    }
    defer func() { _ = f.Close() }()
    if len(f.GetSheetList()) == 0 {
        t.Fatalf("no sheets found")
    }
}

func TestOpenFromPathOrURL_HTTPURL(t *testing.T) {
    data := createSampleExcelBytes(t)
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Disposition", "attachment; filename=remote.xlsx")
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write(data)
    }))
    defer ts.Close()

    rc, name, ft, err := OpenFromPathOrURL(ts.URL)
    if err != nil {
        t.Fatalf("OpenFromPathOrURL http URL failed: %v", err)
    }
    defer func() { _ = rc.Close() }()

    if name != "remote.xlsx" {
        t.Fatalf("unexpected file name: got %s want remote.xlsx", name)
    }
    if ft != FileTypeXlsx {
        t.Fatalf("unexpected file type: %s", ft)
    }
    f, err := excelize.OpenReader(rc)
    if err != nil {
        t.Fatalf("OpenReader failed: %v", err)
    }
    defer func() { _ = f.Close() }()
    if len(f.GetSheetList()) == 0 {
        t.Fatalf("no sheets found")
    }
}

func TestViewSheetRowData_Local(t *testing.T) {
    p := createLocalSampleExcel(t)
    rc, _, _, err := OpenFromPathOrURL(p)
    if err != nil {
        t.Fatalf("OpenFromPathOrURL failed: %v", err)
    }
    defer func() { _ = rc.Close() }()

    rows, err := ViewSheetRowDataFromReader(rc, 0, 0, 1)
    if err != nil {
        t.Fatalf("ViewSheetRowData failed: %v", err)
    }
    if len(rows) != 2 {
        t.Fatalf("unexpected rows len: %d", len(rows))
    }
    if rows[0][0] != "r1c1" || rows[0][1] != "r1c2" || rows[1][0] != "r2c1" || rows[1][1] != "r2c2" {
        t.Fatalf("unexpected rows: %v", rows)
    }
}

func TestViewSheetRowData_HTTPURL(t *testing.T) {
    data := createSampleExcelBytes(t)
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write(data)
    }))
    defer ts.Close()

    rc, _, _, err := OpenFromPathOrURL(ts.URL)
    if err != nil {
        t.Fatalf("OpenFromPathOrURL failed: %v", err)
    }
    defer func() { _ = rc.Close() }()

    rows, err := ViewSheetRowDataFromReader(rc, 0, 1, 2)
    if err != nil {
        t.Fatalf("ViewSheetRowData http failed: %v", err)
    }
    if len(rows) != 2 {
        t.Fatalf("unexpected rows len: %d", len(rows))
    }
    if rows[0][0] != "r2c1" || rows[0][1] != "r2c2" || rows[1][0] != "r3c1" || rows[1][1] != "r3c2" {
        t.Fatalf("unexpected rows: %v", rows)
    }
}

func TestViewSheetRowData_InvalidRanges(t *testing.T) {
    p := createLocalSampleExcel(t)
    rc1, _, _, err := OpenFromPathOrURL(p)
    if err != nil {
        t.Fatalf("OpenFromPathOrURL failed: %v", err)
    }
    defer func() { _ = rc1.Close() }()

    if _, err := ViewSheetRowDataFromReader(rc1, 0, -1, 1); err == nil {
        t.Fatalf("expected error for negative start index")
    }
    // start > end 应返回空结果，不报错（新的 reader）
    rc2, _, _, err := OpenFromPathOrURL(p)
    if err != nil {
        t.Fatalf("OpenFromPathOrURL failed: %v", err)
    }
    defer func() { _ = rc2.Close() }()
    rows, err := ViewSheetRowDataFromReader(rc2, 0, 2, 1)
    if err != nil {
        t.Fatalf("unexpected error when start > end: %v", err)
    }
    if len(rows) != 0 {
        t.Fatalf("expected empty rows when start > end, got %d", len(rows))
    }

    rc3, _, _, err := OpenFromPathOrURL(p)
    if err != nil {
        t.Fatalf("OpenFromPathOrURL failed: %v", err)
    }
    defer func() { _ = rc3.Close() }()
    rows, err = ViewSheetRowDataFromReader(rc3, 0, 10, 20)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if len(rows) != 0 {
        t.Fatalf("expected empty rows when start beyond total, got %d", len(rows))
    }
}

func TestGetXlsFileInfo_Local(t *testing.T) {
	p := createLocalSampleExcel(t)
	info, err := GetXlsFileInfo(p)
	if err != nil {
		t.Fatalf("GetXlsFileInfo failed: %v", err)
	}
	if info.FileName != filepath.Base(p) {
		t.Fatalf("unexpected FileName: %s", info.FileName)
	}
	if info.SheetCount != 1 || len(info.Sheets) != 1 {
		t.Fatalf("unexpected sheet count: %d, sheets len: %d", info.SheetCount, len(info.Sheets))
	}
	if info.Sheets[0].RowCount != 3 || info.Sheets[0].ColCount < 2 {
		t.Fatalf("unexpected sheet dims: rows=%d cols=%d", info.Sheets[0].RowCount, info.Sheets[0].ColCount)
	}
}

func TestGetXlsFileInfo_HTTPURL(t *testing.T) {
    data := createSampleExcelBytes(t)
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Disposition", "attachment; filename=remote.xlsx")
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write(data)
    }))
    defer ts.Close()

    info, err := GetXlsFileInfo(ts.URL)
    if err != nil {
        t.Fatalf("GetXlsFileInfo http failed: %v", err)
    }
    if info.FileName != "remote.xlsx" {
        t.Fatalf("unexpected FileName: %s", info.FileName)
    }
    if info.SheetCount != 1 {
        t.Fatalf("unexpected sheet count: %d", info.SheetCount)
    }
}

func TestOpenFromPathOrURL_UnsupportedScheme(t *testing.T) {
    if _, _, _, err := OpenFromPathOrURL("ftp://example.com/file.xlsx"); err == nil {
        t.Fatalf("expected error for unsupported scheme")
    }
}

func TestOpenFromPathOrURL_EmptyInput(t *testing.T) {
    if _, _, _, err := OpenFromPathOrURL(""); err == nil {
        t.Fatalf("expected error for empty input")
    }
}

func TestOpenFromPathOrURL_HTTPURL_NoContentDisposition(t *testing.T) {
    data := createSampleExcelBytes(t)
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 不设置 Content-Disposition，文件名应取自 URL 路径
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write(data)
    }))
    defer ts.Close()

    urlWithName := ts.URL + "/folder/test.xlsx"
    rc, name, ft, err := OpenFromPathOrURL(urlWithName)
    if err != nil {
        t.Fatalf("OpenFromPathOrURL http no CD failed: %v", err)
    }
    defer func() { _ = rc.Close() }()

    if name != "test.xlsx" {
        t.Fatalf("unexpected file name: got %s want test.xlsx", name)
    }
    if ft != FileTypeXlsx {
        t.Fatalf("unexpected file type: %s", ft)
    }
}
