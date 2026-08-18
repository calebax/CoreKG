package apis

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

var dbShuziren *gorm.DB

// ExcuteSql 执行sql脚本
// @Tags  执行sql脚本
// @Summary 执行sql脚本
// @Description 执行sql脚本
// @Router /kesearch.ExcuteSql [post]
// @Param request body ExcuteSqlRequest true "request"
// @Success 200 {object} ExcuteSqlResponse "response"
func ExcuteSql(ctx *gin.Context, req *ExcuteSqlRequest, resp *ExcuteSqlResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[RerankSearchChunk] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	if dbShuziren == nil {
		// 通过环境变量读取真实密码，避免把凭据写入源码。
		password := os.Getenv("SHUZIREN_MYSQL_PASSWORD")
		if password == "" {
			logs.ErrorContextf(ctx, "[ExcuteSql] SHUZIREN_MYSQL_PASSWORD 未设置")
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "获取连接失败"
			return
		}
		db, err := dbtools.InitDBConn("shuziren", "mysql://root:"+password+"@CHANGE_ME_HOST:21015/yygu_db?charset=utf8mb4&parseTime=True&loc=Local")
		if err != nil {
			logs.ErrorContextf(ctx, "[ExcuteSql] connect shuziren db failed, %s", err)
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "获取连接失败"
			return
		}
		dbShuziren = db
	}

	// 执行sql，获取全部结果返回
	var result []map[string]interface{}
	err := dbShuziren.Raw(req.Request.Sql).Scan(&result).Error
	if err != nil {
		logs.ErrorContextf(ctx, "[ExcuteSql] exec sql failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "执行sql失败"
		return
	}
	resp.Response.Res = result

}

type ExcuteSqlRequest struct {
	apiobj.BaseRequest
	Request struct {
		Sql string `json:"sql"`
	} `json:"request"`
}

func (req *ExcuteSqlRequest) Validity(resp *ExcuteSqlResponse) {
	if req.Request.Sql == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "sql不能为空"
		return
	}
}

type ExcuteSqlResponse struct {
	apiobj.BaseResponse
	Response struct {
		Res interface{} `json:"res"`
	} `json:"response"`
}
