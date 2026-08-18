package dtokeywords

import (
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type ListSynonymKeywordsRequest struct {
	apiobj.BaseRequest
	Request ListSynonymKeywordsEmbedRequest `json:"request"`
}

type ListSynonymKeywordsEmbedRequest struct {
	apiobj.PageQuery
	Word string `json:"word"`
}

func (opt *ListSynonymKeywordsRequest) Validity(resp *ListSynonymKeywordsResponse) {
}

type GetSynonymKeywordRequest struct {
	apiobj.BaseRequest
	Request GetSynonymKeywordEmbedRequest `json:"request"`
}

type GetSynonymKeywordEmbedRequest struct {
	ID uint `json:"id"`
}

func (opt *GetSynonymKeywordRequest) Validity(resp *GetSynonymKeywordResponse) {
	if opt.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "ID不能为空"
	}
}

type CreateSynonymKeywordRequest struct {
	apiobj.BaseRequest
	Request CreateSynonymKeywordEmbedRequest `json:"request"`
}
type CreateSynonymKeywordEmbedRequest struct {
	Word       string   `json:"word"`
	ChildWords []string `json:"child_words"`
}

func (opt *CreateSynonymKeywordRequest) Validity(resp *CreateSynonymKeywordResponse) {
	if opt.Request.Word == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "主词不能为空"
	}
	if len(opt.Request.ChildWords) == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "子词个数不能为0"
	}
	words := make(map[string]struct{})
	// 检测数组中是否有重复的词
	for _, v := range append(opt.Request.ChildWords, opt.Request.Word) {
		if _, ok := words[v]; ok {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "存在重复的词"
			return
		}
		words[v] = struct{}{}
	}
}

type DeleteSynonymKeywordRequest struct {
	apiobj.BaseRequest
	Request DeleteSynonymKeywordEmbedRequest `json:"request"`
}
type DeleteSynonymKeywordEmbedRequest struct {
	ID uint `json:"id"`
}

func (opt *DeleteSynonymKeywordRequest) Validity(resp *DeleteSynonymKeywordResponse) {
	if opt.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "ID不能为空"
	}
}

type UpdateSynonymKeywordRequest struct {
	apiobj.BaseRequest
	Request UpdateSynonymKeywordEmbedRequest `json:"request"`
}
type UpdateSynonymKeywordEmbedRequest struct {
	ID         uint     `json:"id"`   // 主词ID
	Word       string   `json:"word"` // 主词
	ChildWords []string `json:"child_words"`
}

func (opt *UpdateSynonymKeywordRequest) Validity(resp *UpdateSynonymKeywordResponse) {
	if opt.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "ID不能为空"
	}
	if opt.Request.Word == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "word不能为空"
	}
	if len(opt.Request.ChildWords) == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "子词个数不能为0"
	}
	words := make(map[string]struct{})
	// 检测数组中是否有重复的词
	for _, v := range append(opt.Request.ChildWords, opt.Request.Word) {
		if _, ok := words[v]; ok {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "存在重复的词"
			return
		}
		words[v] = struct{}{}
	}
}

type CreateMajorKeywordRequest struct {
	apiobj.BaseRequest
	Request CreateMajorKeywordEmbedRequest `json:"request"`
}
type CreateMajorKeywordEmbedRequest struct {
	Word        string `json:"word"`
	Description string `json:"description"`
}

func (opt *CreateMajorKeywordRequest) Validity(resp *CreateMajorKeywordResponse) {
	if opt.Request.Word == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "word不能为空"
		return
	}
	if opt.Request.Description == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "description不能为空"
		return
	}
}

type DeleteMajorKeywordRequest struct {
	apiobj.BaseRequest
	Request DeleteMajorKeywordEmbedRequest `json:"request"`
}
type DeleteMajorKeywordEmbedRequest struct {
	ID uint `json:"id"`
}

func (opt *DeleteMajorKeywordRequest) Validity(resp *DeleteMajorKeywordResponse) {
	if opt.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "ID不能为空"
		return
	}
}

type UpdateMajorKeywordRequest struct {
	apiobj.BaseRequest
	Request UpdateMajorKeywordEmbedRequest `json:"request"`
}
type UpdateMajorKeywordEmbedRequest struct {
	ID          uint   `json:"id"`
	Word        string `json:"word"`
	Description string `json:"description"`
}

func (opt *UpdateMajorKeywordRequest) Validity(resp *UpdateMajorKeywordResponse) {
	if opt.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "ID不能为空"
		return
	}
	if opt.Request.Word == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "word不能为空"
		return
	}
	if opt.Request.Description == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "description不能为空"
		return
	}
}

type ListMajorKeywordsRequest struct {
	apiobj.BaseRequest
	Request ListMajorKeywordsEmbedRequest `json:"request"`
}
type ListMajorKeywordsEmbedRequest struct {
	apiobj.PageQuery
	Word string `json:"word"`
}

func (opt *ListMajorKeywordsRequest) Validity(resp *ListMajorKeywordsResponse) {
}

type GetMajorKeywordRequest struct {
	apiobj.BaseRequest
	Request GetMajorKeywordEmbedRequest `json:"request"`
}
type GetMajorKeywordEmbedRequest struct {
	ID uint `json:"id"`
}

func (opt *GetMajorKeywordRequest) Validity(resp *GetMajorKeywordResponse) {
	if opt.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "ID不能为空"
		return
	}
}
