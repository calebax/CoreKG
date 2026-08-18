package wecom

import "github.com/gin-gonic/gin"

// AddApp 添加企业微信应用
// @Tags 企业微信
// @Summary 添加企业微信应用
// @Produce json
// @Param namespace formData string false "命名空间,企业id"
// @Param appid formData string false "appid"
// @Success 200
// @Router /apis/wecom/v1/namespaces/:namespace/app/:appid [post]
// Success 200 {object} account.User
func AddApp(ctx *gin.Context) {

}
