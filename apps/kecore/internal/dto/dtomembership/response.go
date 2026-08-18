package dtomembership

import (
	"time"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type ListPackageResponse struct {
	apiobj.BaseResponse
	Response ListPackageEmbedResponse
}

type ListPackageEmbedResponse struct {
	// List 套餐列表
	List []ListPackageItem `json:"list"`
}

type ListPackageItem struct {
	// PackageID 套餐ID
	PackageID uint `json:"package_id"`
	// Name 套餐名称
	Name string `json:"name"`
	// Description 套餐描述
	Description string `json:"description"`
	// PackageLevel 套餐等级
	PackageLevel foresttype.PackageLevel `json:"package_level"`
	// Price 套餐价格，单位分
	Price int64 `json:"price"`
	// SalePrice 套餐售价，单位分
	SalePrice int64 `json:"sale_price"`
	// AgentQuota 可创建智能体数量
	AgentQuota int64 `json:"agent_quota"`
	// QaQuota 每日问答次数
	QaQuota int64 `json:"qa_quota"`
	// DiskQuota 磁盘配额，单位字节
	// ArticleQuota 可创建文档数量
	ArticleQuota int64 `json:"article_quota"`
	DiskQuota    int64 `json:"disk_quota"`
	// EmployeeQuota 成员数量
	EmployeeQuota int64 `json:"employee_quota"`
	// Edition 套餐类型，free_trail：免费版，professional：专业版
	Edition foresttype.PackageEdition `json:"edition"`
	// AdditionalNotes 辅助说明
	AdditionalNotes []string `json:"additional_notes"`
	// IsPurchased 是否已购买
	IsPurchased bool `json:"is_purchased"`
}

type CreateOrderResponse struct {
	apiobj.BaseResponse
	Response CreateOrderEmbedResponse
}
type CreateOrderEmbedResponse struct {
	// OrderSN 订单号
	OrderSN string `json:"order_sn"`
	// 支付跳转URL
	PayURL string `json:"pay_url"`
	// 有效截止时间
	ExpireTime time.Time `json:"expire_time"`
}

type QueryOrderStatusResponse struct {
	apiobj.BaseResponse
	Response QueryOrderStatusEmbedResponse
}
type QueryOrderStatusEmbedResponse struct {
	// Status 订单状态，pending-待支付，success-已完成，closed-已关闭，cancelled-已取消
	Status string `json:"status"`
}

type ListPaymentOrderRecordResponse struct {
	apiobj.BaseResponse
	Response ListPaymentOrderRecordEmbedResponse
}

type ListPaymentOrderRecordEmbedResponse struct {
	Data []RecordItem `json:"data"`
	apiobj.QueryResponse
}

type RecordItem struct {
	ID        uint `json:"id"`
	Uin       uint `json:"uin"`
	CompanyID uint `json:"company_id"`
	//金额(分)
	Amount  int64  `json:"amount"`
	OrderSn string `json:"order_sn"`
	//流水状态：pending-处理中，success-成功，failed-失败
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	PaidAt    *time.Time `json:"paid_at"`
}
