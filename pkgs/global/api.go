package global

import "github.com/ygpkg/yg-go/apis/constants"

const (
	PrefixAPIV2 = "/v2/"
	PrefixAPIV3 = "/v3/"
)

const (
	// CtxKeyLoginStatus 登录状态
	CtxKeyLoginStatus = constants.CtxKeyLoginStatus
	// CtxKeyCompanyID 公司ID
	CtxKeyCompanyID = constants.CtxKeyCompanyID
	// CtxKeyEmployeeID 员工ID
	CtxKeyEmployeeID = constants.CtxKeyEmployeeID
	// CtxKeySubjectType 主体类型
	// CtxKeySubjectType = constants.CtxKeySubjectType
	// CtxKeyUin UIN
	CtxKeyUin        = constants.CtxKeyUin
	CtxKeyAPIKeyID   = "api_key_id"
	CtxKeyAPIKey     = "api_key"
	CtxKeyAPIKeyInfo = "api_key_info"
	CtxKeyAPIInfo    = "api_info"
	CtxKeyLang       = "lang"

	// CtxKeyClusterID 集群ID
	CtxKeyClusterID = "clusterid"
	// CtxKeyDeviceID 设备ID
	CtxKeyDeviceID = "deviceid"
)

const (
	// SubjectTypeIndividual 个人主体类型
	SubjectTypeIndividual = "individual"
	// SubjectTypeCompany 公司主体类型
	SubjectTypeCompany = "company"
)

const (
	// AudienceAdmin 对应的角色是管理员
	AudienceAdmin = "admin"
	// AudienceUser 对应的角色是普通用户
	AudienceUser = "user"
	// AudienceExternal 对应的角色是外部调用
	AudienceExternal = "external"
)

const (
	// AgentKeyXXX X9IIdTO 干啥的
	AgentKeyXXX = "syspmt_qa"
)
