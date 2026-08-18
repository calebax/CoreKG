package user

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
	random2 "github.com/ygpkg/yg-go/random"
	"gorm.io/gorm"
)

// List 用户列表
func List(ctx context.Context, opt apiobj.PageQuery) ([]*accounttype.User, error) {
	sql := dbutil.Account().Table((&accounttype.User{}).TableName()).
		WithContext(ctx).
		Offset(opt.Offset)
	if !opt.ListAll && opt.Limit > 0 {
		sql = sql.Limit(opt.Limit)
	}

	users := []*accounttype.User{}
	err := sql.Find(&users).Error
	if err != nil {
		logs.ErrorContextf(ctx, "user.List faild err: %v", err)
		return nil, err
	}
	return users, nil
}

// GetUserIdentificationByUIN 通过UIN获取用户标识信息
func GetUserIdentificationByUIN(ctx context.Context, uin uint) (*accounttype.UserIdentification, error) {
	user := &accounttype.UserIdentification{}
	err := dbutil.Account().Table((&accounttype.UserIdentification{}).TableName()).
		WithContext(ctx).
		Where("id = ?", uin).
		Find(user).Error
	if err != nil {
		logs.ErrorContextf(ctx, "[account][%d] found failed, %s", uin, err)
		return nil, err
	}
	return user, nil
}

// GetUserIdentificationByUINs 通过UIN获取用户标识信息
func GetUserIdentificationByUINs(ctx context.Context, uin []uint) ([]*accounttype.UserIdentification, error) {
	users := []*accounttype.UserIdentification{}
	err := dbutil.Account().Table((&accounttype.UserIdentification{}).TableName()).
		Where("id IN ?", uin).
		Find(&users).Error
	if err != nil {
		logs.ErrorContextf(ctx, "[account][%d] found failed, %s", uin, err)
		return nil, err
	}
	return users, nil
}

// GetUserUinsByUserID 通过用户ID获取用户UIN
func GetUserUinsByUserID(ctx context.Context, userid uint, issuer string) ([]*accounttype.UserIdentification, error) {
	var users []*accounttype.UserIdentification
	err := dbutil.Account().Table((&accounttype.UserIdentification{}).TableName()).
		WithContext(ctx).
		Where("user_id = ? AND issuer = ?", userid, issuer).
		Where("deleted_at IS NULL").
		Order("last_login_at desc").
		Find(&users).Error
	if err != nil {
		logs.ErrorContext(ctx, "[account][%d] issuer %v found failed, %s", userid, issuer, err)
		return nil, err
	}
	return users, nil
}

// GetUserByID 根据ID获取客户
func GetUserByID(id uint) (*accounttype.User, error) {
	var u accounttype.User
	err := dbutil.Account().Where("id = ?", id).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetUserByEmail 根据邮箱获取客户
func GetUserByEmail(email string) (*accounttype.User, error) {
	var u accounttype.User
	err := dbutil.Account().Where("email = ?", email).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetUserByPhone 根据手机获取客户
func GetUserByPhone(phone string) (*accounttype.User, error) {
	var u accounttype.User
	err := dbutil.Account().Where("phone = ?", phone).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetUserByGithubID 根据Github ID获取客户
func GetUserByGithubID(githubID uint) (*accounttype.User, error) {
	var u accounttype.User
	err := dbutil.Account().Where("github_id = ?", githubID).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetUserByUin 根据uin获取用户信息
func GetUserByUin(ctx context.Context, uin uint) (*accounttype.User, error) {
	var user accounttype.User
	err := dbutil.Account().
		Table(accounttype.TableNameUser).
		WithContext(ctx).
		Joins("JOIN user_identification uin ON uin.user_id = user.id").
		Where("uin.id = ?", uin).
		Where("uin.uin_status = ?", accounttype.UinStatusNormal).
		First(&user).Error
	if err != nil {
		logs.ErrorContextf(ctx, "GetUserByUin: get user failed, %v", err)
	}
	return &user, nil
}

// GetUserByWorkWechatUserID 根据企业微信用户ID获取客户
func GetUserByWorkWechatUserID(workWechatUserID string) (*accounttype.User, error) {
	var u accounttype.User
	err := dbutil.Account().Where("work_wechat_user_id = ?", workWechatUserID).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetUserByWechatUnionID 根据微信UnionID获取客户
func GetUserByWechatUnionID(unionID string) (*accounttype.User, error) {
	var u accounttype.User
	err := dbutil.Account().Where("wechat_union_id = ?", unionID).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// ExistsUserByIIdentify 根据标识判断客户是否存在
func ExistsUserByIIdentify(ctx context.Context, identify string) (bool, error) {
	var count int64
	err := dbutil.Account().Table(accounttype.TableNameUser).
		WithContext(ctx).
		Where("identify = ?", identify).
		Count(&count).
		Error
	if err != nil {
		logs.ErrorContextf(ctx, "[account][%s] found failed, %s", identify, err)
		return false, err
	}
	return count > 0, nil
}

// GetUserByName 根据用户名获取客户
func GetUserByName(ctx context.Context, username string) (*accounttype.User, error) {
	user := &accounttype.User{}
	err := dbutil.Account().Table(user.TableName()).
		Where("name = ?", username).
		Find(user).Error
	if err != nil {
		logs.ErrorContextf(ctx, "[account][%s] found failed, %s", username, err)
		return nil, err
	}
	if user.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return user, nil
}

// CreateUser 创建用户
func CreateUser(ctx context.Context, tx *gorm.DB, req *UserInfo, oriUserinfo *UserInfo) (*accounttype.User, error) {
	// 初始化客户信息
	cus := &accounttype.User{
		Identify:  oriUserinfo.Identify,
		Name:      oriUserinfo.Name,
		Bio:       req.Bio,
		AvatarURL: oriUserinfo.AvatarURL,
	}
	if oriUserinfo.Email != "" {
		cus.Email = &oriUserinfo.Email
	}
	if oriUserinfo.Phone != "" {
		cus.Phone = &oriUserinfo.Phone
	}
	if oriUserinfo.GithubID > 0 {
		cus.GithubID = &oriUserinfo.GithubID
	}
	if oriUserinfo.WorkWechatUserID != "" {
		cus.WorkWechatUserID = &oriUserinfo.WorkWechatUserID
	}
	if oriUserinfo.WechatUnionID != "" {
		cus.WechatUnionID = &oriUserinfo.WechatUnionID
	}
	if oriUserinfo.WechatWebOpenID != "" {
		cus.WechatWebOpenID = &oriUserinfo.WechatWebOpenID
	}

	// 创建客户
	if err := tx.Create(cus).Error; err != nil {
		logs.ErrorContextf(ctx, "RegisterThird: create user failed, %+v", err)
		return nil, err
	}
	return cus, nil
}

// CreateUserIdentification 创建用户标识
func CreateUserIdentification(ctx context.Context, tx *gorm.DB, userID uint, issuer string) (*accounttype.UserIdentification, error) {
	// 初始化用户标识信息
	userIdentification := &accounttype.UserIdentification{
		UserID:      userID,
		SubjectType: accounttype.SubjectTypeIndividual,
		UinStatus:   accounttype.UinStatusNormal,
		Issuer:      issuer,
	}
	// 创建用户标识
	if err := tx.Create(userIdentification).Error; err != nil {
		logs.ErrorContextf(ctx, "RegisterThird: create uin failed, %+v", err)
		return nil, err
	}

	// 更新 SubjectID 为 userIdentification 的 ID
	userIdentification.SubjectID = userIdentification.ID

	if err := tx.Model(userIdentification).Update("SubjectID", userIdentification.ID).Error; err != nil {
		logs.ErrorContextf(ctx, "RegisterThird: update SubjectID failed, %+v", err)
		return nil, err
	}

	return userIdentification, nil
}

// CreateUserCompanyIdentification 创建用户公司类型标识
func CreateUserCompanyIdentification(ctx context.Context, tx *gorm.DB, userID uint, issuer string) (*accounttype.UserIdentification, error) {
	// 初始化用户标识信息
	userIdentification := &accounttype.UserIdentification{
		UserID:      userID,
		SubjectType: accounttype.SubjectTypeCompany,
		UinStatus:   accounttype.UinStatusNormal,
		Issuer:      issuer,
	}
	// 创建用户标识
	if err := tx.Create(userIdentification).Error; err != nil {
		logs.ErrorContextf(ctx, "CreateUserCompanyIdentification: create uin failed, %+v", err)
		return nil, err
	}

	// 更新 SubjectID 为 userIdentification 的 ID
	userIdentification.SubjectID = userIdentification.ID

	if err := tx.Model(userIdentification).Update("SubjectID", userIdentification.ID).Error; err != nil {
		logs.ErrorContextf(ctx, "CreateUserCompanyIdentification: update SubjectID failed, %+v", err)
		return nil, err
	}

	return userIdentification, nil
}

// CreateIndividual 创建个人账户
func CreateIndividual(ctx context.Context, tx *gorm.DB, userId uint) error {
	// 初始化个人账户信息
	individual := &accounttype.Individual{
		UserID: userId,
	}

	// 创建个人账户
	if err := tx.Create(individual).Error; err != nil {
		logs.ErrorContextf(ctx, "RegisterThird: create individual failed, %+v", err)
		return err
	}
	return nil
}

// GetIndividual 根据 user_id 获取个人账户
func GetIndividual(user_id uint) (*accounttype.Individual, error) {
	var individual accounttype.Individual
	err := dbutil.Account().Where("user_id = ?", user_id).First(&individual).Error
	if err != nil {
		return nil, err
	}
	return &individual, nil
}

// SaveIndividual 保存个人账户
func SaveIndividual(ctx context.Context, individual *accounttype.Individual) error {
	if err := dbutil.Account().Save(individual).Error; err != nil {
		logs.ErrorContextf(ctx, "RegisterThird: create individual failed, %+v", err)
		return err
	}
	return nil
}

// QueryIndividualsList 查询个人账户列表,,所有等待中的
func QueryIndividualsList(opt apiobj.PageQuery, ret *accounttype.IndividualItemList) error {
	query := dbutil.Account().Table(accounttype.TableNameIndividual).
		Where("deleted_at is null").
		Where("real_name_status = ?", accounttype.IndividualStatuPending)

	for _, filter := range opt.Filters {
		switch filter.Field {
		case "real_name":
			query = query.Where("individual.real_name = ?", filter.Value[0])
		case "id_card":
			query = query.Where("individual.id_card = ?", filter.Value[0])
		default:
			return fmt.Errorf("invalid filter field: %s", filter.Field)
		}
	}

	if err := query.Count(&ret.Total).Error; err != nil {
		return err
	}
	if ret.Total == 0 {
		return nil
	}

	// 排序
	if len(opt.OrderBy) > 0 {
		query = query.Order(strings.Join(opt.OrderBy, ","))
	}

	// 分页
	query = query.Offset(opt.Offset)
	if !opt.ListAll && opt.Limit > 0 {
		query = query.Limit(opt.Limit)
	}

	// 查询数据
	err := query.Find(&ret.Data).Error
	if err != nil {
		return err
	}
	return nil
}

// CreateUserByBindLogin 创建用户信息
func CreateUserByBindLogin(ctx context.Context, tx *gorm.DB, user *UserInfo) (*accounttype.User, error) {
	// 初始化用户信息
	cus := &accounttype.User{
		Identify:  user.Identify,
		Name:      user.Name,
		Bio:       user.Bio,
		AvatarURL: user.AvatarURL,
	}
	if user.WechatUnionID != "" {
		cus.WechatUnionID = &user.WechatUnionID
	}
	if user.WechatWebOpenID != "" {
		cus.WechatWebOpenID = &user.WechatWebOpenID
	}
	if user.Email != "" {
		cus.Email = &user.Email
	}

	// 确保 identify 是唯一的
	for {
		exists, err := ExistsUserByIIdentify(ctx, cus.Identify)
		if err != nil {
			logs.ErrorContextf(ctx, "CreateUserByBindLogin: check user existence failed, %+v", err)
			return nil, err
		}
		if !exists {
			break
		}
		// 生成新的 identify，追加一位随机数
		random := random2.Number(7) // 生成 0 到 9 的随机数
		cus.Identify = fmt.Sprintf("%s%s", user.Identify, random)
	}

	// 创建用户
	if err := tx.Create(cus).Error; err != nil {
		logs.ErrorContextf(ctx, "RegisterThird: create user failed, %+v", err)
		return nil, err
	}
	return cus, nil
}

// GetUserInfo 获取用户信息
func GetUserInfo(ctx context.Context, uin uint) (userinfo *UserInfo, err error) {
	userinfo = &UserInfo{}
	err = dbutil.Account().Table(accounttype.TableNameUserIdentification+" ui").
		Select(`
			i.real_name,
			i.id_card,
			u.*,
			IF(u.password IS NOT NULL AND u.password != '',1,0)as has_password
		`).
		Joins("LEFT JOIN individual i ON ui.user_id = i.user_id").
		Joins("LEFT JOIN user u ON ui.user_id = u.id").
		Where("ui.id = ?", uin).Take(&userinfo).Error
	if err != nil {
		logs.ErrorContextf(ctx, "GetUserInfo: get user info failed, %+v", err)
		return nil, err
	}
	return userinfo, nil
}

var (
	ErrNameAlreadyExist  = errors.New("name already exist")
	ErrEmailAlreadyExist = errors.New("email already exist")
)

func UpdateUserInfo(ctx context.Context, userID, companyID, uin uint, name, avatarURL string, email *string) error {
	return dbutil.Account().Transaction(func(tx *gorm.DB) error {
		updates := make(map[string]interface{})
		if name != "" {
			updates["name"] = name
		}
		if avatarURL != "" {
			updates["avatar_url"] = avatarURL
		}

		if email != nil {
			if (*email) != "" {
				exist, err := ExistEmail(ctx, userID, *email)
				if err != nil {
					logs.ErrorContextf(ctx, "ExistEmail check failed err: %v", err)
					return err
				}
				if exist {
					return ErrEmailAlreadyExist
				}
				updates["email"] = *email
			} else {
				updates["email"] = nil
			}
		}
		if len(updates) == 0 {
			return nil
		}
		return tx.Table(accounttype.TableNameUser).
			Where("id = ?", userID).
			Updates(updates).Error
	})
}

func ExistUinName(ctx context.Context, name string, uin, companyID uint) bool {
	var result int64
	if err := dbutil.Account().Table(accounttype.TableNameUserIdentification).
		Where("deleted_at IS NULL").
		Where("id != ?", uin).
		Where("name = ?", name).
		Where("subject_type = ?", accounttype.SubjectTypeCompany).
		Where("subject_id = ?", companyID).
		Count(&result).
		Error; err != nil {
		logs.ErrorContextf(ctx, "ExistUinName: get user info failed, %+v", err)
		return true
	}
	return result > 0
}

func ExistUser(ctx context.Context, phone, email *string, companyID uint) bool {
	var c int64
	query := dbutil.Account().Table(accounttype.TableNameUser+" u ").
		Where("u.deleted_at IS NULL").
		Where("ui.subject_type = ?", accounttype.SubjectTypeCompany).
		Where("ui.subject_id = ?", companyID).
		Joins("LEFT JOIN " + accounttype.TableNameUserIdentification + " ui ON u.id = ui.user_id AND ui.deleted_at IS NULL")

	// 只有当 phone 或 email 不为 nil 时才添加对应的查询条件
	var conditions []string
	var args []any
	if phone != nil && *phone != "" {
		conditions = append(conditions, "(u.phone != '' AND u.phone = ?)")
		args = append(args, *phone)
	}
	if email != nil && *email != "" {
		conditions = append(conditions, "(u.email != '' AND u.email = ?)")
		args = append(args, *email)
	}

	if len(conditions) > 0 {
		query = query.Where("("+strings.Join(conditions, " OR ")+")", args...)
	} else {
		logs.ErrorContextf(ctx, "ExistUser: phone and email are nil or empty")
		return true
	}

	if err := query.Count(&c).Error; err != nil {
		logs.ErrorContextf(ctx, "ExistUser: get user exist faild, %+v", err)
		return true
	}
	return c > 0
}

func ExistUserAfterEdit(ctx context.Context, userName string, phone, email *string, companyID, userID uint) bool {
	var c int64
	if err := dbutil.Account().Table(accounttype.TableNameUser+" u ").
		Where("u.deleted_at IS NULL").
		Where("u.phone = ? OR u.email = ? OR ui.name = ?", phone, email, userName).
		Where("ui.subject_type = ?", accounttype.SubjectTypeCompany).
		Where("ui.subject_id = ?", companyID).
		Where("ui.user_id != ?", userID).
		Joins("LEFT JOIN " + accounttype.TableNameUserIdentification + " ui ON u.id = ui.user_id AND ui.deleted_at IS NULL").
		Count(&c).Error; err != nil {
		logs.ErrorContextf(ctx, "ExistUserAfterEdit: get user exist faild, %+v", err)
		return true
	}
	return c > 0
}

func ExistPhone(ctx context.Context, userID uint, phone string) (bool, error) {
	var c int64
	if err := dbutil.Account().Table(accounttype.TableNameUser+" u").
		Where("u.deleted_at IS NULL").
		Where("u.phone = ?", phone).
		Where("u.id != ?", userID).
		Count(&c).Error; err != nil {
		logs.ErrorContextf(ctx, "ExistPhone: (userid:%v|phone:%v) faild, %+v", userID, phone, err)
		return true, err
	}
	return c > 0, nil
}

func ExistEmail(ctx context.Context, userID uint, email string) (bool, error) {
	var c int64
	if err := dbutil.Account().Table(accounttype.TableNameUser+" u").
		Where("u.deleted_at IS NULL").
		Where("u.email = ?", email).
		Where("u.id != ?", userID).
		Count(&c).Error; err != nil {
		logs.ErrorContextf(ctx, "ExistPhone: (userid:%v|email:%v) faild, %+v", userID, email, err)
		return true, err
	}
	return c > 0, nil
}

// GetUserByPhoneAndEmail will get user by phone or email
func GetUserByPhoneAndEmail(ctx context.Context, phone, email *string) (u *accounttype.User, err error) {
	sql := dbutil.Account().WithContext(ctx).Table(accounttype.TableNameUser).
		Where("deleted_at IS NULL")
	if phone != nil && *phone != "" {
		sql.Where("phone = ?", *phone)
	}
	if email != nil && *email != "" {
		sql.Where("email = ?", *email)
	}

	if err = sql.First(&u).Error; err != nil {
		logs.WarnContextf(ctx, "GetUserByPhoneAndEmail faild err: %v", err)
		return nil, err
	}
	return u, nil
}

// GetCompanyUinByUserID will get an uin by userid and companyID
func GetCompanyUinByUserID(ctx context.Context, userID, companyID uint) (ui *accounttype.UserIdentification, err error) {
	if err = dbutil.Account().WithContext(ctx).
		Where("deleted_at IS NULL").
		Where("user_id = ?", userID).
		Where("subject_type = ?", accounttype.SubjectTypeCompany).
		Where("subject_id = ?", companyID).
		First(&ui).Error; err != nil {
		logs.WarnContextf(ctx, "GetCompanyUinByUserID faild err: %v", err)
		return nil, err
	}
	return ui, nil
}

type UinCompanyItem struct {
	accounttype.UserIdentification
	CompanyName string `json:"company_name"`
}

// GetUinCompanyByUins will get uins company list include company's name
func GetUinCompanyByUins(ctx context.Context, uins []uint) ([]*UinCompanyItem, error) {
	if len(uins) == 0 {
		logs.WarnContextf(ctx, "GetUinCompanyByUins accept len=0 uins return empty value")
		return nil, nil
	}

	items := make([]*UinCompanyItem, 0, len(uins))
	if err := dbutil.Account().Table(accounttype.TableNameUserIdentification+" u ").
		Unscoped().
		Select("u.*,c.name as company_name").
		Where("u.deleted_at IS NULL").
		Where("u.subject_type = ?", accounttype.SubjectTypeCompany).
		Where("u.id in ?", uins).
		Joins("LEFT JOIN " + accounttype.TableNameCompany + " c ON u.subject_id = c.id AND c.deleted_at IS NULL").
		Find(&items).
		Error; err != nil {
		logs.ErrorContextf(ctx, "GetUinCompanyByUins failed err: %v", err)
		return nil, err
	}
	return items, nil
}

// CanCreateCompany 判断用户是否可以创建公司
func CanCreateCompany(ctx context.Context, userID uint) bool {
	// 获取用户信息，查询公司配额
	user, err := GetUserByID(userID)
	if err != nil {
		logs.ErrorContextf(ctx, "CanCreateCompany: get user failed, userID:%d, err:%v", userID, err)
		return false
	}

	// 查询该用户已创建的公司数量（排除已删除的）
	var createdCount int64
	err = dbutil.Account().WithContext(ctx).
		Table(accounttype.TableNameCompany).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Count(&createdCount).Error
	if err != nil {
		logs.ErrorContextf(ctx, "CanCreateCompany: count created companies failed, userID:%d, err:%v", userID, err)
		return false
	}

	// 判断已创建数量是否小于配额
	return uint(createdCount) < user.CompanyQuota
}
