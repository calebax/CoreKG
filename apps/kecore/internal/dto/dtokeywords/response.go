package dtokeywords

import (
	"github.com/insmtx/corekg/apps/kecore/models/forestkeywords"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type ListSynonymKeywordsResponse struct {
	apiobj.BaseResponse
	Response ListSynonymKeywordsEmbedResponse `json:"response"`
}

type ListSynonymKeywordsEmbedResponse struct {
	apiobj.QueryResponse
	Data []*forestkeywords.SynonymKeywordDetail `json:"data"`
}

type GetSynonymKeywordResponse struct {
	apiobj.BaseResponse
	Response GetSynonymKeywordEmbedResponse `json:"response"`
}

type GetSynonymKeywordEmbedResponse struct {
	Data forestkeywords.SynonymKeywordDetail `json:"data"`
}

type CreateSynonymKeywordResponse struct {
	apiobj.BaseResponse
	Response CreateSynonymKeywordEmbedResponse `json:"response"`
}
type CreateSynonymKeywordEmbedResponse struct {
}

type AddSynonymKeywordResponse struct {
	apiobj.BaseResponse
	Response AddSynonymKeywordEmbedResponse `json:"response"`
}
type AddSynonymKeywordEmbedResponse struct {
}

type DeleteSynonymKeywordResponse struct {
	apiobj.BaseResponse
	Response DeleteSynonymKeywordEmbedResponse `json:"response"`
}
type DeleteSynonymKeywordEmbedResponse struct {
}

type RemoveSynonymKeywordResponse struct {
	apiobj.BaseResponse
	Response RemoveSynonymKeywordEmbedResponse `json:"response"`
}
type RemoveSynonymKeywordEmbedResponse struct {
}

type UpdateSynonymKeywordResponse struct {
	apiobj.BaseResponse
	Response UpdateSynonymKeywordEmbedResponse `json:"response"`
}
type UpdateSynonymKeywordEmbedResponse struct {
}

type CreateMajorKeywordResponse struct {
	apiobj.BaseResponse
	Response CreateMajorKeywordEmbedResponse `json:"response"`
}
type CreateMajorKeywordEmbedResponse struct {
}

type DeleteMajorKeywordResponse struct {
	apiobj.BaseResponse
	Response DeleteMajorKeywordEmbedResponse `json:"response"`
}
type DeleteMajorKeywordEmbedResponse struct {
}

type UpdateMajorKeywordResponse struct {
	apiobj.BaseResponse
	Response UpdateMajorKeywordEmbedResponse `json:"response"`
}
type UpdateMajorKeywordEmbedResponse struct {
}

type ListMajorKeywordsResponse struct {
	apiobj.BaseResponse
	Response ListMajorKeywordsEmbedResponse `json:"response"`
}
type ListMajorKeywordsEmbedResponse struct {
	apiobj.QueryResponse
	Data []*forestkeywords.MajorKeywordDetail `json:"data"`
}

type GetMajorKeywordResponse struct {
	apiobj.BaseResponse
	Response GetMajorKeywordEmbedResponse `json:"response"`
}
type GetMajorKeywordEmbedResponse struct {
	Data *foresttype.Keywords `json:"data"`
}
