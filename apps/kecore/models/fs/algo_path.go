package fs

import (
	"fmt"
)

// FileContentPath markdown地址(包含prefix)
func FileContentPath(algoFilePath string, fileID uint) string {
	return algoFilePath + fmt.Sprintf(algoFileContentPath, fileID)
}

// FileFileAlgoPath 获取文件的算法目录位置
func FileFileAlgoPath(algoFilePath string, fileID uint) string {
	return algoFilePath + fmt.Sprintf(algoFileDir, fileID)
}

// FileAnalysisPath 文件智能分析地址(包含prefix)
func FileAnalysisPath(algoFilePath string, fileID uint) string {
	return algoFilePath + fmt.Sprintf(algoFileAnalysisPath, fileID)
}

// FileKnowledgeDirPath 文件知识库目录(包含prefix)
func FileKnowledgeDirPath(algoFilePath string, fileID uint) string {
	return algoFilePath + fmt.Sprintf(algoFileKnowledgeDirPath, fileID)
}

// ForestKnowledgeDirPath 知识森林知识库目录(包含prefix)
func ForestKnowledgeDirPath(uin, forestid uint) string {
	return fmt.Sprintf("%s/%d/%d%s", PurposeForestAlgo, uin, forestid, algoForestKnowledgeDirPath)
}

// FileGraphPath 文件知识图谱地址(包含prefix)
func FileGraphPath(algoFilePath string, fileID uint) string {
	return algoFilePath + fmt.Sprintf(algoFileGraphPath, fileID)
}
