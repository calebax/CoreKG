package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/user"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/ketask/models/ragtypes"
	"github.com/insmtx/corekg/apps/kesearch/internal/dto/dtochunk"
	"github.com/insmtx/corekg/apps/kesearch/services/svcchunk"
	"github.com/ygpkg/yg-go/apis/errcode"
)

// GetChunkBySequence 根据序列获取分片
// @Tags chunk 管理
// @Summary 根据序列获取分片
// @Description 根据序列获取分片
// @Router /kesearch.GetChunkBySequence [post]
// @Param user body dtochunk.GetChunkBySequenceRequest true "入参"
// @Success 200 {object} dtochunk.GetChunkBySequenceResponse "返回值"
func GetChunkBySequence(ctx *gin.Context, req *dtochunk.GetChunkBySequenceRequest, resp *dtochunk.GetChunkBySequenceResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}
	chunkInfo, err := svcchunk.GetChunkBySequence(ctx, req)
	if err != nil {
		resp.Code = err.Code
		resp.Message = err.Message
		return
	}
	resp.Response.Chunk = chunkInfo
}

// GetChunkDetail 根据序列获取分片详情（包含文件信息）
// @Tags chunk 管理
// @Summary 根据序列获取分片详情
// @Description 根据序列获取分片详情（包含文件信息）
// @Router /kesearch.GetChunkDetail [post]
// @Param user body dtochunk.GetChunkBySequenceRequest true "入参"
// @Success 200 {object} dtochunk.GetChunkDetailResponse "返回值"
func GetChunkDetail(ctx *gin.Context, req *dtochunk.GetChunkBySequenceRequest, resp *dtochunk.GetChunkDetailResponse) {
	if req.ValidityDetail(resp); resp.Code != errcode.CodeOK {
		return
	}

	fileInfo, err := forest.GetForestFileByID(req.Request.FileID)
	if err != nil {
		// 获取文件失败
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kesearch_get_file_failed"
		return
	}
	if fileInfo.ID == 0 {
		// 文件不存在
		resp.Code = errcode.ErrCode_NotFound
		resp.Message = "kecore_file_not_found"
		return
	}

	chunkInfo, errInfo := svcchunk.GetChunkBySequence(ctx, req)
	if err != nil {
		resp.Code = errInfo.Code
		resp.Message = errInfo.Message
		return
	}
	if chunkInfo == nil {
		resp.Response.Detail = &chattype.QueryReference{}
		return
	}

	userInfo, _ := user.GetUserByUin(ctx, fileInfo.Uin)

	filechunk := &chattype.QueryReference{
		FileID:         fileInfo.ID,
		Abstract:       "",
		DataSourceType: "DC",
		FileName:       fileInfo.Name,
		Uin:            fileInfo.Uin,
		CreatedAt:      fileInfo.CreatedAt,
	}
	if userInfo != nil {
		filechunk.UserName = userInfo.Name
		filechunk.AvatarURL = userInfo.AvatarURL
	}

	ck := chattype.QueryReferenceChunk{
		ChunkID:  chunkInfo.ID,
		Sequence: chunkInfo.Source.Sequence,
		Content:  chunkInfo.Source.Description,
		ImageURL: chunkInfo.Source.ImageUrl,
		Location: ragtypes.Location(chunkInfo.Source.Location),
		Type:     chunkInfo.Source.Type,
	}

	filechunk.ChunkList = []chattype.QueryReferenceChunk{ck}

	resp.Response.Detail = filechunk
}
