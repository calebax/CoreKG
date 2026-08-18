package forestctl

import (
	"errors"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/user"
	"github.com/insmtx/corekg/apps/keapi/internal/dto/dtokeapi"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/services/svcforest"
	"github.com/insmtx/corekg/apps/kecore/services/svcforestfile"
	"github.com/insmtx/corekg/apps/ketask/models/ragtypes"
	"github.com/insmtx/corekg/apps/kesearch/models/chunk"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
)

var errKeAPIFileNotFound = errors.New("keapi file not found")

// GetFileChunksBySequences 根据知识库文件 ID 和 chunk 序号列表查询 chunk 信息
// @Tags 对外文档接口
// @Summary 根据知识库文件 ID 和 chunk 序号列表查询 chunk 信息
// @Description 根据知识库文件 ID 和 chunk 序号列表查询 chunk 信息
// @Router /keapi.GetFileChunksBySequences [post]
// @Param user body dtokeapi.GetFileChunksBySequencesRequest true "入参"
// @Success 200 {object} dtokeapi.GetFileChunksBySequencesResponse "返回值"
func GetFileChunksBySequences(ctx *gin.Context, req *dtokeapi.GetFileChunksBySequencesRequest, resp *dtokeapi.GetFileChunksBySequencesResponse) {
	if !req.ValidGetFileChunksBySequences(&resp.BaseResponse) {
		return
	}

	fileID := req.Request.ForestFileID
	fileInfo, forestInfo, err := getAuthorizedFileAndForest(ctx, fileID)
	if err != nil {
		switch {
		case errors.Is(err, errKeAPIFileNotFound):
			resp.Code = errcode.ErrCode_NotFound
			resp.Message = "keapi_file_not_found"
		case errors.Is(err, svcforestfile.ErrUserIDEmpty):
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kecore_user_id_empty"
		default:
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "keapi_get_file_failed"
		}
		return
	}

	sequences := uniquePositiveInts(req.EffectiveChunkSequences())
	chunks, err := chunk.ListChunksByFileIDAndSequences(ctx, forestInfo.EsIndex(), fileID, sequences)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "keapi_get_chunks_failed"
		return
	}
	abstract, err := chunk.GetFileAbstractByFileID(ctx, forestInfo.EsIndex(), fileID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "keapi_get_file_abstract_failed"
		return
	}

	detail := &chattype.QueryReference{
		FileID:         fileInfo.ID,
		FileName:       fileInfo.Name,
		ForestID:       fileInfo.ForestID,
		CreatedAt:      fileInfo.CreatedAt,
		Uin:            fileInfo.Uin,
		Abstract:       abstract,
		DataSourceType: chattype.DataSourceTypeDC,
		ChunkList:      make(chattype.QueryReferenceChunkList, 0, len(chunks)),
	}
	if userInfo, err := user.GetUserByUin(ctx, fileInfo.Uin); err != nil {
		logs.WarnContextf(ctx, "GetFileChunksBySequences GetUserByUin(%d) error: %v", fileInfo.Uin, err)
	} else if userInfo != nil {
		detail.UserName = userInfo.Name
		detail.AvatarURL = userInfo.AvatarURL
	}

	for _, item := range chunks {
		if item == nil || item.Source == nil {
			continue
		}
		detail.ChunkList = append(detail.ChunkList, chattype.QueryReferenceChunk{
			Type:     item.Source.Type,
			ChunkID:  item.ID,
			Sequence: item.Source.Sequence,
			Content:  item.Source.Description,
			ImageURL: item.Source.ImageUrl,
			Score:    item.Score,
			Location: ragtypes.Location(item.Source.Location),
		})
	}
	sort.Sort(detail.ChunkList)
	resp.Response = detail
}

func getAuthorizedFileAndForest(ctx *gin.Context, fileID uint) (*forest.File, *foresttype.KnownowForest, error) {
	fileOut, err := svcforestfile.ListFile(ctx, &svcforestfile.ListFileRequest{
		Uin:       runtime.Uin(ctx),
		CompanyID: runtime.CompanyID(ctx),
		PageQuery: apiobj.PageQuery{
			ListAll: true,
			Filters: []apiobj.Filter{{
				Field: "id",
				Value: []string{strconv.FormatUint(uint64(fileID), 10)},
			}},
		},
	})
	if err != nil {
		return nil, nil, err
	}
	if fileOut == nil || len(fileOut.Response.Data) == 0 {
		return nil, nil, errKeAPIFileNotFound
	}
	fileInfo := fileOut.Response.Data[0]

	forestOut, err := svcforest.ListForest(ctx, &svcforest.ListForestRequest{
		Uin:       runtime.Uin(ctx),
		CompanyID: runtime.CompanyID(ctx),
		Query: apiobj.PageQuery{
			ListAll: true,
			Filters: []apiobj.Filter{{
				Field: "id",
				Value: []string{strconv.FormatUint(uint64(fileInfo.ForestID), 10)},
			}},
		},
		PresetWhenListing: false,
	})
	if err != nil {
		return nil, nil, err
	}
	if forestOut == nil {
		return nil, nil, errKeAPIFileNotFound
	}
	for _, item := range forestOut.Data {
		if item != nil && item.ID == fileInfo.ForestID {
			return fileInfo, &item.KnownowForest, nil
		}
	}
	return nil, nil, errKeAPIFileNotFound
}

func uniquePositiveInts(values []int) []int {
	seen := make(map[int]struct{}, len(values))
	out := make([]int, 0, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
