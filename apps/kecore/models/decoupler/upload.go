package decoupler

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/fs"
	"github.com/insmtx/corekg/apps/ketask/models/ragtask"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/xuri/excelize/v2"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
	"github.com/ygpkg/yg-go/storage"
	"gorm.io/gorm"
)

const (
	MaxFileNameSize = 200
	TimeStepFmt     = "20060102_150405"
)

// UploadFile 上传文件
func UploadFile(ctx *gin.Context, forestID int) (*foresttype.KnownowForestFile, error) {

	var err error
	pidstr := ctx.Request.FormValue("parent_id")
	var pid int
	if pidstr != "" {
		pid, err = strconv.Atoi(pidstr)
		if err != nil {
			return nil, err
		}
	}
	var parent *foresttype.KnownowForestFile
	if pid > 0 {
		parent, err = forest.GetForestFileByID(uint(pid))
		if err != nil {
			return nil, err
		}
	}
	// 获取文件
	f, fh, err := ctx.Request.FormFile("file")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	//get ext and cut over_size filename
	logs.InfoContextf(ctx, "[UploadFile] filename: %s", fh.Filename)
	ext := filepath.Ext(fh.Filename)
	// fNameTrimExt := fh.Filename[0 : len(fh.Filename)-len(ext)]
	// if len(fNameTrimExt) > MaxFileNameSize {
	// 	fh.Filename = fNameTrimExt[:MaxFileNameSize] + ext
	// }
	runes := []rune(fh.Filename)
	if len(runes) > MaxFileNameSize {
		return nil, fmt.Errorf("文件名过长")
	}

	isExist, err := forest.IsExistForestFile(uint(forestID), uint(pid), fh.Filename)
	if err != nil {
		return nil, err
	}
	if isExist {
		//try to use a timestamp name suffix
		logs.WarnContextf(ctx, "该文件已经存在,创建同名文件_timestep")
		fh.Filename = TruncateName(fh.Filename[:len(fh.Filename)-len(ext)]) + ext
	}
	// 创建file 对象
	finfo := &foresttype.KnownowForestFile{
		CompanyID: runtime.CompanyID(ctx),
		Uin:       runtime.Uin(ctx),
		ForestID:  uint(forestID),
		IsDir:     -1,
		Name:      fh.Filename,
		Ext:       strings.ToLower(filepath.Ext(fh.Filename)),
		Size:      fh.Size,
		FileConfig: foresttype.FileConfig{
			SplitConfig: &ragtask.SplitConfig{},
		},
	}
	if !forest.ParsAble(finfo) {
		finfo.ParseStatus = foresttype.TaskStatusUnsupported
		finfo.KnowledgeStatus = foresttype.TaskStatusUnsupported
		finfo.AnalysisStatus = foresttype.TaskStatusUnsupported
		finfo.GraphStatus = foresttype.TaskStatusUnsupported
	}
	if !forest.PreViewAble(finfo) {
		finfo.PreViewAble = foresttype.PreViewAbleStatusUnsupported
	}

	if parent == nil {
		// 知识森林根目录下创建
		finfo.Depth = 1
	} else {
		finfo.ParentID = parent.ID
		finfo.ParentIDs = fmt.Sprintf("%s%d/", parent.ParentIDs, parent.ID)
		finfo.Depth = parent.Depth + 1
	}

	err = dbutil.Knownow().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(finfo).Error; err != nil {
			return err
		}
		fi := &storage.FileInfo{
			CompanyID: runtime.CompanyID(ctx),
			Uin:       runtime.Uin(ctx),
			Filename:  fh.Filename,
			Size:      fh.Size,
			FileExt:   finfo.Ext,
			// StoragePath: finfo.GetForestFilePath(),
			StoragePath: storage.GenerateFileStoragePath(foresttype.PurposeForestFile, runtime.Uin(ctx), finfo.Ext),
		}
		err = fs.Forest.Save(ctx, fi, f)
		if err != nil {
			return err
		}
		fi.PublicURL = fs.Forest.GetPublicURL(fi.StoragePath, false)
		err = fs.CreateFileInfo(fi)
		if err != nil {
			return err
		}
		ok, previewFileEntity := buildPreviewFileEntity(ctx, fi)
		if ok {
			if err := fs.CreateFileInfo(previewFileEntity); err != nil {
				logs.ErrorContextf(ctx, "forest: UploadFileCallBack CreatePreviewFileInfo failed %v", err)
				return err
			}
		}
		finfo.PriviewFileID = previewFileEntity.ID
		finfo.PriviewExt = previewFileEntity.FileExt
		// finfo.PriviewFileID = fi.ID
		// finfo.PriviewExt = fi.FileExt
		finfo.CoreFileID = fi.ID
		// 如果是ppt或者word，转pdf
		{
			// switch finfo.Ext {
			// case ".ppt", ".pptx", ".doc", ".docx":
			// 	pdf, err := FileToPDF(f, finfo.Name)
			// 	if err != nil {
			// 		return err
			// 	}
			// 	defer pdf.Close()
			// 	finfo.PriviewExt = ".pdf"
			// 	fi := &storage.FileInfo{
			// 		CompanyID: runtime.CompanyID(ctx),
			// 		Uin:       runtime.Uin(ctx),
			// 		Filename:  fh.Filename + finfo.PriviewExt,
			// 		Size:      fh.Size,
			// 		FileExt:   finfo.PriviewExt,
			// 		// StoragePath: finfo.GetForestPriviewFilePath(),
			// 		StoragePath: storage.GenerateFileStoragePath(foresttype.PurposeForestFile, runtime.Uin(ctx), finfo.PriviewExt),
			// 	}
			// 	err = fs.Forest.Save(ctx, fi, pdf)
			// 	if err != nil {
			// 		return err
			// 	}
			// 	fi.PublicURL = fs.Forest.GetPublicURL(fi.StoragePath, false)
			// 	err = fs.CreateFileInfo(fi)
			// 	if err != nil {
			// 		return err
			// 	}
			// 	finfo.PriviewFileID = fi.ID
			// case ".csv":
			// 	excel, err := CSVToExcel(f)
			// 	if err != nil {
			// 		return err
			// 	}
			// 	finfo.PriviewExt = ".xlsx"
			// 	fi := &storage.FileInfo{
			// 		CompanyID: runtime.CompanyID(ctx),
			// 		Uin:       runtime.Uin(ctx),
			// 		Filename:  fh.Filename + finfo.PriviewExt,
			// 		Size:      fh.Size,
			// 		FileExt:   finfo.PriviewExt,
			// 		// StoragePath: finfo.GetForestPriviewFilePath(),
			// 		StoragePath: storage.GenerateFileStoragePath(foresttype.PurposeForestFile, runtime.Uin(ctx), finfo.PriviewExt),
			// 	}
			// 	err = fs.Forest.Save(ctx, fi, excel)
			// 	if err != nil {
			// 		return err
			// 	}
			// 	fi.PublicURL = fs.Forest.GetPublicURL(fi.StoragePath, false)
			// 	err = fs.CreateFileInfo(fi)
			// 	if err != nil {
			// 		return err
			// 	}
			// 	finfo.PriviewFileID = fi.ID
			// }
			// if err := tx.Save(finfo).Error; err != nil {
			// 	return err
			// }
		}
		if err := tx.Save(finfo).Error; err != nil {
			return err
		}
		// 生产任务
		// if err := task.GenerateParseTask(*finfo); err != nil {
		// 	return err
		// }
		return nil
	})
	if err != nil {
		return nil, err
	}

	return finfo, nil
}

type ConvertPDFConfig struct {
	Default ConvertPDFServiceConfig `yaml:"default"`
	OFD     ConvertPDFServiceConfig `yaml:"ofd"`
}

type ConvertPDFServiceConfig struct {
	URL            string `yaml:"url"`
	APIKey         string `yaml:"api_key"`
	FileField      string `yaml:"file_field"`
	ResponseIsFile bool   `yaml:"response_is_file"`
	FileURLField   string `yaml:"file_url_field"`
}

func buildPreviewFileEntity(ctx *gin.Context, coreFile *storage.FileInfo) (bool, *storage.FileInfo) {
	var priviewExt string
	switch coreFile.FileExt {
	case global.FileExtPPT, global.FileExtPPTX, global.FileExtDOC, global.FileExtDOCX, global.FileExtOFD:
		priviewExt = global.FileExtPDF
	case global.FileExtCSV:
		priviewExt = global.FileExtXLSX
	default:
		// 不需要生成预览，直接返回原文件
		return false, coreFile
	}
	storagePath := storage.GenerateFileStoragePath(foresttype.PurposeForestFile, coreFile.Uin, priviewExt)
	fileEntity := &storage.FileInfo{
		CompanyID:   runtime.CompanyID(ctx),
		Uin:         runtime.Uin(ctx),
		Filename:    coreFile.Filename + priviewExt,
		Size:        coreFile.Size,
		FileExt:     priviewExt,
		StoragePath: storagePath,
		PublicURL:   fs.Forest.GetPublicURL(storagePath, false),
		Status:      storage.FileStatusUploading,
	}

	return true, fileEntity
}

// FileToPDF 将上传的文件发送至转换服务，返回 PDF 内容或错误
func FileToPDF(ctx context.Context, f io.Reader, filename string) (io.ReadCloser, error) {
	cfg := &ConvertPDFConfig{}
	err := settings.GetYaml("knowledge", "convert_to_pdf", cfg)
	if err != nil {
		logs.ErrorContextf(ctx, "read convert pdf config error:%v", err)
		return nil, err
	}

	serviceCfg, err := cfg.getServiceConfig(filename)
	if err != nil {
		return nil, err
	}

	// 1. 创建缓冲区和 multipart writer
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 2. 创建表单文件字段
	fileField := serviceCfg.FileField
	if fileField == "" {
		fileField = "files"
	}
	part, err := writer.CreateFormFile(fileField, filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %v", err)
	}

	// 3. 把上传的文件内容拷贝到 multipart body 中
	_, err = io.Copy(part, f)
	if err != nil {
		return nil, fmt.Errorf("writing file content failed: %v", err)
	}

	// 4. 关闭 writer，结束 multipart 数据
	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %v", err)
	}
	// 5. 创建并发送 POST 请求
	req, err := http.NewRequest("POST", serviceCfg.URL, body)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if serviceCfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+serviceCfg.APIKey)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request failed: %v", err)
	}

	// 6. 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		responseBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("the server returned an error status code: %d, in response to: %s", resp.StatusCode, string(responseBody))
	}

	if serviceCfg.ResponseIsFile {
		return resp.Body, nil
	}

	fileURL, err := extractConvertPDFFileURL(resp.Body, serviceCfg.FileURLField)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}

	downloadReq, err := http.NewRequest("GET", fileURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create download request failed: %v", err)
	}

	downloadResp, err := client.Do(downloadReq)
	if err != nil {
		return nil, fmt.Errorf("downloading converted pdf failed: %v", err)
	}
	if downloadResp.StatusCode != http.StatusOK {
		defer downloadResp.Body.Close()
		responseBody, _ := io.ReadAll(downloadResp.Body)
		return nil, fmt.Errorf("download converted pdf failed, status code: %d, response: %s", downloadResp.StatusCode, string(responseBody))
	}

	return downloadResp.Body, nil
}

func (c *ConvertPDFConfig) getServiceConfig(filename string) (ConvertPDFServiceConfig, error) {
	if strings.EqualFold(filepath.Ext(filename), global.FileExtOFD) && c.OFD.URL != "" {
		return c.OFD, nil
	}
	if c.Default.URL != "" {
		return c.Default, nil
	}
	if c.OFD.URL != "" {
		return c.OFD, nil
	}
	return ConvertPDFServiceConfig{}, fmt.Errorf("convert pdf config url is empty")
}

func extractConvertPDFFileURL(body io.Reader, fileURLField string) (string, error) {
	if fileURLField == "" {
		return "", fmt.Errorf("convert pdf config file_url_field is empty")
	}

	var response map[string]any
	if err := json.NewDecoder(body).Decode(&response); err != nil {
		return "", fmt.Errorf("decode convert pdf response failed: %v", err)
	}

	current := any(response)
	for _, key := range strings.Split(fileURLField, ".") {
		data, ok := current.(map[string]any)
		if !ok {
			return "", fmt.Errorf("convert pdf response field %q not found", fileURLField)
		}
		current, ok = data[key]
		if !ok {
			return "", fmt.Errorf("convert pdf response field %q not found", fileURLField)
		}
	}

	fileURL, ok := current.(string)
	if !ok || fileURL == "" {
		return "", fmt.Errorf("convert pdf response field %q is empty", fileURLField)
	}

	return fileURL, nil
}

// CSVToExcel 使用 excelize 将 CSV 文件转换为 Excel 文件
func CSVToExcel(ctx context.Context, f io.Reader) (io.Reader, error) {

	csvReader := csv.NewReader(f)
	csvReader.FieldsPerRecord = -1
	csvReader.LazyQuotes = true // 关键设置：允许宽松的引号处理
	csvReader.TrimLeadingSpace = true
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV data: %v", err)
	}

	// 创建一个新的 Excel 文件
	ex := excelize.NewFile()
	defer func() {
		if err := ex.Close(); err != nil {
			logs.ErrorContextf(ctx, "failed to close Excel file: %v", err)
		}
	}()

	// 获取默认的工作表名称（如 Sheet1）
	sheetName := "Sheet1"
	index, _ := ex.NewSheet(sheetName)

	// 写入 CSV 数据到 Excel 表格中
	for rowIndex, row := range records {
		for colIndex, cell := range row {
			cellName, err := excelize.CoordinatesToCellName(colIndex+1, rowIndex+1)
			if err != nil {
				return nil, fmt.Errorf("failed to generate cell location: %v", err)
			}
			ex.SetCellValue(sheetName, cellName, cell)
		}
	}

	// 设置默认工作表（可选）
	ex.SetActiveSheet(index)

	exbuffer, err := ex.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("writing Excel data failed: %v", err)
	}

	return exbuffer, nil
}

// TruncateName will truncate name, whose length will append with a preserved timeStep len
func TruncateName(fileName string) string {
	cutLen := MaxFileNameSize - len(TimeStepFmt) + 1
	if len(fileName) >= cutLen {
		fileName = fileName[:cutLen]
	}
	fileName = fmt.Sprintf("%v_%v%v", filepath.Base(fileName), time.Now().Format(TimeStepFmt), filepath.Ext(fileName))
	return fileName
}
