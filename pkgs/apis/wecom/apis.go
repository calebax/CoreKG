package wecom

import (
	"github.com/gin-gonic/gin"
)

// 企业微信

const (
	DefaultAPIPrefix = "/apis/wecom/v1"
)

func RegistryRouter(rg *gin.RouterGroup) {
	appGroup := rg.Group("namespaces/:namespace/apps/:appid")
	appGroup.POST("", AddApp)
	appGroup.GET("serve", VerifyServerMessage)
	appGroup.POST("serve", ReceiveServerMessage)
}
