package keqa

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
)

type QA interface {
	Chat(ctx *gin.Context, qs *foresttype.KnownowForestQA, session *foresttype.KnownowQASession) (*foresttype.KnownowForestQA, error)
}

func NewQA(sessionBaseType foresttype.KnownowQASessionBaseType) QA {
	switch sessionBaseType {
	case foresttype.KnownowQASessionBaseTypeExcel:
		return &excelQA{}
	default:
		return nil
	}
}
