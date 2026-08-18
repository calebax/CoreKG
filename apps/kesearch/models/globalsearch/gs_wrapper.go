package globalsearch

import (
	"context"

	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/keqa"
	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	"github.com/ygpkg/yg-go/logs"
)

// NewForestWrapper 知识库检索
func NewForestWrapper(wrapper *GlobalSearchWrapper) (*GlobalSearchWrapper, error) {
	if wrapper.ImageUrl != "" {
		imageContent, err := keqa.DoImageParseRequest(wrapper.Ctx, wrapper.ImageUrl)
		if err != nil {
			logs.ErrorContextf(wrapper.Ctx, "[image_search] [DoImageParseRequest] error: %v", err)
			return nil, err
		}
		wrapper.imageContent = imageContent

		wrapper.Text = wrapper.Text + "\n" + imageContent
	}
	// 拆词
	keywords, err := essearch.Analyze(wrapper.Ctx, wrapper.Text)
	if err != nil {
		return nil, err
	}
	if wrapper.IsSemantics {
		eb, err := essearch.GetEmbedding(wrapper.Text)
		if err != nil {
			logs.ErrorContextf(wrapper.Ctx, "GetEmbedding error: %v", err)
			return nil, err
		}
		wrapper.embedding = eb
	}
	wrapper.keywords = keywords

	viewForests, err := forest.ViewAbleForests(wrapper.Uin, wrapper.CompanyID)
	if err != nil {
		return nil, err
	}
	wrapper.ViewForestIDs = viewForests
	wrapper.userMap = map[uint]*accounttype.User{}
	return wrapper, nil
}

// NewGlobalWrapper  知识库检索
func NewGlobalWrapper(wrapper *GlobalSearchWrapper) (*GlobalSearchWrapper, error) {
	if wrapper.ImageUrl != "" {
		imageContent, err := keqa.DoImageParseRequest(wrapper.Ctx, wrapper.ImageUrl)
		if err != nil {
			logs.ErrorContextf(wrapper.Ctx, "[image_search] [DoImageParseRequest] error: %v", err)
			return nil, err
		}
		wrapper.imageContent = imageContent

		wrapper.Text = wrapper.Text + "\n" + imageContent
	}
	// 拆词
	keywords, err := essearch.Analyze(wrapper.Ctx, wrapper.Text)
	if err != nil {
		return nil, err
	}
	if wrapper.IsSemantics {
		eb, err := essearch.GetEmbedding(wrapper.Text)
		if err != nil {
			logs.ErrorContextf(wrapper.Ctx, "GetEmbedding error: %v", err)
			return nil, err
		}
		wrapper.embedding = eb
	}
	wrapper.keywords = keywords

	viewForests, err := forest.ViewAbleForests(wrapper.Uin, wrapper.CompanyID)
	if err != nil {
		return nil, err
	}
	wrapper.ViewForestIDs = viewForests
	wrapper.userMap = map[uint]*accounttype.User{}
	return wrapper, nil
}

type GlobalSearchWrapper struct {
	Ctx           context.Context
	Text          string
	Uin           uint
	CompanyID     uint
	ForestIDs     []uint
	ImageUrl      string
	imageContent  string
	SubjectCount  int // 主体数量，文件数量，agent数量
	ItemCount     int // 子项数量，比如文件中的chunk数
	embedding     []float32
	IsSemantics   bool
	EsIndex       string
	keywords      *essearch.AnalyzeResultList
	ViewForestIDs []uint
	userMap       map[uint]*accounttype.User // 用户uin到用户信息的映射
}
