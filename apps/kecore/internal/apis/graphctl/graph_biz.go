package graphctl

import (
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/graph"
	"github.com/insmtx/corekg/apps/kecore/models/nebulagraph"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/types"
)

// CreateGraphRequest 创建图谱请求
type CreateGraphRequest struct {
	apiobj.BaseRequest
	Request struct {
		Name        string                 `json:"name"`
		Description string                 `json:"description"`
		PublicScope foresttype.PublicScope `json:"public_scope"`
		ScopeIDs    types.UintArray        `json:"scope_ids"`
		ManagerIDs  types.UintArray        `json:"manager_ids"`
		AvatarUrl   string                 `json:"avatar_url"`
	}
}

// Validity 校验创建图谱请求
func (opt *CreateGraphRequest) Validity(resp *CreateGraphResponse) {
	if opt.Request.Name == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_forest_name_empty" // 知识图谱名称不能为空
		return
	}
}

// CreateGraphResponse 创建图谱响应
type CreateGraphResponse struct {
	apiobj.BaseResponse
	Response struct {
		*foresttype.ForestGraphInfo
	}
}

// UpdateGraphRequest 更新图谱请求
type UpdateGraphRequest struct {
	apiobj.BaseRequest
	Request struct {
		GraphID     uint                   `json:"graph_id"`
		Name        string                 `json:"name"`
		Description string                 `json:"description"`
		FileIDList  []uint                 `json:"file_id_list"`
		ParseMode   foresttype.ParseMode   `json:"parse_mode"`
		PublicScope foresttype.PublicScope `json:"public_scope"`
		ScopeIDs    types.UintArray        `json:"scope_ids"`
		ManagerIDs  types.UintArray        `json:"manager_ids"`
	}
}

// Validity 校验更新图谱请求
func (req *UpdateGraphRequest) Validity(resp *UpdateGraphResponse) {
	if req.Request.GraphID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_forest_id" // 请选择图谱
		return
	}
	switch req.Request.ParseMode {
	case foresttype.ParseModeAuto:
		req.Request.ParseMode = foresttype.ParseModeAuto
	case foresttype.ParseModeRule:
		req.Request.ParseMode = foresttype.ParseModeRule
	case "":
		req.Request.ParseMode = foresttype.ParseModeAuto
	default:
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_parse_mode" // 请选择正确的模式
		return
	}
}

// UpdateGraphResponse 更新图谱响应
type UpdateGraphResponse struct {
	apiobj.BaseResponse
}

// CreateTagRequest 创建标签请求
type CreateTagRequest struct {
	apiobj.BaseRequest
	Request struct {
		GraphID     uint                  `json:"graph_id"`
		TagName     string                `json:"tag_name"`
		Description string                `json:"description"`
		Properties  foresttype.Properties `json:"properties"`
	}
}

// Validity 校验创建标签请求
func (req *CreateTagRequest) Validity(resp *CreateTagResponse) {
	if req.Request.GraphID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_forest_id" // 请选择图谱
		return
	}
	if graph.ReplaceString(req.Request.TagName) == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_tag_name_empty" // 类型名不能为空或包含特殊字符
		return
	}
	req.Request.TagName = graph.ReplaceString(req.Request.TagName)
	for i, v := range req.Request.Properties {
		req.Request.Properties[i].Comment = graph.ReplaceString(v.Comment)
		req.Request.Properties[i].Name = graph.ReplaceString(v.Name)
		if req.Request.Properties[i].Type == "string" && req.Request.Properties[i].Defaults != nil {
			req.Request.Properties[i].Defaults = graph.ReplaceString(req.Request.Properties[i].Defaults.(string))
		}
		if req.Request.Properties[i].Name == "" {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kecore_property_name_empty" // 属性名不能为空
			return
		}
	}

	err := req.Request.Properties.ValidateProperties()
	if err != nil {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = err.Error()
		return
	}
}

// CreateTagResponse 创建标签响应
type CreateTagResponse struct {
	apiobj.BaseResponse
	Response struct {
		*foresttype.GraphTag
	}
}

// UpdateTagRequest 更新标签请求
type UpdateTagRequest struct {
	apiobj.BaseRequest
	Request struct {
		GraphID     uint                  `json:"graph_id"`
		TagID       uint                  `json:"tag_id"`
		TagName     string                `json:"tag_name"`
		Description string                `json:"description"`
		Properties  foresttype.Properties `json:"properties"`
	}
}

// Validity 校验更新标签请求
func (req *UpdateTagRequest) Validity(resp *UpdateTagResponse) {
	if req.Request.TagID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_tag_id" // 请选择实体
		return
	}
	if req.Request.GraphID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_forest_id" // 请选择图谱
		return
	}
	if graph.ReplaceString(req.Request.TagName) == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_tag_name_empty" // 类型名不能为空
		return
	}
	req.Request.TagName = graph.ReplaceString(req.Request.TagName)
	for i, v := range req.Request.Properties {
		req.Request.Properties[i].Comment = graph.ReplaceString(v.Comment)
		req.Request.Properties[i].Name = graph.ReplaceString(v.Name)
		if req.Request.Properties[i].Type == "string" && req.Request.Properties[i].Defaults != nil {
			req.Request.Properties[i].Defaults = graph.ReplaceString(req.Request.Properties[i].Defaults.(string))
		}
		if req.Request.Properties[i].Name == "" {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kecore_property_name_empty" // 属性名不能为空
			return
		}
	}
	err := req.Request.Properties.ValidateProperties()
	if err != nil {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = err.Error()
		return
	}
}

// UpdateTagResponse 更新标签响应
type UpdateTagResponse struct {
	apiobj.BaseResponse
	Response struct {
		*foresttype.GraphTag
	}
}

// DeleteTagRequest 删除标签请求
type DeleteTagRequest struct {
	apiobj.BaseRequest
	Request struct {
		GraphID uint `json:"graph_id"`
		TagID   uint `json:"tag_id"`
	}
}

// Validity 校验删除标签请求
func (req *DeleteTagRequest) Validity(resp *DeleteTagResponse) {
	if req.Request.TagID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_tag_id" // 请选择实体
		return
	}
	if req.Request.GraphID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_forest_id" // 请选择图谱
		return
	}
}

// DeleteTagResponse 删除标签响应
type DeleteTagResponse struct {
	apiobj.BaseResponse
	Response struct {
		*foresttype.GraphTag
	}
}

// CreateEdgeRequest 创建边请求
type CreateEdgeRequest struct {
	apiobj.BaseRequest
	Request struct {
		GraphID  uint   `json:"graph_id"`
		EdgeName string `json:"edge_name"`
		SrcTagID uint   `json:"src_tag_id"`
		DstTagID uint   `json:"dst_tag_id"`
	}
}

// Validity 校验创建边请求
func (req *CreateEdgeRequest) Validity(resp *CreateEdgeResponse) {
	if req.Request.GraphID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_forest_id" // 请选择图谱
		return
	}
	if req.Request.SrcTagID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_src_tag" // 请选择起点属性
		return
	}
	if req.Request.DstTagID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_dst_tag" // 请选择终点属性
		return
	}
	if graph.ReplaceString(req.Request.EdgeName) == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_enter_edge_name" // 请输入边名
		return
	}
	req.Request.EdgeName = graph.ReplaceString(req.Request.EdgeName)
}

// CreateEdgeResponse 创建边响应
type CreateEdgeResponse struct {
	apiobj.BaseResponse
	Resopnse struct {
		*foresttype.GraphEdgeTag
	}
}

// DeleteEdgeRequest 删除边请求
type DeleteEdgeRequest struct {
	apiobj.BaseRequest
	Request struct {
		GraphID uint `json:"graph_id"`
		EgdeID  uint `json:"edge_id"`
	}
}

// Validity 校验删除边请求
func (req *DeleteEdgeRequest) Validity(resp *DeleteEdgeResponse) {
	if req.Request.EgdeID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_edge_id" // 请选择边
		return
	}
	if req.Request.GraphID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_forest_id" // 请选择图谱
		return
	}
}

// DeleteEdgeResponse 删除边响应
type DeleteEdgeResponse struct {
	apiobj.BaseResponse
	Response struct {
		*foresttype.GraphTag
	}
}

// UpdateEdgeRequest 更新边请求
type UpdateEdgeRequest struct {
	apiobj.BaseRequest
	Request struct {
		GraphID  uint `json:"graph_id"`
		EdgeID   uint `json:"edge_id"`
		SrcTagID uint `json:"src_tag_id"`
		DstTagID uint `json:"dst_tag_id"`
	}
}

// Validity 校验更新边请求
func (req *UpdateEdgeRequest) Validity(resp *UpdateEdgeResponse) {
	if req.Request.GraphID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_forest_id" // 请选择图谱
		return
	}
	if req.Request.SrcTagID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_src_tag" // 请选择起点属性
		return
	}
	if req.Request.DstTagID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_dst_tag" // 请选择终点属性
		return
	}
}

// UpdateEdgeResponse 更新边响应
type UpdateEdgeResponse struct {
	apiobj.BaseResponse
	Resopnse struct {
		*foresttype.GraphEdgeTag
	}
}

// ListForestGraphRequest 获取知识森林图谱请求
type ListForestGraphRequest struct {
	apiobj.BaseRequest
	Request apiobj.PageQuery
}

// Validity 校验获取知识森林图谱请求
func (req *ListForestGraphRequest) Validity(resp *ListForestGraphResponse) {
	if req.Request.Offset < 0 || req.Request.Limit < 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_offset_limit_invalid" // offset和limit必须大于0
		return
	}
	for _, v := range req.Request.OrderBy {
		switch v {
		case "created_at", "updated_at", "name":
		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kecore_order_by_empty" // orderBy不能为空
			return
		}
	}
	for _, v := range req.Request.Filters {
		switch v.Field {
		case "name", "created_at", "updated_at":
			if len(v.Value) != 1 {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "kecore_filter_field_single_value" // 查询条件中的字段只能有一个值
				return
			}
			if v.Value[0] == "" {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "kecore_filter_field_empty_value" // 查询条件中的值不能为空
				return
			}
		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kecore_invalid_filter_field_data" // 查询条件中的字段不存在
			return
		}
	}
}

// ListForestGraphResponse 获取知识森林图谱响应
type ListForestGraphResponse struct {
	apiobj.BaseResponse
	Response *graph.ForestInfoItemList
}

// GetGraphInfoRequest 获取图谱信息请求
type GetGraphInfoRequest struct {
	apiobj.BaseRequest
	Request struct {
		GraphID uint `json:"graph_id"`
	}
}

// Validity 校验获取图谱信息请求
func (req *GetGraphInfoRequest) Validity(resp *GetGraphInfoResponse) {
	if req.Request.GraphID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_forest_id" // 请选择图谱
		return
	}
}

// GetGraphInfoResponse 获取图谱信息响应
type GetGraphInfoResponse struct {
	apiobj.BaseResponse
	Response struct {
		*foresttype.ForestGraphInfo
		ScopeIDs   types.UintArray `json:"scope_ids"`
		ManagerIDs types.UintArray `json:"manager_ids"`
		IsAdmin    bool            `json:"is_admin"`
	}
}

// DeleteGraphRequest 删除图谱请求
type DeleteGraphRequest struct {
	apiobj.BaseRequest
	Request struct {
		GraphID uint `json:"graph_id"`
	}
}

// Validity 校验删除图谱请求
func (req *DeleteGraphRequest) Validity(resp *DeleteGraphResponse) {
	if req.Request.GraphID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_forest_id" // 请选择图谱
		return
	}
}

// DeleteGraphResponse 删除图谱响应
type DeleteGraphResponse struct {
	apiobj.BaseResponse
	Response struct{}
}

// ListGraphTagRequest 获取图谱标签请求
type ListGraphTagRequest struct {
	apiobj.BaseRequest
	Request struct {
		apiobj.PageQuery
		GraphID uint `json:"graph_id"`
	}
}

// Validity 校验获取图谱标签请求
func (req *ListGraphTagRequest) Validity(resp *ListGraphTagResponse) {
	if req.Request.GraphID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_forest_id" // 请选择图谱
		return
	}
	if req.Request.Offset < 0 || req.Request.Limit < 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_offset_limit_invalid" // offset和limit必须大于0
		return
	}
	for _, v := range req.Request.OrderBy {
		switch v {
		case "created_at", "updated_at", "name":
		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kecore_order_by_empty" // orderBy不能为空
			return
		}
	}
	for _, v := range req.Request.Filters {
		switch v.Field {
		case "name", "tag_type", "created_at", "updated_at":
			if len(v.Value) != 1 {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "kecore_filter_field_single_value" // 查询条件中的字段只能有一个值
				return
			}
			if v.Value[0] == "" {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "kecore_filter_field_empty_value" // 查询条件中的值不能为空
				return
			}
		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kecore_invalid_filter_field_data" // 查询条件中的字段不存在
			return
		}
	}
}

// ListGraphTagResponse 获取图谱标签响应
type ListGraphTagResponse struct {
	apiobj.BaseResponse
	Response *graph.TagInfoList
}

// ListGraphNodeRequest 获取图谱节点请求
type ListGraphNodeRequest struct {
	apiobj.BaseRequest
	Request ListGraphNodeEmbedRequest
}

// ListGraphNodeEmbedRequest 嵌套请求
type ListGraphNodeEmbedRequest struct {
	// GraphID 图谱ID
	GraphID uint `json:"graph_id" validate:"required"`
	// GraphTagID 图谱标签ID
	GraphTagID uint `json:"graph_tag_id"`
	// GraphNodeName 图谱节点名称
	GraphNodeName string `json:"graph_node_name"`
}

// Validity 校验获取图谱节点请求
func (req *ListGraphNodeRequest) Validity(resp *ListGraphNodeResponse) {
	if req.Request.GraphID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_forest_id" // 请选择图谱
		return
	}
}

// ListGraphNodeResponse 获取图谱节点响应
type ListGraphNodeResponse struct {
	apiobj.BaseResponse
	Response ListGraphNodeEmbedResponse
}

// ListGraphNodeEmbedResponse 嵌套响应
type ListGraphNodeEmbedResponse struct {
	// List 图谱节点列表
	List []ListGraphNodeItem `json:"list"`
	// Total 总数
	Total int64 `json:"total"`
}

// ListGraphNodeItem 图谱节点项
type ListGraphNodeItem struct {
	// GraphNodeID 图谱节点ID
	GraphNodeID uint `json:"graph_node_id"`
	// GraphNodeName 图谱节点名称
	GraphNodeName string `json:"graph_node_name"`
}

// GetKnowledgeGraphRequest 获取知识图谱请求
type GetKnowledgeGraphRequest struct {
	apiobj.BaseRequest
	Request struct {
		GraphID uint `json:"graph_id"`
		nebulagraph.KnowledgeGraphReq
	}
}

// Validity 校验获取知识图谱请求
func (req *GetKnowledgeGraphRequest) Validity(resp *GetKnowledgeGraphResponse) {
	if req.Request.GraphID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_forest_id" // 请选择图谱
		return
	}
	if req.Request.Limit == 0 {
		req.Request.Limit = 200
	}
}

// GetKnowledgeGraphResponse 获取知识图谱响应
type GetKnowledgeGraphResponse struct {
	apiobj.BaseResponse
	Response struct {
		KnowledgeGraph *nebulagraph.Graph `json:"knowledge_graph"`
	}
}

// GetTagEdgeRequest 获取标签边请求
type GetTagEdgeRequest struct {
	apiobj.BaseRequest
	Request struct {
		GraphID uint `json:"graph_id"`
		TagID   uint `json:"tag_id"`
	}
}

// Validity 校验获取标签边请求
func (req *GetTagEdgeRequest) Validity(resp *GetTagEdgeResponse) {
	if req.Request.GraphID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_forest_id" // 请选择图谱
		return
	}
	if req.Request.TagID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_tag_id" // 请选择实体
		return
	}
}

// GetTagEdgeResponse 获取标签边响应
type GetTagEdgeResponse struct {
	apiobj.BaseResponse
	Response struct {
		Data []graph.EdgeTagInfo `json:"data"`
	}
}

// SubmitTemplateRequest 提交模板请求
type SubmitTemplateRequest struct {
	apiobj.BaseRequest
	Request struct {
		GraphID  uint            `json:"graph_id"`
		Template *graph.Template `json:"template"`
	}
}

// Validity 校验提交模板请求
func (req *SubmitTemplateRequest) Validity(resp *SubmitTemplateResponse) {
	if req.Request.GraphID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_forest_id" // 请选择图谱
		return
	}
}

// SubmitTemplateResponse 提交模板响应
type SubmitTemplateResponse struct {
	apiobj.BaseResponse
	Response struct {
	}
}
