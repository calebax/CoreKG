package fileconvert

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/ygpkg/yg-go/storage"
)

type XLSStrategy struct{}

func (s *XLSStrategy) SourceExt() string {
	return global.FileExtXLS
}

func (s *XLSStrategy) TargetExt() string {
	return global.FileExtXLSX
}

func (s *XLSStrategy) ShouldConvert(ctx *gin.Context, fileInfo *storage.FileInfo) (bool, error) {
	return strings.ToLower(fileInfo.FileExt) == global.FileExtXLS, nil
}
