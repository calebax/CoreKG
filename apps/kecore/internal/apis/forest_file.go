package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtoforestfile"
	"github.com/insmtx/corekg/apps/kecore/services/svcforestfile"
	"github.com/ygpkg/yg-go/apis/errcode"
)

// PreUploadFile 获取文件上传预签名
// @Tags 知识森林文件
// @Summary 获取文件上传预签名
// @Description 获取文件上传预签名，用于 PDF 文件上传预览
// @Param user body dtoforestfile.PreUploadFileRequest true "入参,当前只支持pdf预览"
// @Success 200 {object} dtoforestfile.PreUploadFileResponse "成功返回预签名信息"
// @Router /forest/preuploadfile [post]
func PreUploadFile(ctx *gin.Context, req *dtoforestfile.PreUploadFileRequest, resp *dtoforestfile.PreUploadFileResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}

	files, errInfo := svcforestfile.PreUploadFile(ctx, req)
	if errInfo != nil {
		resp.Code = errInfo.Code
		resp.Message = errInfo.Message
		return
	}
	resp.Response.Files = files
}

// UploadFileCallBack 上传完成回调
// @Tags 知识森林文件
// @Summary 预签名上传回调
// @Description 预签名上传回调
// @Param user body dtoforestfile.UploadFileCallBackRequest true "入参,当前只支持pdf预览"
// @Success 200 {object} dtoforestfile.UploadFileCallBackResponse "返回值"
// @Router /forest/UploadFileCallBack [post]
func UploadFileCallBack(ctx *gin.Context, req *dtoforestfile.UploadFileCallBackRequest, resp *dtoforestfile.UploadFileCallBackResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}

	fileID, errInfo := svcforestfile.UploadFileCallBack(ctx, req)
	if errInfo != nil {
		resp.Code = errInfo.Code
		resp.Message = errInfo.Message
		return
	}
	resp.Response.FileID = fileID

}

// AbortUpload 取消上传并释放资源
// @Tags 知识森林文件
// @Summary 取消上传并释放资源
// @Description 取消上传并释放资源
// @Param user body dtoforestfile.AbortUploadRequest true "入参"
// @Success 200 {object} dtoforestfile.AbortUploadResponse "返回值"
// @Router /v3/forest/AbortUpload [post]
func AbortUpload(ctx *gin.Context, req *dtoforestfile.AbortUploadRequest, resp *dtoforestfile.AbortUploadResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}

	errInfo := svcforestfile.AbortUpload(ctx, req)
	if errInfo != nil {
		resp.Code = errInfo.Code
		resp.Message = errInfo.Message
	}
}

// RenewUploadUrl 重新生成预签名
// @Tags 知识森林文件
// @Summary 重新生成预签名
// @Description 重新生成预签名
// @Param user body dtoforestfile.RenewUploadUrlRequest true "入参"
// @Success 200 {object} dtoforestfile.RenewUploadUrlResponse "返回新的预签名信息"
// @Router /v3/forest/RenewUploadUrl [post]
func RenewUploadUrl(ctx *gin.Context, req *dtoforestfile.RenewUploadUrlRequest, resp *dtoforestfile.RenewUploadUrlResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}

	urls, errInfo := svcforestfile.RenewUploadUrl(ctx, req)
	if errInfo != nil {
		resp.Code = errInfo.Code
		resp.Message = errInfo.Message
		return
	}
	resp.Response.RenewedUrls = urls
}
