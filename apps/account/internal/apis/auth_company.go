package apis

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/internal/dto/dtocompany"
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/company"
	"github.com/insmtx/corekg/apps/account/models/employee"
	"github.com/insmtx/corekg/apps/account/models/user"
	"github.com/insmtx/corekg/apps/account/services/svccompany"
	"github.com/insmtx/corekg/apps/kecore/services/svccoze"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/apis/runtime/auth"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

// CompanyAuth 认证公司
// @Tags Company
// @Summary 认证公司
// @Description 认证公司
// @Router /account.CompanyAuth [post]
// @Param user body CompanyAuthRequest true "入参"
// @Success 200 {object} CompanyAuthResponse "返回值"
func CompanyAuth(ctx *gin.Context, req *CompanyAuthRequest, resp *CompanyAuthResponse) {
	if req.Validity(resp); resp.Message != "" {
		return
	}
	// 获取uin信息
	uin, err := user.GetUserIdentificationByUIN(ctx, runtime.Uin(ctx))
	if err != nil {
		logs.ErrorContextf(ctx, "ChooseUin: get uin failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_get_user_info_failed"
		return
	}

	// u, err := user.GetUserByID(uin.UserID)
	// if err != nil {
	// 	logs.Errorf("CompanyAuth: GetUserByID failed, %+v", err)
	// 	resp.Code = errcode.ErrCode_BadRequest
	// 	resp.Message = "account_user_not_exist"
	// 	return
	// }

	// individual, err := user.GetIndividual(uin.UserID)
	// if err != nil {
	// 	logs.Errorf("ReviewPersonAuth: GetIndividual failed, %+v", err)
	// 	resp.Code = errcode.ErrCode_InternalError
	// 	resp.Message = "account_get_individual_failed"
	// 	return
	// }

	err = dbutil.Account().Transaction(func(tx *gorm.DB) error {
		// 创建公司
		com, err := company.CreateCompany(ctx, tx, req.Request.CompanyInfo)
		if err != nil {
			logs.ErrorContextf(ctx, "CompanyAuth: CreateCompany failed, %+v", err)
			return err
		}
		// 创建员工身份标识
		uin, err := company.CreateEmployeeIdentification(ctx, tx, uin.UserID, com.ID, uin.Issuer, uin.Name)
		if err != nil {
			logs.ErrorContextf(ctx, "CompanyAuth: CreateEmployeeIdentification failed, %+v", err)
			return err
		}
		// err = user.CreateIndividual(tx, uin.UserID)
		// if err != nil {
		// 	logs.Errorf("CompanyAuth: CreateEmployeeIndividual failed, %+v", err)
		// 	return err
		// }
		emp := &accounttype.Employee{
			CompanyID: com.ID,
			UserID:    uin.UserID,
			Uin:       uin.ID,
			SysRole:   accounttype.SysRoleTeacher,
		}
		// 创建员工
		if err := employee.CreateEmployee(ctx, tx, emp); err != nil {
			logs.ErrorContextf(ctx, "CompanyAuth: CreateEmployee failed, %+v", err)
			return err
		}
		return nil
	})
	if err != nil {
		runtime.InternalError(ctx, err)
		return
	}
}

// ListCompany 获取公司列表
// @Tags Company
// @Summary 获取公司列表
// @Description 获取公司列表
// @Router /account.ListCompany [post]
// @Param user body ListCompanyRequest true "入参"
// @Success 200 {object} ListCompanyResponse "返回值"
func ListCompany(ctx *gin.Context, req *ListCompanyRequest, resp *ListCompanyResponse) {
	if req.Validity(resp); resp.Message != "" {
		return
	}
	companys, err := company.QueryCompanyList(ctx, req.Request.PageQuery)
	if err != nil {
		logs.ErrorContextf(ctx, "ListCompanyAuth: QueryCompanyList failed, %+v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_get_company_list_failed"
		return
	}
	resp.BaseResponse = companys.BaseResponse
	resp.Response = companys.Response
	resp.Response.QueryResponse = companys.Response.QueryResponse
}

// GetCompany 获取公司详情
// @Tags Company
// @Summary 获取公司详情
// @Description 获取公司详情
// @Router /account.GetCompany [post]
// @Param user body GetCompanyRequest true "入参"
// @Success 200 {object} GetCompanyResponse "返回值"
func GetCompany(ctx *gin.Context, req *GetCompanyRequest, resp *GetCompanyResponse) {
	if req.Validity(resp); resp.Message != "" {
		return
	}
	com, err := company.GetCompany(req.Request.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetCompany: GetCompany failed, %+v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_get_company_info_failed"
		return
	}
	resp.Response.Company = com
}

// ReviewCompanyAuth 审阅公司信息
// @Tags Company
// @Summary 审阅公司信息
// @Description 审阅公司信息
// @Router /account.ReviewCompanyAuth [post]
// @Param user body ReviewCompanyAuthRequest true "入参"
// @Success 200 {object} ReviewCompanyAuthResponse "返回值"
func ReviewCompanyAuth(ctx *gin.Context, req *ReviewCompanyAuthRequest, resp *ReviewCompanyAuthResponse) {
	if req.Request.CompanyID == 0 {
		logs.ErrorContextf(ctx, "ReviewCompanyAuth: invalid params")
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_parameters"
		return
	}
	com, err := company.GetCompany(req.Request.CompanyID)
	if err != nil {
		logs.ErrorContextf(ctx, "ReviewCompanyAuth: GetCompany failed, %+v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_get_company_info_failed"
		return
	}
	if req.Request.Review {
		com.CompanyStatus = accounttype.CompanyStatusPassed
	} else {
		com.CompanyStatus = accounttype.CompanyStatusFialed
	}
	err = company.UpdateCompany(com)
	if err != nil {
		logs.ErrorContextf(ctx, "ReviewCompanyAuth: UpdateCompany failed, %+v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_save_company_info_failed"
		return
	}
}

// GetBindCompanyKey 生成绑定公司密钥
// @Tags Company
// @Summary 生成绑定公司密钥
// @Description 生成绑定公司密钥
// @Router /account.GetBindCompanyKey [post]
// @Param user body GetBindCompanyKeyRequest true "入参"
// @Success 200 {object} GetBindCompanyKeyResponse "返回值"
func GetBindCompanyKey(ctx *gin.Context, req *GetBindCompanyKeyRequest, resp *GetBindCompanyKeyResponse) {
	if req.Validity(&resp.BaseResponse); resp.BaseResponse.Code != 0 {
		logs.WarnContextf(ctx, "GetBindCompanyKey: validity failed, err = %v", resp.BaseResponse.Message)
		return
	}
	com, err := company.GetCompany(runtime.CompanyID(ctx))
	if err != nil {
		logs.ErrorContextf(ctx, "GetBindCompanyKey: GetCompany failed, %+v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_get_company_info_failed"
		return
	}
	if req.Request.Expired == 0 {
		req.Request.Expired = time.Hour * 24
	} else {
		req.Request.Expired = req.Request.Expired * time.Second
	}
	invitation, err := company.CreateInvitation(com.ID, req.Request.Count, req.Request.Issuer, req.Request.InvitationRole, time.Now().Add(req.Request.Expired))
	if err != nil {
		logs.ErrorContextf(ctx, "GetBindCompanyKey: CreateInvitation failed, %+v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_create_invitation_failed"
		return
	}

	resp.Response.Key = invitation.Key
}

// BindCompany 绑定公司
// @Tags Company
// @Summary 绑定公司
// @Description 绑定公司
// @Router /account.BindCompany [post]
// @Param user body BindCompanyRequest true "入参"
// @Success 200 {object} BindCompanyResponse "返回值"
func BindCompany(ctx *gin.Context, req *BindCompanyRequest, resp *BindCompanyResponse) {
	if req.Validity(resp); resp.Message != "" {
		logs.WarnContextf(ctx, "BindCompany: validity failed,err = %v", resp.Message)
		return
	}
	var way auth.LoginWay
	if req.Request.Way == "wechat_web" {
		way = auth.LoginWayWxWeb
	}
	// 通过key获取绑定信息
	invitation, err := company.GetInvitationByKey(req.Request.Key)
	if err != nil {
		logs.WarnContextf(ctx, "BindCompany: GetInvitationByKey failed, %+v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_get_invitation_failed"
		return
	}
	// 判断邀请码是否有效
	if invitation.IsExpire() || invitation.Count == 0 || invitation.AlreadyBind {
		logs.WarnContextf(ctx, "BindCompany: invitation status error")
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invitation_expired"
		return
	}

	loginThird := &LoginThirdRequest{
		Request: struct {
			Way        string      `json:"way"`
			Code       string      `json:"code"`
			DomainName string      `json:"domain_name"`
			Option     LoginOption `json:"option,omitempty"`
		}{
			Way:        req.Request.Way,
			Code:       req.Request.Code,
			DomainName: req.Request.DomainName,
		},
	}
	weuserInfo, userInfo, err := loginWechatWeb(ctx, loginThird)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = dbutil.Account().Transaction(func(tx *gorm.DB) error {
			userInfo, err = user.CreateUserByBindLogin(ctx, tx, weuserInfo)
			if err != nil {
				logs.ErrorContextf(ctx, "BindCompany: CreateUserByWechatWebLogin failed, %+v", err)
				return err
			}
			err = user.CreateIndividual(ctx, tx, userInfo.ID)
			if err != nil {
				logs.ErrorContextf(ctx, "BindCompany: CreateIndividual failed, %+v", err)
				return err
			}
			return nil
		})
	} else if err != nil {
		logs.ErrorContextf(ctx, "BindCompany: GetUserByWechatUnionID failed, %+v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_get_user_info_failed"
		return
	}

	// 获取公司信息
	com, err := company.GetCompany(invitation.CompanyID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetBindCompanyKey: GetCompany failed, %+v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_get_company_info_failed"
		return
	}

	// 获取用户标识信息
	uins, err := user.GetUserUinsByUserID(ctx, userInfo.ID, invitation.Issuer)
	if err != nil {
		logs.ErrorContextf(ctx, "loginSuccess: get user uins failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		return
	}
	var emp *accounttype.Employee
	// 判断用户是否已经绑定公司
	for _, uin := range uins {
		if uin.SubjectType == accounttype.SubjectTypeCompany {
			if uin.SubjectID == com.ID {
				emp, err = employee.GetEmployeeByUin(uin.ID)
				if err != nil {
					logs.ErrorContextf(ctx, "BindCompany: get employee failed, %s", err)
					resp.Code = errcode.ErrCode_InternalError
					resp.Message = "account_get_employee_failed"
					return
				}
				if emp.SysRole == invitation.InvitationRole {
					logs.WarnContextf(ctx, "BindCompany: user has bind company")
					resp.Code = errcode.ErrCode_InternalError
					resp.Message = "account_user_already_bound"
					return
				}

			}
		}
	}
	var uident *accounttype.UserIdentification
	// 绑定公司
	err = dbutil.Account().Transaction(func(tx *gorm.DB) error {
		// 创建员工身份标识
		uident, err = company.CreateEmployeeIdentification(ctx, tx, userInfo.ID, com.ID, invitation.Issuer, userInfo.Name)
		if err != nil {
			logs.ErrorContextf(ctx, "CompanyAuth: CreateEmployeeIdentification failed, %+v", err)
			return err
		}
		// _, err = company.CreateEmployeeIndividual(tx, createdUin.ID, weuserInfo.AvatarURL, req.Request.RealName, req.Request.IDCard)
		// if err != nil {
		// 	logs.Errorf("CompanyAuth: CreateEmployeeIndividual failed, %+v", err)
		// 	return err
		// }
		emp := &accounttype.Employee{
			CompanyID: com.ID,
			UserID:    userInfo.ID,
			Uin:       uident.ID,
			SysRole:   invitation.InvitationRole,
		}
		// 创建员工
		if err := employee.CreateEmployee(ctx, tx, emp); err != nil {
			logs.ErrorContextf(ctx, "CompanyAuth: CreateEmployee failed, %+v", err)
			return err
		}
		// 更新邀请信息
		invitation.Count--
		if invitation.Count == 0 {
			invitation.AlreadyBind = true
		}
		if err := company.UpdateInvitation(tx, invitation); err != nil {
			logs.ErrorContextf(ctx, "CompanyAuth: UpdateInvitation failed, %+v", err)
			return err
		}
		return nil
	})
	if err != nil {
		runtime.InternalError(ctx, err)
		return
	}
	token := user.GenerateJwtToken(ctx, uident.ID, way, runtime.GetRealIP(ctx.Request), invitation.Issuer)
	if token == "" {
		logs.ErrorContextf(ctx, "loginSuccess: generate jwt token failed")
		resp.Code = errcode.ErrCode_InternalError
		return
	}
	uins = append(uins, uident)
	// 分类uin
	luin, err := getUinType(ctx, uins, resp.Response.Issuer)
	if err != nil {
		logs.ErrorContextf(ctx, "loginSuccess: get uin type failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		return
	}
	resp.Response.Uin = luin
	resp.Response.UserID = userInfo.ID
	resp.Response.LoginStatus = "success"
	resp.Response.UserInfo = &user.UserInfo{
		Identify:  userInfo.Identify,
		AvatarURL: userInfo.AvatarURL,
		Bio:       userInfo.Bio,
		Name:      userInfo.Name,
	}
	// resp.Response.Uin = luin
	resp.Response.JwtToken = token
	resp.Response.Issuer = invitation.Issuer
	resp.Response.LoginWay = auth.LoginWayWxWeb

	_ = svccoze.SpaceSyncWithToken(ctx, token)
}

// CreateCompany 创建公司
// @Tags Company
// @Summary 创建公司
// @Description 创建公司
// @Router /account.CreateCompany [post]
// @Param user body dtocompany.CreateCompanyRequest true "入参"
// @Success 200 {object} dtocompany.CreateCompanyResponse "返回值"
func CreateCompany(ctx *gin.Context, req *dtocompany.CreateCompanyRequest, resp *dtocompany.CreateCompanyResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}

	uin := runtime.Uin(ctx)
	if uin == 0 {
		refreshToken, err := user.GetRefreshToken(req.Request.UserID)
		if err != nil {
			logs.ErrorContextf(ctx, "CreateCompany: get refresh token failed, %s", err)
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "account_refresh_token_mismatch" // 初始化用户身份失败
			return
		}
		if refreshToken != req.Request.RefreshToken {
			logs.ErrorContextf(ctx, "CreateCompany: refresh token not match, %s", err)
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "account_refresh_token_mismatch" // 身份校验失败
			return
		}
	} else {
		userInfo, err := user.GetUserByUin(ctx, uin)
		if err != nil {
			logs.ErrorContextf(ctx, "CreateCompany: get user by uin failed, %s", err)
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "account_get_user_failed" // 获取用户失败
			return
		}
		req.Request.UserID = userInfo.ID
	}

	if !user.CanCreateCompany(ctx, req.Request.UserID) {
		logs.ErrorContextf(ctx, "CreateCompany: user can not create company, userID:%d", req.Request.UserID)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_company_quota_exceeded"
		return
	}

	cmp, errResp := svccompany.CreateCompany(ctx, req)
	if errResp != nil {
		resp.Code = errResp.Code
		resp.Message = errResp.Message
		return
	}
	resp.Response = cmp
}
