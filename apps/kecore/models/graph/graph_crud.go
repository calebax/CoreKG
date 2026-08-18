package graph

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/nebulagraph"
	"github.com/insmtx/corekg/apps/kecore/models/perm"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/random"
	"gorm.io/gorm"
)

const (
	graph_prefix = "ke_graph_"
)

// CreateGraph creates a new graph
func CreateGraph(ctx context.Context, graph *foresttype.ForestGraphInfo, tx *gorm.DB) error {
	if graph.SpaceName == "" {
		graph.SpaceName = graph_prefix + random.String(20)
	}
	cli, err := nebulagraph.NewNebulaCLI(ctx, "")
	if err != nil {
		logs.ErrorContextf(ctx, "NewNebulaCLI error: %v", err)
		return err
	}
	defer cli.Release()
	err = cli.CreateSpace(graph.SpaceName)
	if err != nil {
		logs.ErrorContextf(ctx, "CreateSpace error: %v", err)
		return err
	}
	// 创建图谱
	g := &foresttype.ForestGraph{
		Uin:         graph.Uin,
		CompanyID:   graph.CompanyID,
		Name:        graph.Name,
		Description: graph.Description,
		PublicScope: graph.PublicScope,
		ForestID:    graph.ForestID,
		AvatarUrl:   graph.AvatarUrl,
	}
	err = tx.WithContext(ctx).Create(g).Error
	if err != nil {
		logs.ErrorContextf(ctx, "CreateGraph create graph error: %v", err)
		return err
	}
	v := &foresttype.ForestGraphVersion{
		Uin:       graph.Uin,
		CompanyID: graph.CompanyID,
		GraphID:   g.ID,
		SpaceName: graph.SpaceName,
	}
	err = tx.WithContext(ctx).Create(v).Error
	if err != nil {
		logs.ErrorContextf(ctx, "CreateGraph create version error: %v", err)
		return err
	}

	g.VersionID = v.ID
	err = tx.WithContext(ctx).Save(g).Error
	if err != nil {
		logs.ErrorContextf(ctx, "CreateGraph update graph error: %v", err)
		return err
	}
	graph.ID = g.ID
	graph.CreatedAt = g.CreatedAt
	graph.UpdatedAt = g.UpdatedAt
	graph.Uin = g.Uin
	graph.CompanyID = g.CompanyID
	graph.Name = g.Name
	graph.Description = g.Description
	graph.ParseMode = v.ParseMode
	graph.PublicScope = g.PublicScope
	graph.ForestID = g.ForestID
	graph.VersionID = v.ID
	graph.Status = v.Status
	graph.SpaceName = v.SpaceName
	graph.FileIDList = v.FileIDList
	graph.AvatarUrl = g.AvatarUrl
	return nil
}

// CreateGraphVersion 创建图谱新版本
func CreateGraphVersion(ctx context.Context, graph *foresttype.ForestGraphInfo, tx *gorm.DB) error {
	if graph.ID == 0 {
		return fmt.Errorf("CreateGraphVersion: graph ID is zero")
	}
	graph.SpaceName = graph_prefix + random.String(20)
	cli, err := nebulagraph.NewNebulaCLI(ctx, "")
	if err != nil {
		logs.ErrorContextf(ctx, "NewNebulaCLI error: %v", err)
		return err
	}
	defer cli.Release()
	err = cli.CreateSpace(graph.SpaceName)
	if err != nil {
		logs.ErrorContextf(ctx, "CreateSpace error: %v", err)
		return err
	}

	v := &foresttype.ForestGraphVersion{
		Uin:       graph.Uin,
		CompanyID: graph.CompanyID,
		GraphID:   graph.ID,
		SpaceName: graph.SpaceName,
	}
	err = tx.WithContext(ctx).Create(v).Error
	if err != nil {
		logs.ErrorContextf(ctx, "CreateGraph create version error: %v", err)
		return err
	}

	graph1Update := foresttype.ForestGraph{
		VersionID: v.ID,
	}
	err = tx.WithContext(ctx).Model(&foresttype.ForestGraph{}).Where("id = ?", graph.ID).Updates(&graph1Update).Error
	if err != nil {
		logs.ErrorContextf(ctx, "CreateGraphVersion: failed to update ke_forest_graph table, error: %v", err)
		return err
	}
	// 创建version
	graph.VersionID = v.ID
	graph.Status = v.Status
	graph.SpaceName = v.SpaceName
	graph.FileIDList = v.FileIDList
	return nil
}

// GetGraph 获取图 ,最新的一个版本
func GetGraph(ctx context.Context, graphID uint) (*foresttype.ForestGraphInfo, error) {
	var graph foresttype.ForestGraphInfo
	err := dbutil.Knownow().WithContext(ctx).
		Table(foresttype.ForestGraph{}.TableName()+" AS g").
		Select("g.*, v.status, v.file_id_list, v.space_name, v.parse_mode").
		Joins("LEFT JOIN "+foresttype.ForestGraphVersion{}.TableName()+" AS v ON g.version_id = v.id").
		Where("g.id = ?", graphID).First(&graph).Error
	if err != nil {
		logs.ErrorContextf(ctx, "GetGraph error: %v", err)
		return nil, err
	}
	return &graph, nil
}

// GetForestGraph 获取知识库图 ,最新的一个版本
func GetForestGraph(ctx context.Context, forestID uint) (*foresttype.ForestGraphInfo, error) {
	var graph foresttype.ForestGraphInfo
	err := dbutil.Knownow().WithContext(ctx).
		Table(foresttype.ForestGraph{}.TableName()+" AS g").
		Select("g.*, v.status, v.file_id_list, v.space_name, v.parse_mode").
		Joins("LEFT JOIN "+foresttype.ForestGraphVersion{}.TableName()+" AS v ON g.version_id = v.id").
		Where("g.forest_id = ?", forestID).First(&graph).Error
	if err != nil {
		logs.WarnContextf(ctx, "GetForestGraph error: %v", err)
		return nil, err
	}
	return &graph, nil
}

// GetGraphWithVersionID 根据版本id获取图谱信息
func GetGraphWithVersionID(ctx context.Context, versionID uint) (*foresttype.ForestGraphInfo, error) {
	var graph foresttype.ForestGraphInfo
	err := dbutil.Knownow().WithContext(ctx).
		Table(foresttype.ForestGraphVersion{}.TableName()+" AS v").
		Select("g.*, v.status, v.file_id_list, v.space_name, v.parse_mode").
		Joins("LEFT JOIN "+foresttype.ForestGraph{}.TableName()+" AS g ON v.graph_id = g.id").
		Where("v.id = ?", versionID).First(&graph).Error
	if err != nil {
		logs.ErrorContextf(ctx, "GetGraphWithVersionID error: %v", err)
		return nil, err
	}
	return &graph, nil
}

// GetPreviousVersion 获取指定图谱的上一个版本
func GetPreviousVersion(ctx context.Context, graphID, currentVersionID uint) (*foresttype.ForestGraphInfo, error) {
	var graph foresttype.ForestGraphInfo
	err := dbutil.Knownow().WithContext(ctx).
		Table(foresttype.ForestGraphVersion{}.TableName()+" AS v").
		Select("g.id, g.created_at, g.updated_at,g.deleted_at, g.uin, g.company_id, g.name, g.description, g.public_scope, g.forest_id, g.avatar_url, "+
			"v.id as version_id, v.status, v.file_id_list, v.space_name, v.parse_mode").
		Joins("LEFT JOIN "+foresttype.ForestGraph{}.TableName()+" AS g ON v.graph_id = g.id").
		Where("v.graph_id = ?", graphID).
		Where("v.id < ?", currentVersionID).
		Order("v.id DESC").
		Limit(1).
		First(&graph).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logs.WarnContextf(ctx, "GetPreviousVersion: no previous version found for graph %d", graphID)
			return nil, err
		}
		logs.ErrorContextf(ctx, "GetPreviousVersion error: %v", err)
		return nil, err
	}
	return &graph, nil
}

// DeleteGraph 删除图谱
func DeleteGraph(ctx context.Context, graphID uint) error {
	err := dbutil.Knownow().WithContext(ctx).Where("id = ?", graphID).
		Delete(&foresttype.ForestGraph{}).Error
	if err != nil {
		logs.ErrorContextf(ctx, "GetGraph error: %v", err)
		return err
	}
	return nil
}

// DeleteGraph 删除图谱
func DeleteGraphTX(ctx context.Context, tx *gorm.DB, graphID uint) error {
	if graphID == 0 {
		return nil
	}
	err := tx.WithContext(ctx).Where("id = ?", graphID).
		Delete(&foresttype.ForestGraph{}).Error
	if err != nil {
		logs.ErrorContextf(ctx, "GetGraph error: %v", err)
		return err
	}
	return nil
}

// UpdateGraph 更新图谱信息。此函数会分别更新 ke_forest_graph 表和 ke_forest_graph_version 表。
func UpdateGraph(ctx context.Context, graph *foresttype.ForestGraphInfo, tx *gorm.DB) error {
	if tx == nil {
		return fmt.Errorf("UpdateGraph: transaction is nil")
	}

	db := tx.WithContext(ctx)

	graph1Update := foresttype.ForestGraph{
		Name:        graph.Name,
		Description: graph.Description,
		PublicScope: graph.PublicScope,
		ForestID:    graph.ForestID,
		VersionID:   graph.VersionID,
		Uin:         graph.Uin,
		CompanyID:   graph.CompanyID,
	}
	err := db.Model(&foresttype.ForestGraph{}).Where("id = ?", graph.ID).Updates(&graph1Update).Error
	if err != nil {
		logs.ErrorContextf(ctx, "UpdateGraph: failed to update ke_forest_graph table, error: %v", err)
		return err
	}

	graphVersionUpdate := foresttype.ForestGraphVersion{
		Status:      graph.Status,
		Description: graph.Description,
		GraphID:     graph.ID,
		SpaceName:   graph.SpaceName,
		FileIDList:  graph.FileIDList,
		ParseMode:   graph.ParseMode,
		Uin:         graph.Uin,
		CompanyID:   graph.CompanyID,
	}
	// 注意：这里使用 graph.ID 作为 ke_forest_graph_version 表的 graph_id
	err = db.Model(&foresttype.ForestGraphVersion{}).Where("id = ?", graph.VersionID).Updates(&graphVersionUpdate).Error
	if err != nil {
		logs.ErrorContextf(ctx, "UpdateGraph: failed to update ke_forest_graph_version table, error: %v", err)
		return err
	}

	return nil
}

// UpdateGraphStatus 修改图谱状态
func UpdateGraphStatus(ctx context.Context, graphID uint, status foresttype.GraphStatus) error {
	graphinfo, err := GetGraph(ctx, graphID)
	if err != nil {
		logs.ErrorContextf(ctx, "UpdateGraphStatus GetGraph error: %v", err)
		return err
	}

	// 更新版本状态
	err = dbutil.Knownow().WithContext(ctx).
		Table(foresttype.ForestGraphVersion{}.TableName()).
		Where("id = ?", graphinfo.VersionID).
		Update("status", status).Error

	if err != nil {
		logs.ErrorContextf(ctx, "UpdateGraphStatus update error: %v", err)
		return err
	}

	return nil
}

// UpdateGraphStatus 修改图谱状态
func UpdateGraphStatusWithForestID(ctx context.Context, forestID uint, status foresttype.GraphStatus) error {
	graphinfo, err := GetForestGraph(ctx, forestID)
	if err != nil {
		logs.ErrorContextf(ctx, "UpdateGraphStatus GetGraph error: %v", err)
		return err
	}

	// 更新版本状态
	err = dbutil.Knownow().WithContext(ctx).
		Table(foresttype.ForestGraphVersion{}.TableName()).
		Where("id = ?", graphinfo.VersionID).
		Update("status", status).Error

	if err != nil {
		logs.ErrorContextf(ctx, "UpdateGraphStatus update error: %v", err)
		return err
	}

	return nil
}

// ListForestGraph 获取所有图谱列表
func ListForestGraph(ctx context.Context, opt apiobj.PageQuery) (*ForestInfoItemList, error) {
	var queryList = &ForestInfoItemList{}
	sql := dbutil.Knownow().WithContext(ctx).
		Table(foresttype.ForestGraph{}.TableName()+" AS g").
		Select(`g.*, v.status, v.file_id_list, v.space_name, v.parse_mode,
            (SELECT COUNT(*) FROM ke_graph_tag_node WHERE graph_id = g.id AND graph_version_id = g.version_id AND deleted_at IS NULL) as node_count,
            (SELECT COUNT(*) FROM ke_graph_edge WHERE graph_id = g.id AND graph_version_id = g.version_id AND deleted_at IS NULL) as edge_count`).
		Joins("LEFT JOIN "+foresttype.ForestGraphVersion{}.TableName()+" AS v ON g.version_id = v.id").
		Where("g.deleted_at IS NULL").
		Where("g.company_id = ?", opt.CompanyID)

	for _, filter := range opt.Filters {
		switch filter.Field {
		case "uin":
			sql = sql.Where("g.uin = ?", filter.Value[0])
		case "name":
			sql = sql.Where("g.name LIKE ?", fmt.Sprintf("%%%s%%", filter.Value[0]))
		default:
			logs.ErrorContextf(ctx, "ListForestGraph invalid filtering field: %s", filter.Field)
			return nil, fmt.Errorf("invalid filtering field: %s", filter.Field)
		}
	}

	var authIDs []uint
	if err := dbutil.Knownow().WithContext(ctx).Table(foresttype.TableNameKeResourceScope).
		Where("resource_type = ? AND deleted_at IS NULL", foresttype.ResourceTypeGraph).
		Where("("+
			// 1. 公开权限 (action = 'view' 且 scope_type = 'public')
			"(action = ? AND scope_type = ?) OR "+
			// 2. 个人管理权限 (action = 'manage' 且 scope_type = 'user' 且 scope_id = 当前用户)
			"(action = ? AND scope_type = ? AND scope_id = ?) OR "+
			// 3. 个人查看权限 (action = 'view' 且 scope_type = 'user' 且 scope_id = 当前用户)
			"(action = ? AND scope_type = ? AND scope_id = ?) OR "+
			// 4. 公司权限 (action = 'view' 且 scope_type = 'company' 且 scope_id = 当前公司)
			"(action = ? AND scope_type = ? AND scope_id = ?)"+
			")",
			foresttype.ActionView, foresttype.ScopeTypePublic, // 公开
			foresttype.ActionManage, foresttype.ScopeTypeUser, opt.Uin, // 个人管理
			foresttype.ActionView, foresttype.ScopeTypeUser, opt.Uin, // 个人查看
			foresttype.ActionView, foresttype.ScopeTypeCompany, opt.CompanyID). // 公司权限
		Pluck("distinct resource_id", &authIDs).
		Error; err != nil {
		return nil, err
	}

	otherAuth := dbutil.Knownow().WithContext(ctx).Table(foresttype.TableNameKeForestGraph+" g ").
		Joins("LEFT JOIN "+foresttype.TableNameKeForestGraphVersion+" v ON g.version_id = v.id").
		Select("g.id").
		Where("g.deleted_at IS NULL").
		Where("v.status = ? AND g.uin = ?", foresttype.GraphStatusDraft, opt.Uin)

	sql = sql.Where("("+
		//有权限查看
		"g.id IN (?) OR "+
		//草稿
		"g.id IN (?)"+
		")",
		authIDs,
		otherAuth)

	// 处理 BeginTime 和 EndTime
	if !opt.BeginTime.IsZero() {
		sql = sql.Where("g.created_at >= ?", opt.BeginTime)
	}
	if !opt.EndTime.IsZero() {
		sql = sql.Where("g.created_at <= ?", opt.EndTime)
	}

	if err := sql.Count(&queryList.Total).Error; err != nil {
		logs.ErrorContextf(ctx, "ListForestGraph Statistical project failed: %v", err)
		return nil, err
	}
	if queryList.Total == 0 {
		return queryList, nil
	}
	if len(opt.OrderBy) > 0 {
		sql = sql.Order(strings.Join(opt.OrderBy, ","))
	}
	sql = sql.Offset(opt.Offset)
	if !opt.ListAll && opt.Limit > 0 {
		sql = sql.Limit(opt.Limit)
	}
	if err := sql.Find(&queryList.Data).Error; err != nil {
		logs.ErrorContextf(ctx, "ListForestGraph Retrieval project failed: %v", err)
		return nil, err
	}
	queryList.Limit = opt.Limit
	queryList.Offset = opt.Offset

	// 查询管理权限
	manageForestIDs := make(map[uint]bool)

	manageScopesIDs := perm.GetManageList(ctx, opt.Uin, foresttype.ResourceTypeGraph)
	for _, scope := range manageScopesIDs {
		manageForestIDs[scope] = true
	}

	for _, frs := range queryList.Data {
		frs.IsAdmin = manageForestIDs[frs.ID]
	}

	return queryList, nil
}

// ForestInfoItemList 知识森林信息
type ForestInfoItemList struct {
	apiobj.QueryResponse
	Data []*ForestGraphInfo
}

// ForestGraphInfo 图谱信息
type ForestGraphInfo struct {
	foresttype.ForestGraphInfo
	NodeCount int  `json:"node_count"`
	EdgeCount int  `json:"edge_count"`
	IsAdmin   bool `json:"is_admin"`
}

func GetGraphByCompanyID(ctx context.Context, companyID uint) ([]*foresttype.ForestGraphInfo, error) {
	var res []*foresttype.ForestGraphInfo
	if err := dbutil.Knownow().WithContext(ctx).
		Table(foresttype.ForestGraph{}.TableName()+" AS g").
		Select("g.*, v.status, v.file_id_list, v.space_name, v.parse_mode").
		Joins("LEFT JOIN "+foresttype.ForestGraphVersion{}.TableName()+" AS v ON g.version_id = v.id").
		Where("g.deleted_at IS NULL").
		// Where("v.status != ?", foresttype.GraphStatusDraft).
		Where("g.company_id = ?", companyID).
		Find(&res).Error; err != nil {
		logs.ErrorContextf(ctx, "GetGraphByCompanyID error: %v", err)
		return nil, err
	}
	return res, nil
}

// GetGraphsByForestIDs 批量获取 forest 对应的 graph 列表
// 返回 forestID -> graphInfo 的映射
func GetGraphsByForestIDs(ctx context.Context, forestIDs []uint) (map[uint]*foresttype.ForestGraphInfo, error) {
	if len(forestIDs) == 0 {
		return make(map[uint]*foresttype.ForestGraphInfo), nil
	}
	var graphs []*foresttype.ForestGraphInfo
	if err := dbutil.Knownow().WithContext(ctx).
		Table(foresttype.ForestGraph{}.TableName()+" AS g").
		Select("g.*, v.status, v.file_id_list, v.space_name, v.parse_mode").
		Joins("LEFT JOIN "+foresttype.ForestGraphVersion{}.TableName()+" AS v ON g.version_id = v.id").
		Where("g.deleted_at IS NULL").
		Where("g.forest_id IN ?", forestIDs).
		Find(&graphs).Error; err != nil {
		logs.ErrorContextf(ctx, "GetGraphsByForestIDs error: %v", err)
		return nil, err
	}
	result := make(map[uint]*foresttype.ForestGraphInfo, len(graphs))
	for _, g := range graphs {
		result[g.ForestID] = g
	}
	return result, nil
}

// ExistNodeName 检查节点名称是否存在
func ExistNodeName(ctx context.Context, graphID, versionID uint, nodeName string, tagID uint) bool {
	var cnt int64
	if err := dbutil.Knownow().WithContext(ctx).
		Table(foresttype.GraphTagNode{}.TableName()).
		Where("deleted_at IS NULL").
		Where("graph_id = ?", graphID).
		Where("graph_version_id = ?", versionID).
		Where("name = ?", nodeName).
		Where("tag_id = ?", tagID).
		Count(&cnt).Error; err != nil {
		logs.ErrorContextf(ctx, "ExistNodeName(graphID:%v, versionID:%v, nodeName:%s, tagID:%v) error: %v", graphID, versionID, nodeName, tagID, err)
		return false
	}
	return cnt > 0
}

// ExistEdgeName 检查边名称是否存在
func ExistEdgeName(ctx context.Context, graphID, versionID uint, edgeName string) bool {
	var cnt int64
	if err := dbutil.Knownow().WithContext(ctx).
		Table(foresttype.TableNameKeGraphTag+" AS tag").
		Where("tag.deleted_at IS NULL").
		Where("tag.graph_id = ?", graphID).
		Where("tag.graph_version_id = ?", versionID).
		Where("tag.tag_name = ? AND tag.tag_type = ?", edgeName, foresttype.TagTypeEdge).
		Count(&cnt).Error; err != nil {
		logs.ErrorContextf(ctx, "ExistEdgeName(graphID:%v, versionID:%v, edgeName:%s) error: %v", graphID, versionID, edgeName, err)
		return false
	}
	return cnt > 0
}
