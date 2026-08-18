package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/corekg/models/license"
	"github.com/ygpkg/yg-go/logs"
)

// GetClusterID returns the cluster id from gin context.
func GetClusterID(ctx *gin.Context) {
	e, err := license.NewEnvironment(license.EnvTypeKubernetes)
	if err != nil {
		logs.ErrorContextf(ctx, "GetClusterID NewEnvironment failed, %s", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	uid, err := e.GetUID(ctx)
	if err != nil {
		logs.ErrorContextf(ctx, "GetClusterID failed, %s", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	ctx.String(200, uid)
}

func Ping(ctx *gin.Context) {
	ctx.String(200, "pong")
}
