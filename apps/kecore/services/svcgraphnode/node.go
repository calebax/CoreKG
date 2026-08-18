package svcgraphnode

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtographnode"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/graph"
	"github.com/insmtx/corekg/apps/kecore/models/nebulagraph"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

// edgeToCreate 表示需要创建的边及其相关信息
type edgeToCreate struct {
	edgeTag       *foresttype.GraphTag
	edge          *foresttype.GraphEdge
	srcNodeName   string
	dstNodeName   string
	edgeTagID     uint // 边的 tag ID（如果已存在）
	needCreateTag bool // 是否需要创建边的 tag
	needCreateRel bool // 是否需要创建 tag relation
	srcTagID      uint // 源节点 tag ID
	dstTagID      uint // 目标节点 tag ID
}

// editNodeContext 编辑节点的上下文数据
type editNodeContext struct {
	graphInfo     *foresttype.ForestGraphInfo
	tag           *foresttype.GraphTag
	oldTNode      *foresttype.GraphTagNode
	cli           *nebulagraph.NebulaCli
	nodeEntityMap map[string]*foresttype.GraphTagNode
	edgeTagMap    map[string]*foresttype.GraphTag
	oldEdgeMap    map[string]*foresttype.EdgeInfo
	// 原始值备份，用于比较是否需要更新
	originalNodeName         string
	originalPropertiesValues foresttype.PropertiesValues
	originalTagProperties    foresttype.Properties
}

// validateEditNodeRequest 验证编辑节点请求
func validateEditNodeRequest(ctx *gin.Context, req *dtographnode.EditNodeRequest) (*editNodeContext, *dtographnode.EditNodeResponse, error) {
	res := &dtographnode.EditNodeResponse{}
	ectx := &editNodeContext{}

	// 获取图谱信息
	graphInfo, err := graph.GetGraph(ctx, req.Request.GraphID)
	if err != nil {
		logs.ErrorContextf(ctx, "EditNode.GetGraph(graphID:%v) error: %v", req.Request.GraphID, err)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_graph_not_found"
		return nil, res, nil
	}
	ectx.graphInfo = graphInfo

	// 获取 Tag
	tag, err := graph.GetTagByID(ctx, req.Request.Tags[0].TagID)
	if err != nil {
		logs.ErrorContextf(ctx, "EditNode.GetTag(id:%v) error: %v", req.Request.Tags[0].TagID, err)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_tag_not_found"
		return nil, res, nil
	}
	ectx.tag = tag

	// 检测旧节点是否存在
	oldTNode, err := graph.GetTNodeByName(ctx, graphInfo.ID, graphInfo.VersionID, req.Request.OldNodeName, tag.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "EditNode.GetTNodeByName(graphID:%v, versionID:%v, oldNodeName:%s, tagID:%v) error: %v", graphInfo.ID, graphInfo.VersionID, req.Request.OldNodeName, tag.ID, err)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_old_node_not_found"
		return nil, res, nil
	}
	ectx.oldTNode = oldTNode

	// 保存原始值用于后续比较（使用 JSON 序列化进行深拷贝）
	ectx.originalNodeName = oldTNode.Name
	if oldTNode.PropertiesValues != nil {
		jsonData, _ := json.Marshal(oldTNode.PropertiesValues)
		json.Unmarshal(jsonData, &ectx.originalPropertiesValues)
	}
	if tag.Properties != nil {
		jsonData, _ := json.Marshal(tag.Properties)
		json.Unmarshal(jsonData, &ectx.originalTagProperties)
	}

	// 创建 Nebula CLI
	cli, err := nebulagraph.NewNebulaCLI(ctx, graphInfo.SpaceName)
	if err != nil {
		logs.ErrorContextf(ctx, "EditNode.NewNebulaCLI(spaceName:%s) error: %v", graphInfo.SpaceName, err)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_create_graph_cli_failed"
		return nil, res, nil
	}
	ectx.cli = cli

	return ectx, res, nil
}

// prepareNodeAndTag 准备节点和 tag 数据
func prepareNodeAndTag(req *dtographnode.EditNodeRequest, ectx *editNodeContext) {
	// 更新 tag 属性
	ectx.tag.Properties = req.Request.Tags[0].Properties

	// 更新节点属性（不更新节点名称）
	ectx.oldTNode.PropertiesValues = req.Request.Tags[0].PropertiesValues
}

// collectEdgeRelatedData 收集边相关的节点和 tag 数据
func collectEdgeRelatedData(ctx *gin.Context, req *dtographnode.EditNodeRequest, ectx *editNodeContext) (*dtographnode.EditNodeResponse, error) {
	res := &dtographnode.EditNodeResponse{}

	// 收集节点 tag 映射和边名称
	// 注意：现在只处理起点是当前节点的边，所以源节点一定是当前节点
	nodeTagMap := make(map[string]uint)
	edgeNamesSet := make(map[string]bool)
	for _, v := range req.Request.Edges {
		// 源节点必须是当前节点（已在请求验证中检查）
		// 只收集目标节点的 tag_id
		if v.DstNodeName != "" {
			if v.DstTagID == 0 {
				logs.ErrorContextf(ctx, "EditNode: destination node '%s' missing dst_tag_id", v.DstNodeName)
				res.Code = errcode.ErrCode_BadRequest
				res.Message = "kecore_dst_node_missing_tag_id"
				return res, nil
			}
			nodeTagMap[v.DstNodeName] = v.DstTagID
		}
		if v.EdgeName != "" {
			edgeNamesSet[v.EdgeName] = true
		}
	}

	// 批量获取节点实体映射
	nodeEntityMap := make(map[string]*foresttype.GraphTagNode)
	if len(nodeTagMap) > 0 {
		var err error
		nodeEntityMap, err = graph.GetNodeEntityMapByNodeNames(ctx, ectx.graphInfo.ID, ectx.graphInfo.VersionID, nodeTagMap)
		if err != nil {
			logs.ErrorContextf(ctx, "EditNode.GetNodeEntityMapByNodeNames(graphID:%v, versionID:%v, nodeTagMap:%v) error: %v", ectx.graphInfo.ID, ectx.graphInfo.VersionID, nodeTagMap, err)
			res.Code = errcode.ErrCode_BadRequest
			res.Message = "kecore_get_node_id_failed"
			return res, nil
		}
	}
	ectx.nodeEntityMap = nodeEntityMap

	// 批量获取边 tag 映射
	edgeNames := make([]string, 0, len(edgeNamesSet))
	for name := range edgeNamesSet {
		edgeNames = append(edgeNames, name)
	}

	edgeTagMap := make(map[string]*foresttype.GraphTag)
	if len(edgeNames) > 0 {
		var err error
		edgeTagMap, err = graph.GetEdgesByNames(ctx, ectx.graphInfo.ID, ectx.graphInfo.VersionID, edgeNames)
		if err != nil {
			logs.ErrorContextf(ctx, "EditNode.GetEdgesByNames(graphID:%v, versionID:%v, edgeNames:%v) error: %v", ectx.graphInfo.ID, ectx.graphInfo.VersionID, edgeNames, err)
			res.Code = errcode.ErrCode_BadRequest
			res.Message = "kecore_get_edge_type_failed"
			return res, nil
		}
	}
	ectx.edgeTagMap = edgeTagMap

	return res, nil
}

// calculateEdgeDiff 计算需要删除和保留的边
func calculateEdgeDiff(ctx *gin.Context, req *dtographnode.EditNodeRequest, ectx *editNodeContext) ([]*foresttype.EdgeInfo, *dtographnode.EditNodeResponse, error) {
	res := &dtographnode.EditNodeResponse{}

	// 获取旧节点作为起点的所有边（只查询起点是当前节点的边）
	oldEdges, err := graph.GetEdgesBySrcNodeName(ctx, ectx.graphInfo.ID, ectx.graphInfo.VersionID, req.Request.OldNodeName, ectx.tag.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "EditNode.GetEdgesByNodeName(graphID:%v, versionID:%v, oldNodeName:%s, tagID:%v) error: %v", ectx.graphInfo.ID, ectx.graphInfo.VersionID, req.Request.OldNodeName, ectx.tag.ID, err)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_get_old_node_edges_failed"
		return nil, res, nil
	}

	// 构建旧边的唯一标识映射
	oldEdgeMap := make(map[string]*foresttype.EdgeInfo)
	for _, oldEdge := range oldEdges {
		srcName := req.Request.OldNodeName
		dstName := oldEdge.DstNodeName
		key := fmt.Sprintf("%s:%s:%s", oldEdge.EdgeTypeName, srcName, dstName)
		oldEdgeMap[key] = oldEdge
	}
	ectx.oldEdgeMap = oldEdgeMap

	// 构建新边的唯一标识集合
	newEdgeKeys := make(map[string]bool)
	for _, v := range req.Request.Edges {
		key := fmt.Sprintf("%s:%s:%s", v.EdgeName, v.SrcNodeName, v.DstNodeName)
		newEdgeKeys[key] = true
	}

	// 计算要删除的边：在 oldEdgeMap 中但不在 newEdgeKeys 中
	edgesToDelete := make([]*foresttype.EdgeInfo, 0)
	for key, oldEdge := range oldEdgeMap {
		if !newEdgeKeys[key] {
			edgesToDelete = append(edgesToDelete, oldEdge)
		}
	}

	return edgesToDelete, res, nil
}

// prepareEdgesToCreate 准备要创建的边数据
func prepareEdgesToCreate(ctx *gin.Context, req *dtographnode.EditNodeRequest, ectx *editNodeContext) ([]*edgeToCreate, *dtographnode.EditNodeResponse, error) {
	res := &dtographnode.EditNodeResponse{}

	tagRelationPairs := make([][2]uint, 0, len(req.Request.Edges))
	edgesToCreate := make([]*edgeToCreate, 0, len(req.Request.Edges))

	for _, v := range req.Request.Edges {
		key := fmt.Sprintf("%s:%s:%s", v.EdgeName, v.SrcNodeName, v.DstNodeName)
		// 如果边已存在，跳过
		if ectx.oldEdgeMap[key] != nil {
			continue
		}

		// 确定源节点和目标节点
		srcTagID := ectx.tag.ID
		srcID := ectx.oldTNode.ID

		// 目标节点（不能是当前节点）
		expectedDstTagID := v.DstTagID
		if expectedDstTagID == 0 {
			logs.ErrorContextf(ctx, "EditNode: destination node '%s' missing dst_tag_id", v.DstNodeName)
			res.Code = errcode.ErrCode_BadRequest
			res.Message = "kecore_dst_node_missing_tag_id"
			return nil, res, nil
		}
		dstNodeEntity := ectx.nodeEntityMap[v.DstNodeName]
		if dstNodeEntity == nil || dstNodeEntity.TagID != expectedDstTagID {
			logs.ErrorContextf(ctx, "EditNode: destination node name '%s' with tag_id %v not found", v.DstNodeName, expectedDstTagID)
			res.Code = errcode.ErrCode_BadRequest
			res.Message = "kecore_dst_node_not_found"
			return nil, res, nil
		}
		dstTagID := dstNodeEntity.TagID
		dstID := dstNodeEntity.ID

		// 检查边的 tag 是否存在
		edgeTag := ectx.edgeTagMap[v.EdgeName]
		needCreateTag := edgeTag == nil
		var edgeTagID uint
		if needCreateTag {
			edgeTag = &foresttype.GraphTag{
				GraphID:        ectx.graphInfo.ID,
				GraphVersionID: ectx.graphInfo.VersionID,
				Uin:            ectx.graphInfo.Uin,
				CompanyID:      ectx.graphInfo.CompanyID,
				TagType:        foresttype.TagTypeEdge,
				TagName:        v.EdgeName,
				TagStatus:      foresttype.TagStatusNot,
				Properties:     v.Properties,
			}
		} else {
			edgeTagID = edgeTag.ID
		}

		// 收集 tag relation pair
		tagRelationPairs = append(tagRelationPairs, [2]uint{srcTagID, dstTagID})

		edge := &foresttype.GraphEdge{
			GraphID:          ectx.graphInfo.ID,
			GraphVersionID:   ectx.graphInfo.VersionID,
			SrcID:            srcID,
			DstID:            dstID,
			SrcTagID:         srcTagID,
			DstTagID:         dstTagID,
			PropertiesValues: v.PropertiesValues,
		}

		edgesToCreate = append(edgesToCreate, &edgeToCreate{
			edgeTag:       edgeTag,
			edge:          edge,
			srcNodeName:   v.SrcNodeName,
			dstNodeName:   v.DstNodeName,
			edgeTagID:     edgeTagID,
			needCreateTag: needCreateTag,
			needCreateRel: false, // 稍后设置
			srcTagID:      srcTagID,
			dstTagID:      dstTagID,
		})
	}

	// 批量查询已存在的 tag relations
	if len(tagRelationPairs) > 0 {
		existingRelations, err := graph.GetExistingTagRelations(ctx, ectx.graphInfo.ID, ectx.graphInfo.VersionID, tagRelationPairs)
		if err != nil {
			logs.ErrorContextf(ctx, "EditNode.GetExistingTagRelations error: %v", err)
			res.Code = errcode.ErrCode_BadRequest
			res.Message = "kecore_get_tag_relation_failed"
			return nil, res, nil
		}

		// 设置 needCreateRel
		for i, etc := range edgesToCreate {
			key := graph.TagRelationKey(tagRelationPairs[i][0], tagRelationPairs[i][1])
			etc.needCreateRel = !existingRelations[key]
		}
	}

	return edgesToCreate, res, nil
}

// comparePropertiesValues 比较两个 PropertiesValues 是否相等（使用 JSON 序列化进行深度比较）
func comparePropertiesValues(a, b foresttype.PropertiesValues) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	// 使用 JSON 序列化进行深度比较
	aJSON, err1 := json.Marshal(a)
	bJSON, err2 := json.Marshal(b)
	if err1 != nil || err2 != nil {
		// 如果序列化失败，使用简单比较
		return false
	}
	return string(aJSON) == string(bJSON)
}

// compareProperties 比较两个 Properties 是否相等（使用 JSON 序列化进行深度比较）
func compareProperties(a, b foresttype.Properties) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	// 使用 JSON 序列化进行深度比较
	aJSON, err1 := json.Marshal(a)
	bJSON, err2 := json.Marshal(b)
	if err1 != nil || err2 != nil {
		// 如果序列化失败，使用简单比较
		return false
	}
	return string(aJSON) == string(bJSON)
}

// executeEditNodeTransaction 执行编辑节点的事务操作
func executeEditNodeTransaction(ctx *gin.Context,
	req *dtographnode.EditNodeRequest,
	ectx *editNodeContext, edgesToDelete []*foresttype.EdgeInfo,
	edgesToCreate []*edgeToCreate) error {
	// 在外层判断是否需要更新节点和 tag
	propertiesValuesChanged := !comparePropertiesValues(ectx.originalPropertiesValues, ectx.oldTNode.PropertiesValues)
	tagPropertiesChanged := !compareProperties(ectx.originalTagProperties, ectx.tag.Properties)

	needUpdateNode := propertiesValuesChanged
	needUpdateTag := tagPropertiesChanged
	needRecreateNodeInNebula := propertiesValuesChanged || tagPropertiesChanged

	return dbutil.Knownow().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 更新节点（如果节点名称或属性值有变化）
		if needUpdateNode {
			if err := tx.Save(ectx.oldTNode).Error; err != nil {
				logs.ErrorContextf(ctx, "EditNode.Save(oldTNode:%v) error: %v", ectx.oldTNode, err)
				return err
			}
		}

		// 保存更新的 tag（如果 tag 属性有变化）
		if needUpdateTag {
			if err := tx.Save(ectx.tag).Error; err != nil {
				logs.ErrorContextf(ctx, "EditNode.Save(tag:%v) error: %v", ectx.tag, err)
				return err
			}
		}

		// Nebula 操作：只有在有变化时才执行
		if needRecreateNodeInNebula {
			// // 删除旧节点（只有节点名称变化时才需要删除）
			// if nodeNameChanged {
			// 	if err := ectx.cli.DeleteNode(req.Request.OldNodeName); err != nil {
			// 		logs.WarnContextf(ctx, "EditNode.DeleteNode(oldNodeName:%s) error: %v", req.Request.OldNodeName, err)
			// 		// 继续执行，不返回错误
			// 	}
			// }

			// alter tag（如果 tag 属性有变化）
			if needUpdateTag {
				if err := ectx.cli.AlterTag(ectx.tag, true); err != nil {
					logs.ErrorContextf(ctx, "EditNode.AlterTag(tag:%v) error: %v", ectx.tag, err)
					return err
				}
			}

			// 插入/更新节点到 Nebula（如果属性值有变化）
			// EditNode 不再负责重命名，使用 OldNodeName（即原节点名）
			if needUpdateNode {
				if err := ectx.cli.InsertNode(&foresttype.TagNodeInfo{
					GraphID:          ectx.graphInfo.ID,
					GraphVersionID:   ectx.graphInfo.VersionID,
					Uin:              ectx.graphInfo.Uin,
					CompanyID:        ectx.graphInfo.CompanyID,
					TagID:            ectx.tag.ID,
					TagName:          ectx.tag.TagName,
					Properties:       ectx.tag.Properties,
					PropertiesValues: ectx.oldTNode.PropertiesValues,
					Name:             req.Request.OldNodeName,
				}); err != nil {
					logs.ErrorContextf(ctx, "EditNode.InsertNode(tag:%v) error: %v", ectx.tag, err)
					return err
				}
			}
		}

		// 删除不存在的边
		if len(edgesToDelete) > 0 {
			edgeIDsToDelete := make([]uint, 0, len(edgesToDelete))
			for _, edge := range edgesToDelete {
				edgeIDsToDelete = append(edgeIDsToDelete, edge.EdgeID)
			}
			if err := graph.DeleteEdgeByIDs(ctx, tx, edgeIDsToDelete); err != nil {
				logs.ErrorContextf(ctx, "EditNode.DeleteEdgeByIDs(edgeIDs:%v) error: %v", edgeIDsToDelete, err)
				return err
			}

			// 删除 Nebula 中的边
			for _, edge := range edgesToDelete {
				if err := ectx.cli.DeleteEdge(edge.EdgeTypeName, edge.SrcNodeName, edge.DstNodeName); err != nil {
					logs.WarnContextf(ctx, "EditNode.DeleteEdge(edgeType:%s, src:%s, dst:%s) error: %v", edge.EdgeTypeName, edge.SrcNodeName, edge.DstNodeName, err)
					// 继续执行，不返回错误
				}
			}
		}

		// 创建新的边
		if len(edgesToCreate) > 0 {
			// 创建边的 tag（如果需要）
			for _, etc := range edgesToCreate {
				if etc.needCreateTag {
					if err := tx.Save(etc.edgeTag).Error; err != nil {
						logs.ErrorContextf(ctx, "EditNode.Save(edgeTag:%v) error: %v", etc.edgeTag, err)
						return err
					}
					etc.edgeTagID = etc.edgeTag.ID

					// 检查 NebulaGraph 中边类型是否存在
					if !ectx.cli.EdgeTypeExists(etc.edgeTag.TagName) {
						// 在 NebulaGraph 中创建边类型
						if err := ectx.cli.CreateGraphTag(tx, etc.edgeTag); err != nil {
							logs.ErrorContextf(ctx, "EditNode.CreateGraphTag(edgeTag:%v) error: %v", etc.edgeTag, err)
							return err
						}
					}
				}
			}

			// 创建 tag relation（如果需要）
			for _, etc := range edgesToCreate {
				if etc.needCreateRel {
					edgeTagRelation := &foresttype.GraphEdgeTag{
						GraphID:        ectx.graphInfo.ID,
						GraphVersionID: ectx.graphInfo.VersionID,
						EdgeTypeID:     etc.edgeTagID,
						SrcTagID:       etc.srcTagID,
						DstTagID:       etc.dstTagID,
					}
					if err := tx.Save(edgeTagRelation).Error; err != nil {
						logs.ErrorContextf(ctx, "EditNode.Save(edgeTagRelation:%v) error: %v", edgeTagRelation, err)
						return err
					}
				}
			}

			// 设置边的 TagID
			for _, etc := range edgesToCreate {
				etc.edge.TagID = etc.edgeTagID
			}

			// 保存边
			edgesToSave := make([]*foresttype.GraphEdge, 0, len(edgesToCreate))
			for _, etc := range edgesToCreate {
				edgesToSave = append(edgesToSave, etc.edge)
			}
			if err := tx.CreateInBatches(edgesToSave, 50).Error; err != nil {
				logs.ErrorContextf(ctx, "EditNode.CreateInBatches(edges:%v) error: %v", edgesToSave, err)
				return err
			}

			// 插入新边到 Nebula
			edgeInfos := make([]*foresttype.EdgeInfo, 0, len(edgesToCreate))
			for _, etc := range edgesToCreate {
				edgeInfo := &foresttype.EdgeInfo{
					GraphID:          ectx.graphInfo.ID,
					GraphVersionID:   ectx.graphInfo.VersionID,
					SrcID:            etc.edge.SrcID,
					DstID:            etc.edge.DstID,
					TagID:            etc.edgeTagID,
					SrcTagID:         etc.srcTagID,
					DstTagID:         etc.dstTagID,
					EdgeTypeName:     etc.edgeTag.TagName,
					SrcNodeName:      etc.srcNodeName,
					DstNodeName:      etc.dstNodeName,
					PropertiesValues: etc.edge.PropertiesValues,
				}
				edgeInfos = append(edgeInfos, edgeInfo)
			}

			// 批量插入新边
			if err := ectx.cli.InsertEdges(edgeInfos); err != nil {
				logs.ErrorContextf(ctx, "EditNode.InsertEdges(edges:%v) error: %v", edgeInfos, err)
				return err
			}
		}

		return nil
	})
}
