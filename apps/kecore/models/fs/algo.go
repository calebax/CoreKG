package fs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/storage"
)

const (
	algoForestDir              = "/0/"             // 知识森林算法目录
	algoFileDir                = "/%d/"            // 知识森林算法文件目录
	algoFileContentPath        = "/%d/content.md"  // 文件markdown地址（相对于prefix的地址）
	algoFileGraphPath          = "/%d/graph.json"  // 文件知识图谱地址
	algoFileAnalysisPath       = "/%d/analysis.md" // 文件智能分析地址
	algoFileKnowledgeDirPath   = "/%d/knowledge/"  // 文件知识库目录地址
	algoForestKnowledgeDirPath = "/0/knowledge/"   // 知识森林知识库目录地址
)

var (
	EmptyContent = bytes.NewBuffer([]byte{})
)

// UploadFileContent 上传markdown文件
func UploadFileContent(finfo *foresttype.KnownowForestFile, r io.Reader) error {
	ctx := context.Background()
	// 获取文件大小
	var size int64
	if sr, ok := r.(*strings.Reader); ok {
		size = int64(sr.Len()) // 使用 strings.Reader 的 Len() 方法
	} else {
		return fmt.Errorf("unsupported reader type: cannot determine file size")
	}
	fi := &storage.FileInfo{
		CompanyID:   finfo.CompanyID,
		Uin:         finfo.Uin,
		Filename:    "content.md",
		Size:        size,
		FileExt:     ".md",
		StoragePath: FileContentPath(finfo.GetAlgoFilePath(), finfo.ID),
	}
	err := Forest.Save(ctx, fi, r)
	if err != nil {
		return err
	}
	fi.PublicURL = Forest.GetPublicURL(fi.StoragePath, false)
	err = CreateFileInfo(fi)
	if err != nil {
		return err
	}

	return err
}

// GetFileContent 根据文件获取文件解析内容
func GetFileContent(finfo *foresttype.KnownowForestFile) ([]byte, error) {
	p := FileContentPath(finfo.GetAlgoFilePath(), finfo.ID)
	file, err := Forest.ReadFile(p)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}

// GetFileGraph 读取算法思维导图文件
func GetFileGraph(finfo *foresttype.KnownowForestFile) ([]byte, error) {
	p := FileGraphPath(finfo.GetAlgoFilePath(), finfo.ID)
	file, err := Forest.ReadFile(p)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}

// GetFileAnalysis 读取算法智能分析文件
func GetFileAnalysis(algoFilePath string, fileID uint) ([]byte, error) {
	p := algoFilePath + fmt.Sprintf(algoFileAnalysisPath, fileID)
	file, err := Forest.ReadFile(p)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}

func GetAnalysisContent(f *foresttype.KnownowForestFile) ([]byte, error) {
	p := FileAnalysisPath(f.GetAlgoFilePath(), f.ID)
	file, err := Forest.ReadFile(p)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}

// GetForestGraphmlContent 读取知识森林Graphml文件内容
func GetForestGraphmlContent(algoFilePath string) ([]byte, error) {
	p := algoFilePath + algoForestKnowledgeDirPath + "graph_chunk_entity_relation.graphml"
	file, err := Forest.ReadFile(p)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}

// GetFileGraphmlContent 读取单文件Graphml文件内容
func GetFileGraphmlContent(algoFilePath string, fileID uint) ([]byte, error) {
	p := algoFilePath + fmt.Sprintf(algoFileKnowledgeDirPath+"graph_chunk_entity_relation.graphml", fileID)
	file, err := Forest.ReadFile(p)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}

// GetFileQuestionContent 读取单文件问答预置问题内容
func GetFileQuestionContent(algoFilePath string, fileID uint) ([]byte, error) {
	p := algoFilePath + fmt.Sprintf(algoFileKnowledgeDirPath+"question.json", fileID)
	file, err := Forest.ReadFile(p)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}

// CreateAnalysis 如果智能分析不存在则创建
func CreateAnalysis(finfo *foresttype.KnownowForestFile, r io.Reader) error {
	ctx := context.Background()
	fi := &storage.FileInfo{
		CompanyID:   finfo.CompanyID,
		Uin:         finfo.Uin,
		Filename:    "analysis.md",
		Size:        int64(EmptyContent.Len()),
		FileExt:     ".md",
		StoragePath: FileAnalysisPath(finfo.GetAlgoFilePath(), finfo.ID),
	}
	err := Forest.Save(ctx, fi, r)
	if err != nil {
		return err
	}
	fi.PublicURL = Forest.GetPublicURL(fi.StoragePath, false)
	err = CreateFileInfo(fi)
	if err != nil {
		return err
	}

	return err
}

// CreateFileKnowledgeDir 创建文件知识库目录
func CreateFileKnowledgeDir(finfo *foresttype.KnownowForestFile) error {
	ctx := context.Background()
	fi := &storage.FileInfo{
		CompanyID:   finfo.CompanyID,
		Uin:         finfo.Uin,
		Filename:    "file_knowledg_dir.md",
		Size:        int64(EmptyContent.Len()),
		FileExt:     ".md",
		StoragePath: FileKnowledgeDirPath(finfo.GetAlgoFilePath(), finfo.ID) + "file_knowledg_dir.md",
	}
	err := Forest.Save(ctx, fi, EmptyContent)
	if err != nil {
		return err
	}
	fi.PublicURL = Forest.GetPublicURL(fi.StoragePath, false)
	err = CreateFileInfo(fi)
	if err != nil {
		return err
	}

	return err
}

// CreateGraph 创建思维导图文件
func CreateGraph(finfo *foresttype.KnownowForestFile, r io.Reader) error {
	ctx := context.Background()
	fi := &storage.FileInfo{
		CompanyID:   finfo.CompanyID,
		Uin:         finfo.Uin,
		Filename:    "graph.json",
		Size:        int64(EmptyContent.Len()),
		FileExt:     ".json",
		StoragePath: FileGraphPath(finfo.GetAlgoFilePath(), finfo.ID),
	}
	err := Forest.Save(ctx, fi, r)
	if err != nil {
		return err
	}
	fi.PublicURL = Forest.GetPublicURL(fi.StoragePath, false)
	err = CreateFileInfo(fi)
	if err != nil {
		return err
	}

	return err
}

// CreateForestKnowledgeDir 创建知识森林目录
func CreateForestKnowledgeDir(compid, uin, forestid uint) error {
	ctx := context.Background()
	fi := &storage.FileInfo{
		CompanyID:   compid,
		Uin:         uin,
		Filename:    "forest_knowledg_dir.md",
		Size:        int64(EmptyContent.Len()),
		FileExt:     ".md",
		StoragePath: ForestKnowledgeDirPath(uin, forestid) + "forest_knowledg_dir.md",
	}
	err := Forest.Save(ctx, fi, EmptyContent)
	if err != nil {
		return err
	}
	fi.PublicURL = Forest.GetPublicURL(fi.StoragePath, false)
	err = CreateFileInfo(fi)
	if err != nil {
		return err
	}

	return err
}
