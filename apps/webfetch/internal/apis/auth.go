package apis

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ygpkg/yg-go/apis/apiobj"
	runtimeauth "github.com/ygpkg/yg-go/apis/runtime/auth"
)

func hasAPIKey(expected string) gin.HandlerFunc {
	expected = strings.TrimSpace(expected)
	return func(ctx *gin.Context) {
		actual := strings.TrimSpace(strings.TrimPrefix(ctx.GetHeader("authorization"), runtimeauth.AuthBearer))
		if expected == "" || subtle.ConstantTimeCompare([]byte(actual), []byte(expected)) != 1 {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, apiobj.BaseResponse{Code: http.StatusUnauthorized, Message: "unauthorized"})
			return
		}
		ctx.Next()
	}
}
