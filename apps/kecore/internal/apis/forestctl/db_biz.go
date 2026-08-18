package forestctl

import (
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/plugins/dbplugins"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/types"
)

type CreateForestDBInstanceRequest struct {
	apiobj.BaseRequest
	Request CreateForestDBInstanceEmbedRequest
}

type CreateForestDBInstanceEmbedRequest struct {
	ForestDBInstanceBaseInfo
}

func (opt *CreateForestDBInstanceRequest) Validity(resp *CreateForestDBInstanceResponse) {
	if opt.Request.ForestID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_forest_id_empty" // 知识库 id 为空
		return
	}
	if opt.Request.Host == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_host_empty" // 数据源地址为空
		return
	}
	if opt.Request.Port == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_port_empty" // 数据源端口为空
		return
	}
	if opt.Request.Username == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_username_empty" // 用户名为空
		return
	}
	if opt.Request.Password == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_password_empty" // 密码为空
		return
	}
	if opt.Request.Database == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_database_empty" // 数据库名称为空
		return
	}
}

type CreateForestDBInstanceResponse struct {
	apiobj.BaseResponse
	Response CreateForestDBInstanceEmbedResponse
}

type CreateForestDBInstanceEmbedResponse struct {
	// ForestDBInstanceID 数据库实例 id
	ForestDBInstanceID uint `json:"forest_db_instance_id"`
}

type TestForestDBInstanceConnectionRequest struct {
	apiobj.BaseRequest
	Request CreateForestDBInstanceEmbedRequest
}

type TestForestDBInstanceConnectionEmbedRequest struct {
	ForestDBInstanceBaseInfo
}

func (opt *TestForestDBInstanceConnectionRequest) Validity(resp *TestForestDBInstanceConnectionResponse) {
	if opt.Request.ForestID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_forest_id_empty" // 知识库 id 为空
		return
	}
	if opt.Request.Host == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_host_empty" // 数据源地址为空
		return
	}
	if opt.Request.Port == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_port_empty" // 数据源端口为空
		return
	}
	if opt.Request.Username == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_username_empty" // 用户名为空
		return
	}
	if opt.Request.Password == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_password_empty" // 密码为空
		return
	}
	if opt.Request.Database == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_database_empty" // 数据库名称为空
		return
	}
}

type TestForestDBInstanceConnectionResponse struct {
	apiobj.BaseResponse
	Response TestForestDBInstanceConnectionEmbedResponse
}

type TestForestDBInstanceConnectionEmbedResponse struct {
	// ConnectionStatus 连接状态,success:成功,failed:失败
	ConnectionStatus foresttype.DBInstanceConnectionStatus `json:"connection_status"`
	// FailureReason 失败原因,当连接状态为失败时,返回失败原因
	FailureReason string `json:"failure_reason"`
}

type ModifyForestDBInstanceRequest struct {
	apiobj.BaseRequest
	Request ModifyForestDBInstanceEmbedRequest
}

type ModifyForestDBInstanceEmbedRequest struct {
	// ForestID 知识库 id
	ForestID uint `json:"forest_id" validate:"required"`
	ForestDBInstanceBaseInfo
}

type ForestDBInstanceBaseInfo struct {
	// ForestID 知识库 id
	ForestID uint `json:"forest_id" validate:"required"`
	// Host 数据库地址
	Host string `json:"host" validate:"required"`
	// Port 数据库端口
	Port uint `json:"port" validate:"required"`
	// Username 数据库用户名
	Username string `json:"username" validate:"required"`
	// Password 数据库密码
	Password string `json:"password" validate:"required"`
	// Database 数据库名称
	Database string `json:"database" validate:"required"`
}

func (opt *ModifyForestDBInstanceRequest) Validity(resp *ModifyForestDBInstanceResponse) {
	if opt.Request.ForestID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_forest_id_empty" // 知识库 id 为空
		return
	}
	if opt.Request.Host == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_host_empty" // 数据源地址为空
		return
	}
	if opt.Request.Port == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_port_empty" // 数据源端口为空
		return
	}
	if opt.Request.Username == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_username_empty" // 用户名为空
		return
	}
	if opt.Request.Password == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_password_empty" // 密码为空
		return
	}
	if opt.Request.Database == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_database_empty" // 数据库名称为空
		return
	}
}

type ModifyForestDBInstanceResponse struct {
	apiobj.BaseResponse
	Response ModifyForestDBInstanceEmbedResponse
}

type ModifyForestDBInstanceEmbedResponse struct {
	// ForestDBInstanceID 数据库实例 id
	ForestDBInstanceID uint `json:"forest_db_instance_id"`
}

type GetForestDBInstanceRequest struct {
	apiobj.BaseRequest
	Request GetForestDBInstanceEmbedRequest
}

func (opt *GetForestDBInstanceRequest) Validity(resp *GetForestDBInstanceResponse) {
	if opt.Request.ForestID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_forest_id_empty" // 知识库 id 为空
		return
	}
}

type GetForestDBInstanceEmbedRequest struct {
	// ForestID 知识库 id
	ForestID uint `json:"forest_id" validate:"required"`
}

type GetForestDBInstanceResponse struct {
	apiobj.BaseResponse
	Response GetForestDBInstanceEmbedResponse
}

type GetForestDBInstanceEmbedResponse struct {
	// ForestDBInstanceID 数据库实例 id
	ForestDBInstanceID uint `json:"forest_db_instance_id"`
	ForestDBInstanceBaseInfo
}

type ListForestDBRequest struct {
	apiobj.BaseRequest
	Request ListForestDBEmbedRequest
}

type ListForestDBEmbedRequest struct {
	apiobj.PageQuery
	// ForestID 知识库 id
	ForestID uint `json:"forest_id" validate:"required"`
}

func (opt *ListForestDBRequest) Validity(resp *ListForestDBResponse) {
	if opt.Request.ForestID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_forest_id_empty" // 知识库 id 为空
		return
	}
}

type ListForestDBResponse struct {
	apiobj.BaseResponse
	Response ListForestDBEmbedResponse
}

type ListForestDBEmbedResponse struct {
	apiobj.QueryResponse
	// DBList 数据库列表
	DBList []ListForestDBListItem `json:"db_list"`
}

type ListForestDBListItem struct {
	// ForestID 知识库 id
	ForestID uint `json:"forest_id"`
	// InstanceType 数据库实例类型, mysql
	InstanceType dbplugins.DatabaseType `json:"instance_type"`
	// ForestDBInstanceID 数据库实例 id
	ForestDBInstanceID uint `json:"forest_db_instance_id" validate:"required"`
	// ForestDbID 数据库 id
	ForestDbID uint `json:"forest_db_id"`
	// DNName 数据库名称
	DBName string `json:"db_name"`
	// DataSize 数据库大小, 单位: MB
	DataSize float64 `json:"data_size"`
	// DataRows 数据行数
	DataRows uint `json:"data_rows"`
	// Enable 是否启用
	Enable types.Bool `json:"enable"`
}

type ListForestTableRequest struct {
	apiobj.BaseRequest
	Request ListForestTableEmbedRequest
}

type ListForestTableEmbedRequest struct {
	// ForestID 知识库 id
	ForestID uint `json:"forest_id" validate:"required"`
	// ForestDbName 数据库名称
	ForestDbName string `json:"forest_db_name" validate:"required"`
	apiobj.PageQuery
}

func (opt *ListForestTableRequest) Validity(resp *ListForestTableResponse) {
	if opt.Request.ForestID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_forest_id_empty" // 知识库 id 为空
		return
	}
	if opt.Request.ForestDbName == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_database_empty" // 数据库名称为空
		return
	}
	for _, v := range opt.Request.Filters {
		switch v.Field {
		case "forest_table_name":
			if len(v.Value) != 1 {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "kecore_invalid_filter_value" // 查询条件中的字段只能有一个值
				return
			}
		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kecore_invalid_filter_field" // 查询条件中的字段不存在
			return
		}
	}
}

type ListForestTableResponse struct {
	apiobj.BaseResponse
	Response ListForestTableEmbedResponse
}

type ListForestTableEmbedResponse struct {
	apiobj.QueryResponse
	// TableList 数据表列表
	TableList []ListForestTableItem `json:"table_list"`
}

type ListForestTableItem struct {
	// ForestID 知识库 id
	ForestID uint `json:"forest_id"`
	// InstanceType 数据库实例类型, mysql
	InstanceType dbplugins.DatabaseType `json:"instance_type"`
	// ForestDBInstanceID 数据库实例 id
	ForestDBInstanceID uint `json:"forest_db_instance_id" validate:"required"`
	// ForestDbID 数据库 id
	ForestDbID uint `json:"forest_db_id"`
	// ForestTableName 数据表名称
	ForestTableName string `json:"forest_table_name"`
	// DataSize 数据库大小, 单位: MB
	DataSize float64 `json:"data_size"`
	// DataRows 数据行数
	DataRows uint `json:"data_rows"`
	// Enable 是否启用
	Enable types.Bool `json:"enable"`
}

type GetForestTableMetadataRequest struct {
	apiobj.BaseRequest
	Request GetForestTableMetadataEmbedRequest
}

type GetForestTableMetadataEmbedRequest struct {
	// ForestID 知识库 id
	ForestID uint `json:"forest_id" validate:"required"`
	// ForestTableName 数据表名称
	ForestTableName string `json:"forest_table_name" validate:"required"`
}

func (opt *GetForestTableMetadataRequest) Validity(resp *GetForestTableMetadataResponse) {
	if opt.Request.ForestID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_forest_id_empty" // 知识库 id 为空
		return
	}
	if opt.Request.ForestTableName == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_table_name_empty" // 数据表名称为空
		return
	}
}

type GetForestTableMetadataResponse struct {
	apiobj.BaseResponse
	Response GetForestTableMetadataEmbedResponse
}

type GetForestTableMetadataEmbedResponse struct {
	// ColumnList 字段列表
	ColumnList []GetForestTableMetadataColumnItem `json:"column_list"`
}

type GetForestTableMetadataColumnItem struct {
	// ColumnName 字段名称
	ColumnName string `json:"column_name"`
	// ColumnType 字段类型
	ColumnType string `json:"column_type"`
	// ColumnComment 字段注释
	ColumnComment string `json:"column_comment"`
}
