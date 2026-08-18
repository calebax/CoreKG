package fileconvert

import (
	"github.com/gin-gonic/gin"
	"github.com/ygpkg/yg-go/storage"
)

// FileConvertStrategy 文件转换策略接口
type FileConvertStrategy interface {
	SourceExt() string
	TargetExt() string
	ShouldConvert(ctx *gin.Context, fileInfo *storage.FileInfo) (bool, error)
}

// convertStrategies 转换策略注册中心
var convertStrategies = []FileConvertStrategy{
	// XLS 转换为 xlsx
	&XLSStrategy{},
	// Strict OOXML xlsx 转换为 Transitional xlsx
	&StrictOOXMLStrategy{},
}

// GetConvertStrategies 获取所有转换策略
func GetConvertStrategies() []FileConvertStrategy {
	return convertStrategies
}
