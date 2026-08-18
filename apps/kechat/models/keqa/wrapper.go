package keqa

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
)

type ForestWrapper struct {
	ctx      *gin.Context
	question *chattype.ChatQuestion
	// files     []*foresttype.KnownowForestFile
	refList   chattype.QueryReferenceList
	history   string
	wrapper   *SearchReferenceWrapper
	searchStr string
}

func NewForestWrapper(ctx *gin.Context, question *chattype.ChatQuestion, wrapper *SearchReferenceWrapper, history string) *ForestWrapper {
	return &ForestWrapper{
		ctx:      ctx,
		wrapper:  wrapper,
		question: question,
		history:  history,
	}
}
