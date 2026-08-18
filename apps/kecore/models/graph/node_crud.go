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

// GetNodeByID get node 获取节点
// func GetNodeByID(ctx context.Context, id uint) (*foresttype.GraphNode, error) {
// 	node := &foresttype.GraphNode{}
// 	err := dbutil.Knownow().WithContext(ctx).Where("id = ? ", id).First(node).Error
// 	if err != nil {
// 		logs.ErrorContextf(ctx, "GetNodeByID error: %v", err)
// 		return nil, err
// 	}
// 	return node, nil
// }

// GetNodeByName get node by name 获取节点
// func GetNodeByName(ctx context.Context, graph_id, graphVersion, tagID uint, name string) (*foresttype.GraphNode, error) {
// 	node := &foresttype.GraphNode{}
// 	err := dbutil.Knownow().WithContext(ctx).Where("name = ? ", name).
// 		Where("tag_id = ?", tagID).
// 		Where("graph_id = ?", graph_id).
// 		Where("graph_version_id = ?", graphVersion).First(node).Error
// 	if err != nil {
// 		logs.WarnContextf(ctx, "GetNodeByName error: %v", err)
// 		return nil, err
// 	}
// 	return node, nil
// }

// CreateNode create node 创建节点
// func CreateNode(ctx context.Context, node *foresttype.GraphNode) error {
// 	err := dbutil.Knownow().WithContext(ctx).Create(node).Error
// 	if err != nil {
// 		logs.ErrorContextf(ctx, "CreateNode error: %v", err)
// 		return err
// 	}
// 	return nil
// }

// CreateNodeTX create node 创建节点
// func CreateNodeTX(ctx context.Context, tx *gorm.DB, node *foresttype.GraphNode) error {
// 	err := dbutil.Knownow().WithContext(ctx).Create(node).Error
// 	if err != nil {
// 		logs.ErrorContextf(ctx, "CreateNode error: %v", err)
// 		return err
// 	}
// 	return nil
// }

// CreateTagNode create tag node 创建标签节点
func CreateTagNode(ctx context.Context, tNode *foresttype.GraphTagNode) error {
	err := dbutil.Knownow().WithContext(ctx).Create(tNode).Error
	if err != nil {
		logs.ErrorContextf(ctx, "CreateTagNode error: %v", err)
		return err
	}
	return nil
}

func CreateTagNodeTX(ctx context.Context, tx *gorm.DB, tNode *foresttype.GraphTagNode) error {
	err := tx.WithContext(ctx).Create(tNode).Error
	if err != nil {
		logs.ErrorContextf(ctx, "CreateTagNode error: %v", err)
		return err
	}
	return nil
}

// CreateNewNode create new node 创建新节点并且连接一个tag
// func CreateNewNode(ctx context.Context, db *gorm.DB, node *foresttype.GraphNode, tNode *foresttype.GraphTagNode) error {
// 	// 开启事务
// 	err := db.Transaction(func(tx *gorm.DB) error {
// 		err := tx.Create(node).Error
// 		if err != nil {
// 			logs.ErrorContextf(ctx, "CreateNewNode error: %v", err)
// 			return err
// 		}
// 		tNode.NodeID = node.ID
// 		err = tx.Create(tNode).Error
// 		if err != nil {
// 			logs.ErrorContextf(ctx, "CreateNewNode error: %v", err)
// 			return err
// 		}
// 		return nil
// 	})
// 	if err != nil {
// 		logs.ErrorContextf(ctx, "CreateNewNode error: %v", err)
// 		return err
// 	}
// 	return nil
// }

// GetTNodeByTagIDNodeID get tag node by tag id and node id 获取标签节点
func GetTNodeByTagIDNodeID(ctx context.Context, tagID uint, nodeName string) (*foresttype.GraphTagNode, error) {
	tNode := &foresttype.GraphTagNode{}
	err := dbutil.Knownow().WithContext(ctx).Where("tag_id = ? and name = ?", tagID, nodeName).
		First(tNode).Error
	if err != nil {
		logs.WarnContextf(ctx, "GetTNodeByTagIDNodeID error: %v", err)
		return nil, err
	}
	return tNode, nil
}

// UpdateTNode update tag node 更新标签节点
// func UpdateTNode(ctx context.Context, tNode *foresttype.GraphTagNode) error {
// 	err := dbutil.Knownow().WithContext(ctx).Save(tNode).Error
// 	if err != nil {
// 		logs.ErrorContextf(ctx, "UpdateTNode error: %v", err)
// 		return err
// 	}
// 	return nil
// }

// GetTNodeByID 获取标签节点
func GetTNodeByID(ctx context.Context, id uint) (*foresttype.GraphTagNode, error) {
	tNode := &foresttype.GraphTagNode{}
	err := dbutil.Knownow().WithContext(ctx).Where("id = ?", id).First(tNode).Error
	if err != nil {
		logs.WarnContextf(ctx, "GetTNodeByTagIDNodeID error: %v", err)
		return nil, err
	}
	return tNode, nil
}

// GetNodeNameByNodeIDs 根据标签节点id获取节点名称
func GetNodeNameByNodeIDs(ctx context.Context, nodeIDs []uint) ([]string, error) {
	if len(nodeIDs) == 0 {
		return []string{}, nil
	}
	var results []struct {
		ID   uint   `gorm:"column:id"`
		Name string `gorm:"column:name"`
	}

	if err := dbutil.Knownow().WithContext(ctx).
		Table(foresttype.GraphTagNode{}.TableName()).
		// Joins("JOIN "+foresttype.GraphNode{}.TableName()+" as n ON t.node_id = n.id").
		Select("id, name").
		Where("id IN ?", nodeIDs).
		Find(&results).Error; err != nil {
		logs.ErrorContextf(ctx, "GetNodeNameByNodeIDs error: %v", err)
		return nil, fmt.Errorf("failed to query node names: %w", err)
	}

	// 创建 ID 到 Name 的映射
	nameMap := make(map[uint]string, len(results))
	for _, r := range results {
		nameMap[r.ID] = r.Name
	}

	// 按照输入顺序提取名称
	names := make([]string, 0, len(nodeIDs))
	for _, id := range nodeIDs {
		if name, ok := nameMap[id]; ok {
			names = append(names, name)
		} else {
			logs.WarnContextf(ctx, "GetNodeNameByNodeIDs: node ID %d not found", id)
			names = append(names, "")
		}
	}

	return names, nil
}

// GetNodeIDMapByNodeNames 根据节点名称和图的 ID/VersionID 获取节点 ID 映射
// 返回 map[nodeName]nodeID
func GetNodeIDMapByNodeNames(ctx context.Context, graphID, graphVersionID uint, nodeNames []string) (map[string]uint, error) {
	if len(nodeNames) == 0 {
		return make(map[string]uint), nil
	}

	var results []struct {
		ID   uint   `gorm:"column:id"`
		Name string `gorm:"column:name"`
	}

	if err := dbutil.Knownow().WithContext(ctx).
		Table(foresttype.GraphTagNode{}.TableName()).
		Select("id, name").
		Where("graph_id = ? AND graph_version_id = ? AND name IN ?", graphID, graphVersionID, nodeNames).
		Find(&results).Error; err != nil {
		logs.ErrorContextf(ctx, "GetNodeIDMapByNodeNames error: %v", err)
		return nil, fmt.Errorf("failed to query node IDs: %w", err)
	}

	// 创建 Name 到 ID 的映射
	idMap := make(map[string]uint, len(results))
	for _, r := range results {
		idMap[r.Name] = r.ID
	}

	return idMap, nil
}

// GetNodeEntityMapByNodeNames 根据节点名称和图的 ID/VersionID 获取节点实体映射
// nodeTagMap: map[nodeName]tagID，用于唯一标识节点（tag_id + name 才是唯一标识）
// 返回 map[nodeName]nodeEntity
func GetNodeEntityMapByNodeNames(ctx context.Context, graphID, graphVersionID uint, nodeTagMap map[string]uint) (map[string]*foresttype.GraphTagNode, error) {
	if len(nodeTagMap) == 0 {
		return make(map[string]*foresttype.GraphTagNode), nil
	}

	// 构建查询条件：需要同时匹配 name 和 tag_id
	var conditions []string
	var args []any
	for nodeName, tagID := range nodeTagMap {
		conditions = append(conditions, "(name = ? AND tag_id = ?)")
		args = append(args, nodeName, tagID)
	}

	var results []*foresttype.GraphTagNode
	query := dbutil.Knownow().WithContext(ctx).
		Table(foresttype.GraphTagNode{}.TableName()).
		Where("graph_id = ?", graphID).
		Where("graph_version_id = ?", graphVersionID).
		Where("deleted_at IS NULL")

	// 添加条件
	if len(conditions) > 0 {
		condStr := "(" + strings.Join(conditions, " OR ") + ")"
		query = query.Where(condStr, args...)
	}

	if err := query.Find(&results).Error; err != nil {
		logs.ErrorContextf(ctx, "GetNodeEntityMapByNodeNames(graphID:%v, versionID:%v, nodeTagMap:%v) error: %v", graphID, graphVersionID, nodeTagMap, err)
		return nil, fmt.Errorf("failed to query node entities: %w", err)
	}

	nodeMap := make(map[string]*foresttype.GraphTagNode, len(results))
	for _, r := range results {
		nodeMap[r.Name] = r
	}

	return nodeMap, nil
}

// GetTNodeByName 根据图的 ID、版本 ID、节点名称和 tagID 获取节点
// tag_id + name 才是唯一标识
func GetTNodeByName(ctx context.Context, graphID, graphVersionID uint, nodeName string, tagID uint) (*foresttype.GraphTagNode, error) {
	tNode := &foresttype.GraphTagNode{}
	err := dbutil.Knownow().WithContext(ctx).
		Where("graph_id = ? AND graph_version_id = ? AND name = ? AND tag_id = ? AND deleted_at IS NULL", graphID, graphVersionID, nodeName, tagID).
		First(tNode).Error
	if err != nil {
		logs.WarnContextf(ctx, "GetTNodeByName(graphID:%v, versionID:%v, nodeName:%s, tagID:%v) error: %v", graphID, graphVersionID, nodeName, tagID, err)
		return nil, err
	}
	return tNode, nil
}

// GetTNodesByName 根据图的 ID、版本 ID、节点名称获取所有tag节点（不区分tagID）
func GetTNodesByName(ctx context.Context, graphID, graphVersionID uint, nodeName string) ([]*foresttype.GraphTagNode, error) {
	var tNodes []*foresttype.GraphTagNode
	err := dbutil.Knownow().WithContext(ctx).
		Where("graph_id = ? AND graph_version_id = ? AND name = ? AND deleted_at IS NULL", graphID, graphVersionID, nodeName).
		Find(&tNodes).Error
	if err != nil {
		logs.WarnContextf(ctx, "GetTNodesByName(graphID:%v, versionID:%v, nodeName:%s) error: %v", graphID, graphVersionID, nodeName, err)
		return nil, err
	}
	return tNodes, nil
}

// DeleteNode 删除节点
// func DeleteNode(ctx context.Context, tx *gorm.DB, node *foresttype.GraphNode) error {
// 	err := tx.WithContext(ctx).
// 		Where("id = ?", node.ID).
// 		Delete(&foresttype.GraphNode{}).Error
// 	if err != nil {
// 		logs.ErrorContextf(ctx, "DeleteNode error: %v", err)
// 		return err
// 	}

// 	return nil
// }

// UpdateNode 更新节点
// func UpdateNode(ctx context.Context, tx *gorm.DB, node *foresttype.GraphNode) error {
// 	err := tx.WithContext(ctx).
// 		Save(node).Error
// 	if err != nil {
// 		logs.ErrorContextf(ctx, "UpdateNode error: %v", err)
// 		return err
// 	}
// 	return nil
// }

// DeleteTNode 删除节点
func DeleteTNode(ctx context.Context, tx *gorm.DB, tnode *foresttype.GraphTagNode) error {
	err := tx.WithContext(ctx).
		Where("id = ?", tnode.ID).
		Delete(&foresttype.GraphTagNode{}).Error
	if err != nil {
		logs.ErrorContextf(ctx, "DeleteTNode error: %v", err)
		return err
	}

	return nil
}

func DeleteTNodeByIDs(ctx context.Context, tx *gorm.DB, ids []uint) error {
	err := tx.WithContext(ctx).
		Where("id in ?", ids).
		Delete(&foresttype.GraphTagNode{}).Error
	if err != nil {
		logs.ErrorContextf(ctx, "DeleteTNode error: %v", err)
		return err
	}

	return nil
}
