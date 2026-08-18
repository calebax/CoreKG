package apis

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/company"
	"github.com/insmtx/corekg/apps/account/models/employee"
	"github.com/insmtx/corekg/apps/account/models/user"
	adminuser "github.com/insmtx/corekg/apps/admin/models/user"
	"github.com/insmtx/corekg/apps/kechat/models/agentperm"
	"github.com/insmtx/corekg/apps/kecore/models/perm"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/apis/runtime/auth"
	"github.com/ygpkg/yg-go/logs"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// GetBindCompanyKeyWithPermSet 生成带权限的绑定公司密钥
// @Tags Company
// @Summary 生成带权限的绑定公司密钥
// @Description 生成带权限的绑定公司密钥
// @Router /account.GetBindCompanyKeyWithPermSet [post]
// @Param user body GetBindCompanyKeyWithPermSetRequest true "入参"
// @Success 200 {object} GetBindCompanyKeyWithPermSetResponse "返回值"
func GetBindCompanyKeyWithPermSet(ctx *gin.Context, req *GetBindCompanyKeyWithPermSetRequest, resp *GetBindCompanyKeyWithPermSetResponse) {
	if req.Validity(&resp.BaseResponse); resp.BaseResponse.Code != 0 {
		logs.WarnContextf(ctx, "PersonAuth: validity failed,err = %v", resp.BaseResponse.Message)
		return
	}
	if req.Request.Expired == 0 {
		req.Request.Expired = time.Hour * 24
	} else {
		req.Request.Expired = req.Request.Expired * time.Second
	}
	invitation, err := company.CreateInvitationWithPermSet(ctx,
		runtime.CompanyID(ctx), runtime.Uin(ctx),
		req.Request.Count, req.Request.Issuer, req.Request.InvitationRole,
		time.Now().Add(req.Request.Expired), req.Request.PermSet, req.Request.DepartmentIDs)
	if err != nil {
		logs.ErrorContextf(ctx, "GetBindCompanyKeyWithPermSet: CreateInvitationWithPermSet failed, %+v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_create_invitation_failed" // 创建邀请失败
		return
	}

	resp.Response.Key = invitation.Key
}

// BindCompanyWithPermSet 绑定公司
// @Tags Company
// @Summary 绑定公司
// @Description 绑定公司
// @Router /account.BindCompanyWithPermSet [post]
// @Param user body BindCompanyRequest true "入参"
// @Success 200 {object} BindCompanyResponse "返回值"
func BindCompanyWithPermSet(ctx *gin.Context, req *BindCompanyRequest, resp *BindCompanyResponse) {
	if req.Validity(resp); resp.Message != "" {
		logs.WarnContextf(ctx, "BindCompanyWithPermSet: validity failed,err = %v", resp.Message)
		return
	}
	var (
		way        auth.LoginWay
		userInfo   *accounttype.User
		err        error
		weuserInfo *user.UserInfo
	)
	// 通过key获取绑定信息
	invitation, err := company.GetInvitationByKey(req.Request.Key)
	if err != nil {
		logs.WarnContextf(ctx, "BindCompanyWithPermSet: GetInvitationByKey failed, %+v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_get_invitation_failed" // 获取信息失败
		return
	}
	// 判断邀请码是否有效
	if invitation.IsExpire() || invitation.Count == 0 || invitation.AlreadyBind {
		logs.WarnContextf(ctx, "BindCompanyWithPermSet: invitation status error")
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invitation_expired" // 邀请码已失效
		return
	}

	switch req.Request.Way {
	case "wechat_web":
		way = auth.LoginWayWxWeb
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
		weuserInfo, userInfo, err = loginWechatWeb(ctx, loginThird)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = dbutil.Account().Transaction(func(tx *gorm.DB) error {
				userInfo, err = user.CreateUserByBindLogin(ctx, tx, weuserInfo)
				if err != nil {
					logs.ErrorContextf(ctx, "BindCompanyWithPermSet: CreateUserByWechatWebLogin failed, %+v", err)
					return err
				}
				return nil
			})
		} else if err != nil {
			logs.ErrorContextf(ctx, "BindCompanyWithPermSet: GetUserByWechatUnionID failed, %+v", err)
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "account_get_user_info_failed" // 获取用户信息失败
			return
		}
	case "password_web":
		way = auth.LoginWayEmail
		var isEmail, newUser bool
		if strings.Contains(req.Request.UserName, "@") {
			isEmail = true
			userInfo, err = user.GetUserByEmail(req.Request.UserName)
		} else {
			userInfo, err = user.GetUserByPhone(req.Request.UserName)
		}

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				//? if user not exist, create it
				newUser = true
				opt := &adminuser.CreateUserOption{
					Name:     req.Request.UserName,
					Password: req.Request.Password,
				}
				if isEmail {
					opt.Email = req.Request.UserName
				} else {
					opt.Phone = req.Request.UserName
				}

				userInfo, err = adminuser.CreateUser(ctx, opt)
				if err != nil {
					logs.ErrorContextf(ctx, "adminuser.CreateUser err: %v", err)
					resp.Code = errcode.ErrCode_InternalError
					resp.Message = "account_create_user_failed"
					return
				}
			} else {
				//! unexpected err
				logs.ErrorContextf(ctx, "GetUserByUserName(username:%v) faild err: %v", req.Request.UserName, err)
				resp.Code = errcode.ErrCode_InternalError
				resp.Message = "account_get_user_info_failed"
				return
			}
		}
		//? if not new created user then need check password valid
		if !newUser {
			err = bcrypt.CompareHashAndPassword([]byte(*userInfo.Password), []byte(req.Request.Password))
			if err != nil {
				logs.WarnContextf(ctx, "BindCompanyWithPermSet: password not match for (user_id:%v), %v", userInfo.ID, err)
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "account_invalid_user_or_password" // 用户或密码错误
				return
			}
		}

	default:
		logs.WarnContextf(ctx, "Unknown login way: %v", req.Request.Way)
		way = auth.LoginWayUnknown
	}

	// 获取公司信息
	com, err := company.GetCompany(invitation.CompanyID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetBindCompanyWithPermSet: GetCompany failed, %+v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_get_company_info_failed" // 获取公司信息失败
		return
	}

	// 获取用户标识信息
	uins, err := user.GetUserUinsByUserID(ctx, userInfo.ID, invitation.Issuer)
	if err != nil {
		logs.ErrorContextf(ctx, "loginSuccess: get user uins failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_get_user_uins_failed" // 获取用户标识信息失败
		return
	}
	var emp *accounttype.Employee
	// 判断用户是否已经绑定公司
	for _, uin := range uins {
		if uin.SubjectType == accounttype.SubjectTypeCompany {
			if uin.SubjectID == com.ID {
				emp, err = employee.GetEmployeeByUin(uin.ID)
				if err != nil {
					logs.ErrorContextf(ctx, "BindCompanyWithPermSet: get employee failed, %v", err)
					resp.Code = errcode.ErrCode_InternalError
					resp.Message = "account_get_employee_failed" // 获取员工信息失败
					return
				}
				if emp.SysRole == invitation.InvitationRole {
					logs.WarnContextf(ctx, "BindCompanyWithPermSet: user has bind company")
					resp.Code = errcode.ErrCode_InternalError
					resp.Message = "用户已绑定公司"
					return
				}

			}
		}
	}
	var (
		ui      *accounttype.UserIdentification
		relEmps []accounttype.AccountRelEmployeeDepartment
	)
	// 绑定公司
	err = dbutil.Account().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 创建员工身份标识

		// 检查是否已存在相同用户的身份标识
		if company.ExistEmployee(ctx, userInfo.ID, com.ID, accounttype.UinStatusNormal, invitation.Issuer) {
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "account_user_already_bound" // 用户已绑定公司
			return fmt.Errorf("ExistEmployee with user_id[%v] company_id[%v] uin_status[%v] issuer[%v]",
				userInfo.ID, com.ID, accounttype.UinStatusNormal, invitation.Issuer)
		}

		ui, err = company.CreateEmployeeIdentification(ctx, tx, userInfo.ID, com.ID, invitation.Issuer, userInfo.Name)
		if err != nil {
			logs.ErrorContextf(ctx, "CompanyAuth: CreateEmployeeIdentification failed, %+v", err)
			return err
		}

		emp := &accounttype.Employee{
			CompanyID: com.ID,
			UserID:    userInfo.ID,
			Uin:       ui.ID,
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

		for i, v := range invitation.DepartmentIDs.Slice() {
			r := accounttype.AccountRelEmployeeDepartment{
				Uin:          ui.ID,
				DepartmentID: v,
				EmployeeID:   emp.ID,
				CompanyID:    com.ID,
				IsPrimary:    -1,
			}
			if i == 0 {
				r.IsPrimary = 1
			}
			relEmps = append(relEmps, r)
		}

		//批量创建部门规则
		return tx.CreateInBatches(relEmps, len(relEmps)).Error
	})
	if err != nil {
		logs.ErrorContextf(ctx, "BindCompanyWithPermSet CompanyAuth: failed, %+v", err)
		return
	}

	//=========================Construct PermSet And Grant Perm====================================

	if len(invitation.PermSet.ForestPs) > 0 {
		//绑定Forest权限
		fps := perm.NewWrapperPermSet(ctx, ui.ID, com.ID, invitation.PermSet.ForestPs, nil)
		if err := fps.BuildCurrentPermMap(); err != nil {
			logs.ErrorContextf(ctx, "[BindCompanyWithPermSet] [forest]ApplyChanges failed, %+v", err)
			runtime.InternalError(ctx, err)
			return
		}
		if err := fps.ApplyChanges(); err != nil {
			logs.ErrorContextf(ctx, "[BindCompanyWithPermSet] [forest]ApplyChanges failed, %+v", err)
			runtime.InternalError(ctx, err)
			return
		}
	}

	if len(invitation.PermSet.ChatPs) > 0 {
		//绑定Chat权限
		cps := agentperm.NewWrapperPermSet(ctx, ui.ID, com.ID, invitation.PermSet.ChatPs, nil)
		if err := cps.BuildCurrentPermMap(); err != nil {
			logs.ErrorContextf(ctx, "[BindCompanyWithPermSet] [chat]BuildCurrentPermMap failed, %+v", err)
			runtime.InternalError(ctx, err)
			return
		}
		if err := cps.ApplyChanges(); err != nil {
			logs.ErrorContextf(ctx, "[BindCompanyWithPermSet] [chat]ApplyChanges failed, %+v", err)
			runtime.InternalError(ctx, err)
			return
		}
	}

	//=========================================Done=================================================

	token := user.GenerateJwtToken(ctx, ui.ID, way, runtime.GetRealIP(ctx.Request), invitation.Issuer)
	if token == "" {
		logs.ErrorContextf(ctx, "loginSuccess: generate jwt token failed")
		resp.Code = errcode.ErrCode_InternalError
		return
	}
	uins = append(uins, ui)
	// 分类uin
	luin, err := getUinType(ctx, uins, invitation.Issuer)
	if err != nil {
		logs.ErrorContextf(ctx, "loginSuccess: get uin type failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		return
	}
	resp.Response.Uin = luin

	resp.Response.UserID = userInfo.ID
	resp.Response.LoginStatus = "success"
	resp.Response.UserInfo = &user.UserInfo{
		Uin:       ui.ID,
		Identify:  userInfo.Identify,
		AvatarURL: userInfo.AvatarURL,
		Bio:       userInfo.Bio,
		Name:      userInfo.Name,
	}
	resp.Response.JwtToken = token
	resp.Response.Issuer = invitation.Issuer
	resp.Response.LoginWay = auth.LoginWayWxWeb
}

// GetInviteInfo 获取邀请信息
// @Tags Company
// @Summary 获取邀请信息
// @Description 获取邀请信息
// @Router /account.GetInviteInfo [post]
// @Param user body GetInviteInfoRequest true "入参"
// @Success 200 {object} GetInviteInfoResponse "返回值"
func GetInviteInfo(ctx *gin.Context, req *GetInviteInfoRequest, resp *GetInviteInfoResponse) {
	if req.Validity(resp); resp.BaseResponse.Code != 0 {
		logs.ErrorContextf(ctx, "GetInviteInfo: validity failed,err = %v", resp.BaseResponse.Message)
		return
	}

	inv, err := company.GetInviteByKey(req.Request.Key)
	if err != nil {
		logs.ErrorContextf(ctx, "GetInviteInfo: GetInviteByKey failed, %+v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_query_key_failed" // 查询key失败
		return
	}

	if inv.IsExpire() || inv.Count <= 0 {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_invitation_expired" // 邀请码已失效
		return
	}
	cmp, err := company.GetCompany(inv.CompanyID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetInviteInfo: GetCompany failed, %+v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_get_company_info_failed" // 获取公司信息失败
		return
	}

	u, err := user.GetUserByUin(ctx, inv.Uin)
	if err != nil {
		logs.ErrorContextf(ctx, "GetInviteInfo: GetUserByUin failed, %+v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_get_inviter_info_failed" // 获取邀请人信息失败
		return
	}

	resp.Response.InviterName = u.Name
	resp.Response.CompanyName = cmp.Name
}
