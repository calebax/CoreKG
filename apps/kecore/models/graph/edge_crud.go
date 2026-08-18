package graph

import (
	"context"
	"fmt"
	"strings"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

// CreateGraphEdge 创建一个边
func CreateGraphEdge(ctx context.Context, edge *foresttype.GraphEdge) error {
	err := dbutil.Knownow().WithContext(ctx).Create(edge).Error
	if err != nil {
		logs.WarnContextf(ctx, "CreateGraphEdge: %v", err)
		return err
	}
	return nil
}

// GetGraphEdge retrieves a graph edge by its ID
func GetGraphEdge(ctx context.Context, graphID, scrID, dstID uint) (*foresttype.GraphEdge, error) {
	edge := &foresttype.GraphEdge{}
	err := dbutil.Knownow().WithContext(ctx).
		Where("graph_id = ? AND src_id = ? AND dst_id = ?", graphID, scrID, dstID).
		First(edge).Error
	if err != nil {
		logs.WarnContextf(ctx, "GetGraphEdge: %v", err)
		return nil, err
	}
	return edge, nil
}

// UpdateEdge 更新一个边
func UpdateEdge(ctx context.Context, edge *foresttype.GraphEdge) error {
	err := dbutil.Knownow().WithContext(ctx).Model(edge).Save(edge).Error
	if err != nil {
		logs.WarnContextf(ctx, "UpdateEdge: %v", err)
		return err
	}
	return nil
}

func GetEdgeTagInfoByTagID(ctx context.Context, graphID, tagID uint) ([]EdgeTagInfo, error) {
	var edge []EdgeTagInfo
	err := dbutil.Knownow().WithContext(ctx).
		Table(foresttype.TableNameKeGraphEdgeTag+" AS et").
		Select("et.*, tag1.tag_name AS src_tag_name, tag2.tag_name AS dst_tag_name, edge.tag_name AS edge_name").
		Where("et.graph_id = ?", graphID).
		Where("et.src_tag_id = ? OR et.dst_tag_id = ?", tagID, tagID).
		Where("et.deleted_at IS NULL").
		Joins("JOIN " + foresttype.TableNameKeGraphTag + " tag1 ON et.src_tag_id = tag1.id").
		Joins("JOIN " + foresttype.TableNameKeGraphTag + " tag2 ON et.dst_tag_id = tag2.id").
		Joins("JOIN " + foresttype.TableNameKeGraphTag + " edge ON et.edge_type_id = edge.id").
		Find(&edge).Error
	if err != nil {
		logs.WarnContextf(ctx, "GetEdgeTagInfoByTagID: %v", err)
		return nil, err
	}
	return edge, nil
}

type EdgeTagInfo struct {
	ID         uint   `json:"id"`
	SrcTagName string `json:"src_tag_name"`
	DstTagName string `json:"dst_tag_name"`
	EdgeName   string `json:"edge_name"`
	SrcTagID   uint   `json:"src_tag_id"`
	DstTagID   uint   `json:"dst_tag_id"`
	EdgeTypeID uint   `json:"edge_type_id"`
}

// DeleteEdgeByID 删除一个边
func DeleteEdgeByID(ctx context.Context, tx *gorm.DB, edgeID uint) error {
	err := tx.WithContext(ctx).
		Delete(&foresttype.GraphEdge{}, "id = ?", edgeID).Error
	if err != nil {
		logs.ErrorContextf(ctx, "DeleteEdgeByID: %v", err)
		return err
	}
	return nil
}

func DeleteEdgeByIDs(ctx context.Context, tx *gorm.DB, edgeIDs []uint) error {
	err := tx.WithContext(ctx).
		Delete(&foresttype.GraphEdge{}, "id in ?", edgeIDs).Error
	if err != nil {
		logs.ErrorContextf(ctx, "DeleteEdgeByID: %v", err)
		return err
	}
	return nil
}

// DeleteEdgeWithNodeID 根据节点id删除一个边
func DeleteEdgeWithNodeID(ctx context.Context, tx *gorm.DB, nodeID uint) error {
	err := tx.WithContext(ctx).
		Delete(&foresttype.GraphEdge{}, "src_id = ? OR dst_id = ?", nodeID, nodeID).Error
	if err != nil {
		logs.ErrorContextf(ctx, "DeleteEdgeWithNodeID: %v", err)
		return err
	}
	return nil
}

// DeleteEdgeWithNodeID 根据节点id删除一个边
func DeleteEdgeWithNodeIDs(ctx context.Context, tx *gorm.DB, nodeIDs []uint) error {
	err := tx.WithContext(ctx).
		Delete(&foresttype.GraphEdge{}, "src_id in ? OR dst_id in ?", nodeIDs, nodeIDs).Error
	if err != nil {
		logs.ErrorContextf(ctx, "DeleteEdgeWithNodeID: %v", err)
		return err
	}
	return nil
}

// GetEdgeInfo 获取边信息
func GetEdgeInfo(ctx context.Context, edgeID uint) (*foresttype.EdgeInfo, error) {
	var edgeInfo foresttype.EdgeInfo
	err := dbutil.Knownow().WithContext(ctx).
		Table(foresttype.TableNameKeGraphEdge+" AS e").
		Select("e.graph_id AS graph_id"+
			", e.graph_version_id AS graph_version_id"+
			",e.id AS edge_id"+
			",e.src_id AS src_id"+
			",e.dst_id AS dst_id"+
			",e.tag_id AS tag_id"+
			",e.src_tag_id AS src_tag_id"+
			",e.dst_tag_id AS dst_tag_id"+
			",e.file_id_list AS file_id_list"+
			",e.properties_values AS properties_values"+
			",src.name AS src_node_name"+
			",dst.name AS dst_node_name"+
			",tag.tag_name AS edge_type_name").
		Joins("JOIN "+foresttype.GraphTagNode{}.TableName()+" src ON e.src_id = src.id").
		Joins("JOIN "+foresttype.GraphTagNode{}.TableName()+" dst ON e.dst_id = dst.id").
		Joins("JOIN "+foresttype.TableNameKeGraphTag+" tag ON tag.id = e.tag_id").
		Where("e.id = ?", edgeID).
		Find(&edgeInfo).Error
	if err != nil {
		logs.WarnContextf(ctx, "GetEdgeInfo: %v", err)
		return nil, err
	}
	return nil, nil
}

func GetEdgesByNodeName(ctx context.Context, graphID, versionID uint, nodeName string, tagID uint) ([]*foresttype.EdgeInfo, error) {
	var edgeInfos []*foresttype.EdgeInfo

	if err := dbutil.Knownow().WithContext(ctx).
		Table(foresttype.TableNameKeGraphEdge+" AS e").
		Select("e.graph_id AS graph_id"+
			", e.graph_version_id AS graph_version_id"+
			",e.id AS edge_id"+
			",e.src_id AS src_id"+
			",e.dst_id AS dst_id"+
			",e.tag_id AS tag_id"+
			",e.src_tag_id AS src_tag_id"+
			",e.dst_tag_id AS dst_tag_id"+
			",e.file_id_list AS file_id_list"+
			",e.properties_values AS properties_values"+
			",src.name AS src_node_name"+
			",dst.name AS dst_node_name"+
			",tag.tag_name AS edge_type_name").
		Joins("JOIN "+foresttype.GraphTagNode{}.TableName()+" src ON e.src_id = src.id").
		Joins("JOIN "+foresttype.GraphTagNode{}.TableName()+" dst ON e.dst_id = dst.id").
		Joins("JOIN "+foresttype.TableNameKeGraphTag+" tag ON tag.id = e.tag_id").
		Where("e.graph_id = ?", graphID).
		Where("e.graph_version_id = ?", versionID).
		Where("((src.name = ? AND src.tag_id = ?) OR (dst.name = ? AND dst.tag_id = ?))", nodeName, tagID, nodeName, tagID).
		Find(&edgeInfos).Error; err != nil {
		logs.WarnContextf(ctx, "GetEdgesByNodeName(graphID:%v, versionID:%v, nodeName:%s, tagID:%v) error: %v", graphID, versionID, nodeName, tagID, err)
		return nil, err
	}

	return edgeInfos, nil
}

// GetEdgesByNodeNameAll 根据节点名称获取所有边（不区分tagID）
func GetEdgesByNodeNameAll(ctx context.Context, graphID, versionID uint, nodeName string) ([]*foresttype.EdgeInfo, error) {
	var edgeInfos []*foresttype.EdgeInfo

	if err := dbutil.Knownow().WithContext(ctx).
		Table(foresttype.TableNameKeGraphEdge+" AS e").
		Select("e.graph_id AS graph_id"+
			", e.graph_version_id AS graph_version_id"+
			",e.id AS edge_id"+
			",e.src_id AS src_id"+
			",e.dst_id AS dst_id"+
			",e.tag_id AS tag_id"+
			",e.src_tag_id AS src_tag_id"+
			",e.dst_tag_id AS dst_tag_id"+
			",e.file_id_list AS file_id_list"+
			",e.properties_values AS properties_values"+
			",src.name AS src_node_name"+
			",dst.name AS dst_node_name"+
			",tag.tag_name AS edge_type_name").
		Joins("JOIN "+foresttype.GraphTagNode{}.TableName()+" src ON e.src_id = src.id").
		Joins("JOIN "+foresttype.GraphTagNode{}.TableName()+" dst ON e.dst_id = dst.id").
		Joins("JOIN "+foresttype.TableNameKeGraphTag+" tag ON tag.id = e.tag_id").
		Where("e.graph_id = ?", graphID).
		Where("e.graph_version_id = ?", versionID).
		Where("(src.name = ? OR dst.name = ?)", nodeName, nodeName).
		Where("e.deleted_at IS NULL").
		Where("src.deleted_at IS NULL").
		Where("dst.deleted_at IS NULL").
		Where("tag.deleted_at IS NULL").
		Find(&edgeInfos).Error; err != nil {
		logs.WarnContextf(ctx, "GetEdgesByNodeNameAll(graphID:%v, versionID:%v, nodeName:%s) error: %v", graphID, versionID, nodeName, err)
		return nil, err
	}

	return edgeInfos, nil
}

// GetEdgesBySrcNodeName 获取指定节点作为起点的所有边
func GetEdgesBySrcNodeName(ctx context.Context, graphID, versionID uint, nodeName string, tagID uint) ([]*foresttype.EdgeInfo, error) {
	var edgeInfos []*foresttype.EdgeInfo

	if err := dbutil.Knownow().WithContext(ctx).
		Table(foresttype.TableNameKeGraphEdge+" AS e").
		Select("e.graph_id AS graph_id"+
			", e.graph_version_id AS graph_version_id"+
			",e.id AS edge_id"+
			",e.src_id AS src_id"+
			",e.dst_id AS dst_id"+
			",e.tag_id AS tag_id"+
			",e.src_tag_id AS src_tag_id"+
			",e.dst_tag_id AS dst_tag_id"+
			",e.file_id_list AS file_id_list"+
			",e.properties_values AS properties_values"+
			",src.name AS src_node_name"+
			",dst.name AS dst_node_name"+
			",tag.tag_name AS edge_type_name").
		Joins("JOIN "+foresttype.GraphTagNode{}.TableName()+" src ON e.src_id = src.id").
		Joins("JOIN "+foresttype.GraphTagNode{}.TableName()+" dst ON e.dst_id = dst.id").
		Joins("JOIN "+foresttype.TableNameKeGraphTag+" tag ON tag.id = e.tag_id").
		Where("e.graph_id = ?", graphID).
		Where("e.graph_version_id = ?", versionID).
		Where("src.name = ? AND src.tag_id = ?", nodeName, tagID).
		Where("e.deleted_at IS NULL").
		Where("src.deleted_at IS NULL").
		Where("dst.deleted_at IS NULL").
		Where("tag.deleted_at IS NULL").
		Find(&edgeInfos).Error; err != nil {
		logs.WarnContextf(ctx, "GetEdgesBySrcNodeName(graphID:%v, versionID:%v, nodeName:%s, tagID:%v) error: %v", graphID, versionID, nodeName, tagID, err)
		return nil, err
	}

	return edgeInfos, nil
}

// AlreadyHasEdgeByNodeName 检查是否已经存在名为name的单向边
func AlreadyHasEdgeByNodeName(ctx context.Context, graphID, versionID uint, edgeName, srcNodeName, dstNodeName string) bool {
	var cnt int64
	if err := dbutil.Knownow().WithContext(ctx).
		Table(foresttype.TableNameKeGraphEdge+" AS e").
		Where("e.graph_id = ?", graphID).
		Where("e.graph_version_id = ?", versionID).
		Joins("JOIN "+foresttype.GraphTagNode{}.TableName()+" src ON e.src_id = src.id").
		Joins("JOIN "+foresttype.GraphTagNode{}.TableName()+" dst ON e.dst_id = dst.id").
		Joins("JOIN "+foresttype.TableNameKeGraphTag+" tag ON tag.id = e.tag_id").
		Where("tag.tag_name = ? AND tag.tag_type = ?", edgeName, foresttype.TagTypeEdge).
		Where("src.name = ? AND dst.name = ?", srcNodeName, dstNodeName).
		Count(&cnt).Error; err != nil {
		logs.WarnContextf(ctx, "AlreadyHasEdgeByNodeName(graphID:%v, versionID:%v, srcNodeName:%s, dstNodeName:%s) error: %v", graphID, versionID, srcNodeName, dstNodeName, err)
		return false
	}
	return cnt > 0
}

// AlreadyHasEdgeByTagID 检查是否已经存在边关系
func AlreadyHasEdgeByTagID(ctx context.Context, graphID, versionID uint, edgeTagID, srcID, dstID uint) bool {
	var cnt int64
	if err := dbutil.Knownow().WithContext(ctx).
		Table(foresttype.TableNameKeGraphEdge+" AS e").
		Where("e.graph_id = ?", graphID).
		Where("e.graph_version_id = ?", versionID).
		Where("e.tag_id = ?", edgeTagID).
		Where("e.src_id = ? AND e.dst_id = ?", srcID, dstID).
		Count(&cnt).Error; err != nil {
		logs.WarnContextf(ctx, "AlreadyHasEdgeByTagID(graphID:%v, versionID:%v, edgeTagID:%v, srcID:%v, dstID:%v) error: %v", graphID, versionID, edgeTagID, srcID, dstID, err)
		return false
	}
	return cnt > 0
}

// AlreadyHasTagRelation 检查是否已经存在Tag边
func AlreadyHasTagRelation(ctx context.Context, graphID, versionID uint, srcTagID, dstTagID uint) bool {
	var cnt int64
	if err := dbutil.Knownow().WithContext(ctx).
		Table(foresttype.TableNameKeGraphEdgeTag+" AS et").
		Where("et.graph_id = ?", graphID).
		Where("et.graph_version_id = ?", versionID).
		Where("et.src_tag_id = ? AND et.dst_tag_id = ?", srcTagID, dstTagID).
		Count(&cnt).Error; err != nil {
		logs.WarnContextf(ctx, "AlreadyHasTagRelation(graphID:%v, versionID:%v, srcTagID:%v, dstTagID:%v) error: %v", graphID, versionID, srcTagID, dstTagID, err)
		return false
	}
	return cnt > 0
}

// TagRelationKey 生成 tag relation 的唯一键
func TagRelationKey(srcTagID, dstTagID uint) string {
	return fmt.Sprintf("%d:%d", srcTagID, dstTagID)
}

// GetExistingTagRelations 批量查询已存在的 tag relations
// 返回 map[srcTagID:dstTagID]bool
func GetExistingTagRelations(ctx context.Context, graphID, versionID uint, pairs [][2]uint) (map[string]bool, error) {
	if len(pairs) == 0 {
		return make(map[string]bool), nil
	}

	// 构建查询条件
	var conditions []string
	var args []any
	for _, pair := range pairs {
		conditions = append(conditions, "(et.src_tag_id = ? AND et.dst_tag_id = ?)")
		args = append(args, pair[0], pair[1])
	}

	var results []struct {
		SrcTagID uint `gorm:"column:src_tag_id"`
		DstTagID uint `gorm:"column:dst_tag_id"`
	}

	query := dbutil.Knownow().WithContext(ctx).
		Table(foresttype.TableNameKeGraphEdgeTag+" AS et").
		Select("et.src_tag_id, et.dst_tag_id").
		Where("et.graph_id = ?", graphID).
		Where("et.graph_version_id = ?", versionID).
		Where("et.deleted_at IS NULL")

	// 添加条件
	condStr := "(" + strings.Join(conditions, " OR ") + ")"
	query = query.Where(condStr, args...)

	if err := query.Find(&results).Error; err != nil {
		logs.WarnContextf(ctx, "GetExistingTagRelations error: %v", err)
		return nil, err
	}

	result := make(map[string]bool, len(results))
	for _, r := range results {
		key := TagRelationKey(r.SrcTagID, r.DstTagID)
		result[key] = true
	}
	return result, nil
}

// GetLastTagRelationEdgeTagIDs 获取最后一个关系边的tag id
func GetLastTagRelationEdgeTagIDs(ctx context.Context, graphID, versionID uint, edgeTagIDs []uint) ([]uint, error) {
	if len(edgeTagIDs) == 0 {
		return []uint{}, nil
	}

	var result []uint
	if err := dbutil.Knownow().WithContext(ctx).
		Table(foresttype.TableNameKeGraphEdgeTag+" AS et").
		Where("et.deleted_at IS NULL").
		Where("et.graph_id = ?", graphID).
		Where("et.graph_version_id = ?", versionID).
		Where("et.edge_type_id in ?", edgeTagIDs).
		Group("et.edge_type_id").
		Having("COUNT(*) < ?", 2).
		Pluck("et.edge_type_id", &result).Error; err != nil {
		logs.WarnContextf(ctx, "GetLastTagRelationEdgeTagIDs error: %v", err)
		return nil, err
	}

	return result, nil
}

func GetEdgeByGraphID(ctx context.Context, graphID, graphVersionID uint) ([]*foresttype.GraphTag, error) {
	var edgeTags []*foresttype.GraphTag
	err := dbutil.Knownow().WithContext(ctx).
		Where("deleted_at IS NULL").
		Where("graph_id = ?", graphID).
		Where("graph_version_id = ?", graphVersionID).
		Where("tag_type = ?", foresttype.TagTypeEdge).
		Find(&edgeTags).Error
	if err != nil {
		logs.WarnContextf(ctx, "GetEdgeByGraphID error: %v", err)
		return edgeTags, nil
	}
	return edgeTags, nil
}
