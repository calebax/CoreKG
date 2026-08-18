package deployhandle

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/version"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/settings"
)

// SwitchPrivateEvn 切换私有化环境
func SwitchPrivateEvn(ctx *gin.Context, req *apiobj.BaseRequest, resp *NowDeployModeResponse) {
	if version.DeployMode() != global.DeployModeOpenPO {
		version.SetDeployMode(global.DeployModeOpenPO)
		resp.Response.DeployMode = version.DeployMode()
		settings.SetText("knowledge", "deploy", global.DeployModeOpenPO)
		return
	}
	version.SetDeployMode("")
	settings.SetText("knowledge", "deploy", "")
	resp.Response.DeployMode = version.DeployMode()
}

// NowDeployMode 获取当前部署模式
func NowDeployMode(ctx *gin.Context, req *apiobj.BaseRequest, resp *NowDeployModeResponse) {
	mode, _ := settings.GetText("knowledge", "deploy")
	if mode != version.DeployMode() {
		version.SetDeployMode(mode)
	}
	resp.Response.DeployMode = version.DeployMode()
}

// NowDeployModeResponse .
type NowDeployModeResponse struct {
	apiobj.BaseResponse
	Response struct {
		DeployMode string `json:"deploy_mode"`
	}
}

// GetMode 程序初始化的deploy
func GetMode() {
	mode, _ := settings.GetText("knowledge", "deploy")
	version.SetDeployMode(mode)
}
