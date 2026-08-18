package notify

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kesale"
	"github.com/insmtx/corekg/apps/kesale/models"
	"github.com/ygpkg/yg-go/logs"
)

func HandleWechatNotify(ctx *gin.Context) {
	handleNotify(ctx, models.ChannelWeChatPay)
}

func handleNotify(ctx *gin.Context, channel models.PaymentChannel) {
	logs.InfoContextf(ctx, "handleNotify: %v %v", channel, ctx.Request.URL)

	resuest := ctx.Request
	result, err := kesale.Manager().HandlePaymentCallback(ctx, channel, resuest)

	httpCode := http.StatusOK
	if err != nil {
		logs.ErrorContextf(ctx, "handleNotify: %v callback failed: %v", channel, err)
		httpCode = http.StatusInternalServerError
	}
	ctx.JSON(httpCode, result)
}
