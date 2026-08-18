package license

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/insmtx/corekg/apps/admin/models/admintype"
	"github.com/insmtx/corekg/apps/admin/models/license"
	corekglicense "github.com/insmtx/corekg/apps/corekg/models/license"
	"github.com/ygpkg/yg-go/settings"

	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
)

// GenerateLicense 生成License
// @Tags License管理
// @Summary 生成License
// @Description 生成License
// @Router /admin.GenerateLicense [post]
// @Param user body GenerateLicenseRequest true "入参"
// @Success 200 {object} apiobj.BaseResponse "返回值"
func GenerateLicense(ctx *gin.Context, req *GenerateLicenseRequest, resp *apiobj.BaseResponse) {
	if req.Validate(resp); resp.Code != 0 {
		logs.WarnContextf(ctx, "GenerateLicense validate params failed")
		return
	}

	//construct base license
	l := &admintype.License{
		Subject:    req.Request.Subject,
		Issuer:     req.Request.Issuer,
		Env:        req.Request.Env,
		UID:        req.Request.UID,
		Serial:     uuid.NewString(),
		Note:       req.Request.Note,
		VersionKey: req.Request.VersionKey,
	}
	//set expire time
	if req.Request.ExpiredAt == 0 {
		req.Request.ExpiredAt = 99 * 365
	}
	exp := time.Now().Add(time.Duration(req.Request.ExpiredAt) * 24 * time.Hour)
	l.ExpiredAt = &exp

	//generate key pair
	cfg := corekglicense.KeyConfig{}
	if err := settings.GetYaml("admin", "license_key", &cfg); err != nil {
		logs.ErrorContextf(ctx, "GenerateLicense GetYaml get failed, err=%v", err)
		runtime.InternalError(ctx, "获取密钥配置失败")
		return
	}
	l.PublicKey = cfg.Public
	l.PrivateKey = cfg.Private

	//generate license
	if err := license.GenerateLicense(ctx, l); err != nil {
		logs.ErrorContextf(ctx, "GenerateLicense GenerateLicense failed: %v", err)
		runtime.InternalError(ctx, "生成License失败")
		return
	}

	if err := license.CreateLicense(dbutil.Account(), l); err != nil {
		logs.ErrorContextf(ctx, "[GenerateLicense] [CreateLicense][ENV:%v | UID:%v] failed: %v", req.Request.Env, req.Request.UID, err)
		runtime.InternalError(ctx, "插入License失败")
		return
	}
}

// ListLicense 查询License列表
// @Tags License管理
// @Summary 查询License列表
// @Description 查询License列表
// @Router /admin.ListLicense [post]
// @Param user body ListLicenseRequest true "入参"
// @Success 200 {object} ListLicenseResponse "返回值"
func ListLicense(ctx *gin.Context, req *ListLicenseRequest, resp *ListLicenseResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		logs.WarnContextf(ctx, "ListLicense validate params failed")
		return
	}
	if err := license.QueryLicenseList(ctx, req.Request, &resp.Response); err != nil {
		logs.ErrorContextf(ctx, "ListLicense QueryLicenseList failed: %v", err)
		runtime.InternalError(ctx, "查询license列表失败")
		return
	}

}

// DownloadLicense 下载license
// @Tags License管理
// @Summary 下载license
// @Description 下载license
// @Router /admin.DownloadLicense [post]
// @Param id formData string true "licenseID"
// @Param type formData string false "下载类型" default(license) Enums(license, key)
// @Success 200 {file} file "成功下载license"
func DownloadLicense(ctx *gin.Context) {
	licenseIDStr := ctx.Request.FormValue("id")
	lID, err := strconv.Atoi(licenseIDStr)
	if err != nil {
		logs.ErrorContextf(ctx, "DownloadLicense parse form[%v] failed, %v", licenseIDStr, err)
		runtime.BadRequest(ctx, "licenseID解析失败")
		return
	}
	var lic *admintype.License
	if err = dbutil.Account().
		Where("deleted_at is null").
		Find(&lic, lID).
		Error; err != nil {
		logs.ErrorContextf(ctx, "DownloadLicense GetLicense(%v) failed, %v", lID, err)
		runtime.InternalError(ctx, "查询license失败")
		return
	}

	tp := ctx.Request.FormValue("type")
	ctx.Writer.Header().Set("Content-Type", "application/octet-stream")
	switch tp {
	case "license":
		ctx.Writer.Header().Set("Content-Disposition", "attachment; filename=\"license.dat\"")
		if _, err := ctx.Writer.Write([]byte(lic.Raw)); err != nil {
			logs.ErrorContextf(ctx, "DownloadLicense Write file failed: %v", err)
			runtime.InternalError(ctx, "写入license失败")
			return
		}

	case "key":
		ctx.Writer.Header().Set("Content-Disposition", "attachment; filename=\"public.pem\"")
		if _, err := ctx.Writer.Write([]byte(lic.PublicKey)); err != nil {
			logs.ErrorContextf(ctx, "DownloadLicense Write file failed: %v", err)
			runtime.InternalError(ctx, "写入公钥失败")
			return
		}
	default:
		logs.ErrorContextf(ctx, "DownloadLicense parse form failed, invalid type[%v]", tp)
		runtime.BadRequest(ctx, "未知下载类型")
		return
	}
}

// DistributeLicense 证书分发
// @Tags License管理
// @Summary 证书分发
// @Description 证书分发
// @Router /admin.DistributeLicense [post]
// @Param user body DistributeLicenseRequest true "入参"
// @Success 200 {object} DistributeLicenseResponse "返回值"
func DistributeLicense(ctx *gin.Context, req *DistributeLicenseRequest, resp *DistributeLicenseResponse) {
	if req.Validate(&resp.BaseResponse); resp.Code != 0 {
		logs.WarnContextf(ctx, "DistributeLicense validate params failed")
		return
	}

	//construct base license
	l := &admintype.License{
		Subject:    req.Request.Subject,
		Issuer:     req.Request.Issuer,
		Env:        req.Request.Env,
		UID:        req.Request.UID,
		Serial:     uuid.NewString(),
		Note:       req.Request.Note,
		VersionKey: req.Request.VersionKey,
	}
	//set expire time
	if req.Request.ExpiredAt == 0 {
		req.Request.ExpiredAt = 99 * 365
	}
	exp := time.Now().Add(time.Duration(req.Request.ExpiredAt) * 24 * time.Hour)
	l.ExpiredAt = &exp

	//generate key pair
	cfg := corekglicense.KeyConfig{}
	if err := settings.GetYaml("admin", "license_key_h3c", &cfg); err != nil {
		logs.ErrorContextf(ctx, "GenerateLicense GetYaml get failed, err=%v", err)
		runtime.InternalError(ctx, "获取密钥配置失败")
		return
	}
	logs.DebugContextf(ctx, "get public key=%+v", cfg)
	l.PublicKey = cfg.Public
	l.PrivateKey = cfg.Private

	//generate license
	if err := license.GenerateLicense(ctx, l); err != nil {
		logs.ErrorContextf(ctx, "GenerateLicense GenerateLicense failed: %v", err)
		runtime.InternalError(ctx, "生成License失败")
		return
	}

	if err := license.CreateLicense(dbutil.Account(), l); err != nil {
		logs.ErrorContextf(ctx, "[GenerateLicense] [CreateLicense][ENV:%v | UID:%v] failed: %v", req.Request.Env, req.Request.UID, err)
		runtime.InternalError(ctx, "插入License失败")
		return
	}
	resp.Response.License = *l
}

// ApplyLicense 证书申请
// @Tags License管理
// @Summary 证书申请
// @Description 证书申请
// @Router /admin.ApplyLicense [post]
// @Param user body DistributeLicenseRequest true "入参"
// @Success 200 {object} apiobj.BaseResponse "返回值"
func ApplyLicense(ctx *gin.Context, req *DistributeLicenseRequest, resp *apiobj.BaseResponse) {
	if req.Validate(resp); resp.Code != 0 {
		logs.WarnContextf(ctx, "ApplyLicense validate params failed")
		return
	}

	logs.DebugContextf(ctx, "ApplyLicense request=%+v", req)
	reqBts, err := json.Marshal(&req)
	if err != nil {
		logs.ErrorContextf(ctx, "ApplyLicense marshal failed, %v", err)
		runtime.InternalError(ctx, "申请license失败")
		return
	}
	logs.DebugContextf(ctx, "ApplyLicense request marshaled =%+v", string(reqBts))

	request, err := http.NewRequest("POST", "https://api.example.com/v2/admin.DistributeLicense", bytes.NewReader(reqBts))
	if err != nil {
		logs.ErrorContextf(ctx, "ApplyLicense new request failed, %v", err)
		runtime.InternalError(ctx, "申请license失败")
		return
	}

	apikey, err := settings.GetValue("admin", "distribute_license_apikey")
	if err != nil {
		logs.ErrorContextf(ctx, "ApplyLicense get apikey failed, %v", err)
		runtime.InternalError(ctx, "获取settings配置apikey失败")
		return
	}

	logs.DebugContextf(ctx, "settings distribute_license apikey=%+v", apikey)

	request.Header.Add("Authorization", "Bearer "+apikey)

	// 设置 Content-Type
	request.Header.Add("Content-Type", "application/json")

	client := &http.Client{Timeout: 20 * time.Second} // 设置一个合理的超时时间
	logs.DebugContextf(ctx, "distributeLicense request:%+v", *request)
	response, err := client.Do(request)
	if err != nil {
		logs.ErrorContextf(ctx, "ApplyLicense post failed, %v", err)
		runtime.InternalError(ctx, "send申请license失败")
		return
	}
	defer response.Body.Close()
	bts, err := io.ReadAll(response.Body)
	if err != nil {
		logs.ErrorContextf(ctx, "ApplyLicense read body failed, %v", err)
		runtime.InternalError(ctx, "读取license失败")
		return
	}
	logs.DebugContextf(ctx, "ApplyLicense response=%+v", string(bts))
	l := &DistributeLicenseResponse{}
	if err = json.Unmarshal(bts, &l); err != nil {
		logs.ErrorContextf(ctx, "ApplyLicense unmarshal failed, %v", err)
		runtime.InternalError(ctx, "解析license失败")
		return
	}
	ls := l.Response.License
	ls.ID = 0
	if err := dbutil.Account().Create(&ls).Error; err != nil {
		logs.ErrorContextf(ctx, "ApplyLicense create failed, %v", err)
		runtime.InternalError(ctx, "插入license失败")
		return
	}
}
