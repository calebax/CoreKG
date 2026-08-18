package requestyygu

import (
	"context"
	"time"

	"github.com/ygpkg/yg-go/logs"
)

const (
	listUinPath           = "/v2/account.ListUin"
	getDepartmentTreePath = "/v2/account.GetDepartmentTree"
	getCompanyInfoPath    = "/v2/account.GetCompanyInfo"
)

// ListUin 拉取当前用户在 ROC 中的全部身份（空间/公司）
func ListUin(ctx context.Context) (*ListUinResponse, error) {
	resp := &ListUinResponse{}
	if err := YyguRequest(ctx, listUinPath, map[string]interface{}{}, resp); err != nil {
		logs.ErrorContextf(ctx, "failed to list uin from yygu: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetDepartmentTree 获取空间/部门树以及成员信息
func GetDepartmentTree(ctx context.Context, includeEmployee bool) (*GetDepartmentTreeResponse, error) {
	resp := &GetDepartmentTreeResponse{}
	if err := YyguRequest(ctx, getDepartmentTreePath, map[string]interface{}{
		"include_employee": includeEmployee,
	}, resp); err != nil {
		logs.ErrorContextf(ctx, "failed to get department tree from yygu: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetCompanyInfo 获取公司/空间信息
func GetCompanyInfo(ctx context.Context) (*GetCompanyInfoResponse, error) {
	resp := &GetCompanyInfoResponse{}
	if err := YyguRequest(ctx, getCompanyInfoPath, map[string]interface{}{}, resp); err != nil {
		logs.ErrorContextf(ctx, "failed to get company info from yygu: %v", err)
		return nil, err
	}
	return resp, nil
}

// ListUinResponse 对应 account.ListUin 的响应体
type ListUinResponse struct {
	Uin []LoginUin `json:"uin"`
}

// LoginUin 与 ROC 侧返回结构对齐
type LoginUin struct {
	Uin           UserIdentification `json:"uin"`
	Name          string             `json:"name,omitempty"`
	CompanyName   string             `json:"company_name,omitempty"`
	CompanyLogo   string             `json:"company_logo,omitempty"`
	Role          string             `json:"role,omitempty"`
	CompanyStatus string             `json:"company_status,omitempty"`
}

// UserIdentification 对齐 ROC accounttype.UserIdentification
type UserIdentification struct {
	ID          uint       `json:"ID"`
	CreatedAt   time.Time  `json:"CreatedAt"`
	UpdatedAt   time.Time  `json:"UpdatedAt"`
	DeletedAt   *time.Time `json:"DeletedAt,omitempty"`
	UserID      uint       `json:"UserID"`
	SubjectType string     `json:"SubjectType"`
	SubjectID   uint       `json:"SubjectID"`
	UinStatus   string     `json:"UinStatus"`
	Issuer      string     `json:"Issuer"`
	Name        string     `json:"Name"`
	LastLoginAt *time.Time `json:"LastLoginAt"`
}

// GetDepartmentTreeResponse 对应 account.GetDepartmentTree 的响应体
type GetDepartmentTreeResponse struct {
	Departments []Department   `json:"departments"`
	Employees   []EmployeeInfo `json:"employees,omitempty"`
}

// GetCompanyInfoResponse 对应 account.GetCompanyInfo 的响应体
type GetCompanyInfoResponse struct {
	ID          uint   `json:"ID"`
	Name        string `json:"name"`
	Alias       string `json:"alias"`
	Description string `json:"description"`
	Logo        string `json:"logo"`
}

// Department 对齐 ROC accounttype.AccountDepartment
type Department struct {
	ID        uint       `json:"ID"`
	CreatedAt time.Time  `json:"CreatedAt"`
	UpdatedAt time.Time  `json:"UpdatedAt"`
	DeletedAt *time.Time `json:"DeletedAt,omitempty"`
	Name      string     `json:"Name"`
	ParentID  uint       `json:"ParentID"`
	Sort      uint       `json:"Sort"`
	CompanyID uint       `json:"CompanyID"`
}

// EmployeeInfo 对齐 ROC dtoorganize.EmployeeInfo
type EmployeeInfo struct {
	Uin           uint      `json:"uin"`
	CreatedAt     time.Time `json:"created_at"`
	UserName      string    `json:"user_name"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	Phone         string    `json:"phone"`
	EmployeeID    uint      `json:"employee_id"`
	Role          string    `json:"role"`
	DepartmentIDs []uint    `json:"department_ids"`
}
