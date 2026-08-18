package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/user"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
)

// UserProfile
// @Tags User
// @Summary 获取用户信息
// @Description 获取用户信息
// @Router /account.Profile [post]
// @Param user body apiobj.BaseRequest true "入参"
// @Success 200 {object} LoginThirdResponse "返回值"
func UserProfile(ctx *gin.Context, _ *apiobj.BaseRequest, resp *LoginThirdResponse) {
	cusID := runtime.Uin(ctx)
	cus, err := user.GetUserByID(cusID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_server_error" // 服务器错误
		return
	}
	resp.Response.UserInfo = &user.UserInfo{
		Identify:  cus.Identify,
		Name:      cus.Name,
		Bio:       cus.Bio,
		AvatarURL: cus.AvatarURL,
	}
}
