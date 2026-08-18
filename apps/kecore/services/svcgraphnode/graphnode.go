package svcgraphnode

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtographnode"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/graph"
	"github.com/insmtx/corekg/apps/kecore/models/nebulagraph"
	"github.com/insmtx/corekg/pkgs/utils"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

// CreateNode 创建节点
func CreateNode(ctx *gin.Context, req *dtographnode.CreateNodeRequest) (res *dtographnode.CreateNodeResponse, err error) {
	res = &dtographnode.CreateNodeResponse{}
	// Get graph
	graphInfo, err := graph.GetGraph(ctx, req.Request.GraphID)
	if err != nil {
		logs.ErrorContextf(ctx, "CreateNode.GetGrap(id:%v) error: %v", req.Request.GraphID, err)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_graph_not_found"
		return res, nil
	}

	//check node's name exist
	if graph.ExistNodeName(ctx, graphInfo.ID, graphInfo.VersionID, req.Request.NodeName, req.Request.Tags[0].TagID) {
		logs.ErrorContextf(ctx, "CreateNode.ExistNodeName(graphID:%v, versionID:%v, nodeName:%s) already exist", graphInfo.ID, graphInfo.VersionID, req.Request.NodeName)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_node_name_exists"
		return res, nil
	}

	//Get Tag
	tag, err := graph.GetTagByID(ctx, req.Request.Tags[0].TagID)
	if err != nil {
		logs.ErrorContextf(ctx, "CreateNode.GetTag(id:%v) error: %v", req.Request.Tags[0].TagID, err)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_tag_not_found"
		return res, nil
	}

	// replace tag with request's tag properties
	tag.Properties = req.Request.Tags[0].Properties
	cli, err := nebulagraph.NewNebulaCLI(ctx, graphInfo.SpaceName)
	if err != nil {
		logs.ErrorContextf(ctx, "CreateNode.NewNebulaCLI(spaceName:%s) error: %v", graphInfo.SpaceName, err)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_create_graph_cli_failed"
		return res, nil
	}
	defer cli.Release()

	//Create tag node
	tNode := &foresttype.GraphTagNode{
		TagID:            tag.ID,
		GraphID:          graphInfo.ID,
		GraphVersionID:   graphInfo.VersionID,
		Name:             req.Request.NodeName,
		PropertiesValues: req.Request.Tags[0].PropertiesValues,
		Uin:              graphInfo.Uin,
		CompanyID:        graphInfo.CompanyID,
		CreatedType:      foresttype.CreatedTypeManual,
	}

	// 收集所有涉及的节点名称和对应的 tag_id（SrcNodeName 和 DstNodeName，排除当前节点名）
	// 使用 map[nodeName]tagID 来唯一标识节点
	nodeTagMap := make(map[string]uint)
	edgeNamesSet := make(map[string]bool)
	for _, v := range req.Request.Edges {
		// 源节点：如果是当前节点，跳过（使用当前节点的 tag）
		if v.SrcNodeName != "" && v.SrcNodeName != req.Request.NodeName {
			// 优先使用请求中的 SrcTagID，如果没有则无法确定 tag_id
			if v.SrcTagID == 0 {
				logs.ErrorContextf(ctx, "CreateNode: source node '%s' missing src_tag_id", v.SrcNodeName)
				res.Code = errcode.ErrCode_BadRequest
				res.Message = "kecore_src_node_missing_tag_id"
				return res, nil
			}
			nodeTagMap[v.SrcNodeName] = v.SrcTagID
		}
		// 目标节点：如果是当前节点，跳过（使用当前节点的 tag）
		if v.DstNodeName != "" && v.DstNodeName != req.Request.NodeName {
			// 优先使用请求中的 DstTagID，如果没有则无法确定 tag_id
			if v.DstTagID == 0 {
				logs.ErrorContextf(ctx, "CreateNode: destination node '%s' missing dst_tag_id", v.DstNodeName)
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

	edgeNames := make([]string, 0, len(edgeNamesSet))
	for name := range edgeNamesSet {
		edgeNames = append(edgeNames, name)
	}

	// 批量获取节点实体映射（使用 tag_id + name 查询）
	nodeEntityMap := make(map[string]*foresttype.GraphTagNode)
	if len(nodeTagMap) > 0 {
		var err error
		nodeEntityMap, err = graph.GetNodeEntityMapByNodeNames(ctx, graphInfo.ID, graphInfo.VersionID, nodeTagMap)
		if err != nil {
			logs.ErrorContextf(ctx, "CreateNode.GetNodeEntityMapByNodeNames(graphID:%v, versionID:%v, nodeTagMap:%v) error: %v", graphInfo.ID, graphInfo.VersionID, nodeTagMap, err)
			res.Code = errcode.ErrCode_BadRequest
			res.Message = "kecore_get_node_id_failed"
			return res, nil
		}
	}

	// 批量获取边 tag 映射
	edgeTagMap := make(map[string]*foresttype.GraphTag)
	if len(edgeNames) > 0 {
		var err error
		edgeTagMap, err = graph.GetEdgesByNames(ctx, graphInfo.ID, graphInfo.VersionID, edgeNames)
		if err != nil {
			logs.ErrorContextf(ctx, "CreateNode.GetEdgesByNames(graphID:%v, versionID:%v, edgeNames:%v) error: %v", graphInfo.ID, graphInfo.VersionID, edgeNames, err)
			res.Code = errcode.ErrCode_BadRequest
			res.Message = "kecore_get_edge_type_failed"
			return res, nil
		}
	}

	// 收集所有需要检查的 tag relation pairs
	tagRelationPairs := make([][2]uint, 0, len(req.Request.Edges))

	// 第一次遍历：验证节点存在性，收集 tag relation pairs
	edgesToCreate := make([]*edgeToCreate, 0, len(req.Request.Edges))
	for _, v := range req.Request.Edges {
		// 确定源节点和目标节点
		var srcNodeEntity, dstNodeEntity *foresttype.GraphTagNode
		var srcTagID, dstTagID uint

		// 源节点：如果是当前节点，使用当前节点的 tag
		if v.SrcNodeName == req.Request.NodeName {
			srcTagID = tag.ID // 当前节点的 tag（节点创建后会有 ID）
		} else {
			// 使用 tag_id + name 查询节点
			expectedSrcTagID := v.SrcTagID
			if expectedSrcTagID == 0 {
				logs.ErrorContextf(ctx, "CreateNode: source node '%s' missing src_tag_id", v.SrcNodeName)
				res.Code = errcode.ErrCode_BadRequest
				res.Message = "kecore_src_node_missing_tag_id"
				return res, nil
			}
			srcNodeEntity = nodeEntityMap[v.SrcNodeName]
			if srcNodeEntity == nil || srcNodeEntity.TagID != expectedSrcTagID {
				logs.ErrorContextf(ctx, "CreateNode: source node name '%s' with tag_id %v not found", v.SrcNodeName, expectedSrcTagID)
				res.Code = errcode.ErrCode_BadRequest
				res.Message = "kecore_src_node_not_found"
				return res, nil
			}
			srcTagID = srcNodeEntity.TagID
		}

		// 目标节点：如果是当前节点，使用当前节点的 tag
		if v.DstNodeName == req.Request.NodeName {
			dstTagID = tag.ID // 当前节点的 tag
		} else {
			// 使用 tag_id + name 查询节点
			expectedDstTagID := v.DstTagID
			if expectedDstTagID == 0 {
				logs.ErrorContextf(ctx, "CreateNode: destination node '%s' missing dst_tag_id", v.DstNodeName)
				res.Code = errcode.ErrCode_BadRequest
				res.Message = "kecore_dst_node_missing_tag_id"
				return res, nil
			}
			dstNodeEntity = nodeEntityMap[v.DstNodeName]
			if dstNodeEntity == nil || dstNodeEntity.TagID != expectedDstTagID {
				logs.ErrorContextf(ctx, "CreateNode: destination node name '%s' with tag_id %v not found", v.DstNodeName, expectedDstTagID)
				res.Code = errcode.ErrCode_BadRequest
				res.Message = "kecore_dst_node_not_found"
				return res, nil
			}
			dstTagID = dstNodeEntity.TagID
		}

		// 检查边的 tag 是否存在
		edgeTag := edgeTagMap[v.EdgeName]
		needCreateTag := edgeTag == nil
		var edgeTagID uint
		if needCreateTag {
			edgeTag = &foresttype.GraphTag{
				GraphID:        graphInfo.ID,
				GraphVersionID: graphInfo.VersionID,
				Uin:            graphInfo.Uin,
				CompanyID:      graphInfo.CompanyID,
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

		// 获取节点 ID
		var srcID, dstID uint
		if v.SrcNodeName == req.Request.NodeName {
			srcID = 0 // 稍后设置
		} else {
			srcID = srcNodeEntity.ID
		}
		if v.DstNodeName == req.Request.NodeName {
			dstID = 0 // 稍后设置
		} else {
			dstID = dstNodeEntity.ID
		}

		edge := &foresttype.GraphEdge{
			GraphID:          graphInfo.ID,
			GraphVersionID:   graphInfo.VersionID,
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
	existingRelations, err := graph.GetExistingTagRelations(ctx, graphInfo.ID, graphInfo.VersionID, tagRelationPairs)
	if err != nil {
		logs.ErrorContextf(ctx, "CreateNode.GetExistingTagRelations error: %v", err)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_get_tag_relation_failed"
		return res, nil
	}

	// 设置 needCreateRel
	for i, etc := range edgesToCreate {
		key := graph.TagRelationKey(tagRelationPairs[i][0], tagRelationPairs[i][1])
		etc.needCreateRel = !existingRelations[key]
	}
	if err := dbutil.Knownow().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// ? mysql action
		// save node
		if err := tx.Save(tNode).Error; err != nil {
			logs.ErrorContextf(ctx, "CreateNode.Save(tNode:%v) error: %v", tNode, err)
			return err
		}

		// save updated tag
		if err := tx.Save(tag).Error; err != nil {
			logs.ErrorContextf(ctx, "CreateNode.Save(tag:%v) error: %v", tag, err)
			return err
		}

		// ? nebula action
		// alter tag
		if err := cli.AlterTag(tag, true); err != nil {
			logs.ErrorContextf(ctx, "CreateNode.AlterTag(tag:%v) error: %v", tag, err)
			return err
		}

		// insert node
		if err := cli.InsertNode(&foresttype.TagNodeInfo{
			GraphID:          graphInfo.ID,
			GraphVersionID:   graphInfo.VersionID,
			Uin:              graphInfo.Uin,
			CompanyID:        graphInfo.CompanyID,
			TagID:            tag.ID,
			TagName:          tag.TagName,
			Properties:       tag.Properties,
			PropertiesValues: tNode.PropertiesValues,
			Name:             req.Request.NodeName,
		}); err != nil {
			logs.ErrorContextf(ctx, "CreateNode.InsertNode(tag:%v) error: %v", tag, err)
			return err
		}

		// ? if edges exist, save edges
		if len(edgesToCreate) > 0 {
			// 创建边的 tag（如果需要）
			for _, etc := range edgesToCreate {
				if etc.needCreateTag {
					if err := tx.Save(etc.edgeTag).Error; err != nil {
						logs.ErrorContextf(ctx, "CreateNode.Save(edgeTag:%v) error: %v", etc.edgeTag, err)
						return err
					}
					etc.edgeTagID = etc.edgeTag.ID

					// 检查 NebulaGraph 中边类型是否存在
					if !cli.EdgeTypeExists(etc.edgeTag.TagName) {
						// 在 NebulaGraph 中创建边类型
						if err := cli.CreateGraphTag(tx, etc.edgeTag); err != nil {
							logs.ErrorContextf(ctx, "CreateNode.CreateGraphTag(edgeTag:%v) error: %v", etc.edgeTag, err)
							return err
						}
					}
				}
			}

			// 创建 tag relation（如果需要）
			for _, etc := range edgesToCreate {
				if etc.needCreateRel {
					edgeTagRelation := &foresttype.GraphEdgeTag{
						GraphID:        graphInfo.ID,
						GraphVersionID: graphInfo.VersionID,
						EdgeTypeID:     etc.edgeTagID,
						SrcTagID:       etc.srcTagID,
						DstTagID:       etc.dstTagID,
					}
					if err := tx.Save(edgeTagRelation).Error; err != nil {
						logs.ErrorContextf(ctx, "CreateNode.Save(edgeTagRelation:%v) error: %v", edgeTagRelation, err)
						return err
					}
				}
			}

			// 设置边的 SrcID、DstID 和 TagID（处理当前节点是 src 或 dst 的情况）
			for _, etc := range edgesToCreate {
				if etc.srcNodeName == req.Request.NodeName {
					etc.edge.SrcID = tNode.ID
				}
				if etc.dstNodeName == req.Request.NodeName {
					etc.edge.DstID = tNode.ID
				}
				etc.edge.TagID = etc.edgeTagID
			}

			// 保存边
			edgesToSave := make([]*foresttype.GraphEdge, 0, len(edgesToCreate))
			for _, etc := range edgesToCreate {
				edgesToSave = append(edgesToSave, etc.edge)
			}
			if err := tx.CreateInBatches(edgesToSave, 50).Error; err != nil {
				logs.ErrorContextf(ctx, "CreateNode.CreateInBatches(edges:%v) error: %v", edgesToSave, err)
				return err
			}

			// ? nebula action
			// 插入边到 Nebula
			edgeInfos := make([]*foresttype.EdgeInfo, 0, len(edgesToCreate))
			for _, etc := range edgesToCreate {
				edgeInfo := &foresttype.EdgeInfo{
					GraphID:          graphInfo.ID,
					GraphVersionID:   graphInfo.VersionID,
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

			// 批量插入边
			if err := cli.InsertEdges(edgeInfos); err != nil {
				logs.ErrorContextf(ctx, "CreateNode.InsertEdges(edges:%v) error: %v", edgeInfos, err)
				return err
			}
		}

		return nil
	}); err != nil {
		logs.ErrorContextf(ctx, "CreateNode.Transaction(tag:%v) error: %v", tag, err)
		return res, err
	}
	res.Response.NodeID = tNode.ID
	return res, nil
}

// GetNodeEdges 获取节点边
func GetNodeEdges(ctx *gin.Context, req *dtographnode.GetNodeEdgesRequest) (res *dtographnode.GetNodeEdgesResponse, err error) {
	res = &dtographnode.GetNodeEdgesResponse{}

	graphInfo, err := graph.GetGraph(ctx, req.Request.GraphID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetNodeEdges.GetGraph(graphID:%v) error: %v", req.Request.GraphID, err)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_graph_not_found"
		return res, nil
	}

	edgeInfos, err := graph.GetEdgesBySrcNodeName(ctx, graphInfo.ID, graphInfo.VersionID, req.Request.NodeName, req.Request.TagID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetNodeEdges.GetEdgesBySrcNodeName(graphID:%v, versionID:%v, nodeName:%s, tagID:%v) error: %v", graphInfo.ID, graphInfo.VersionID, req.Request.NodeName, req.Request.TagID, err)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_get_node_edges_failed"
		return res, nil
	}
	res.Response.Edges = make([]dtographnode.EdgeObject, 0, len(edgeInfos))
	for _, v := range edgeInfos {
		res.Response.Edges = append(res.Response.Edges, dtographnode.EdgeObject{
			EdgeID:           v.EdgeID,
			FileIDList:       v.FileIDList,
			ChunkIDList:      v.ChunkIDList,
			Properties:       v.Properties,
			PropertiesValues: v.PropertiesValues,

			// relationship
			EdgeName:    v.EdgeTypeName,
			SrcNodeID:   v.SrcID,
			SrcNodeName: v.SrcNodeName,
			SrcTagID:    v.SrcTagID,
			DstNodeID:   v.DstID,
			DstNodeName: v.DstNodeName,
			DstTagID:    v.DstTagID,
		})
	}

	return res, nil
}

// EditNode 编辑节点
func EditNode(ctx *gin.Context, req *dtographnode.EditNodeRequest) (res *dtographnode.EditNodeResponse, err error) {
	res = &dtographnode.EditNodeResponse{}

	// 1. 验证请求并初始化上下文
	ectx, res, err := validateEditNodeRequest(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "EditNode.validateEditNodeRequest res value: %v, error: %v", res, err)
		return res, err
	}
	defer ectx.cli.Release()

	// 2. 准备节点和 tag 数据
	prepareNodeAndTag(req, ectx)

	// 3. 收集边相关的节点和 tag 数据
	res, err = collectEdgeRelatedData(ctx, req, ectx)
	if err != nil {
		logs.ErrorContextf(ctx, "EditNode.collectEdgeRelatedData res value: %v, error: %v", res, err)
		return res, err
	}

	// 4. 计算需要删除的边
	edgesToDelete, res, err := calculateEdgeDiff(ctx, req, ectx)
	if err != nil {
		logs.ErrorContextf(ctx, "EditNode.calculateEdgeDiff res value: %v, error: %v", res, err)
		return res, err
	}

	// 5. 准备要创建的边
	edgesToCreate, res, err := prepareEdgesToCreate(ctx, req, ectx)
	if err != nil {
		logs.ErrorContextf(ctx, "EditNode.prepareEdgesToCreate res value: %v, error: %v", res, err)
		return res, err
	}

	// 6. 执行事务操作
	if err := executeEditNodeTransaction(ctx, req, ectx, edgesToDelete, edgesToCreate); err != nil {
		logs.ErrorContextf(ctx, "EditNode.executeEditNodeTransaction error: %v", err)
		return res, err
	}

	return res, nil
}

// CreateNodeEdge 创建节点边
func CreateNodeEdge(ctx *gin.Context, req *dtographnode.CreateNodeEdgeRequest) (res *dtographnode.CreateNodeEdgeResponse, err error) {
	res = &dtographnode.CreateNodeEdgeResponse{}

	graphInfo, err := graph.GetGraph(ctx, req.Request.GraphID)
	if err != nil {
		logs.WarnContextf(ctx, "CreateNodeEdge.GetGraph(graphID:%v) error: %v", req.Request.GraphID, err)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_graph_not_found"
		return res, nil
	}

	newTag := false
	edgeTag, err := graph.GetEdgeByName(ctx, graphInfo.ID, graphInfo.VersionID, req.Request.Edge.EdgeName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// if not exist create it
			edgeTag = &foresttype.GraphTag{
				Uin:            graphInfo.Uin,
				CompanyID:      graphInfo.CompanyID,
				GraphID:        graphInfo.ID,
				GraphVersionID: graphInfo.VersionID,
				TagName:        req.Request.Edge.EdgeName,
				TagType:        foresttype.TagTypeEdge,
				TagStatus:      foresttype.TagStatusNot,
				Properties:     req.Request.Edge.Properties,
			}
			newTag = true
		} else {
			logs.WarnContextf(ctx, "CreateNodeEdge.GetEdgeByName(graphID:%v, versionID:%v, edgeName:%s) error: %v", graphInfo.ID, graphInfo.VersionID, req.Request.Edge.EdgeName, err)
			res.Code = errcode.ErrCode_BadRequest
			res.Message = "kecore_get_edge_name_failed"
			return res, nil
		}
	}

	// 构建 nodeTagMap，使用 tag_id + name 查询节点
	nodeTagMap := make(map[string]uint)
	if req.Request.Edge.SrcTagID == 0 {
		logs.WarnContextf(ctx, "CreateNodeEdge: source node '%s' missing src_tag_id", req.Request.Edge.SrcNodeName)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_src_node_missing_tag_id"
		return res, nil
	}
	if req.Request.Edge.DstTagID == 0 {
		logs.WarnContextf(ctx, "CreateNodeEdge: destination node '%s' missing dst_tag_id", req.Request.Edge.DstNodeName)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_dst_node_missing_tag_id"
		return res, nil
	}
	nodeTagMap[req.Request.Edge.SrcNodeName] = req.Request.Edge.SrcTagID
	nodeTagMap[req.Request.Edge.DstNodeName] = req.Request.Edge.DstTagID

	nodeEntityMap, err := graph.GetNodeEntityMapByNodeNames(ctx, graphInfo.ID, graphInfo.VersionID, nodeTagMap)
	if err != nil {
		logs.WarnContextf(ctx, "CreateNodeEdge.GetNodeEntityMapByNodeNames(graphID:%v, versionID:%v, nodeTagMap:%v) error: %v", graphInfo.ID, graphInfo.VersionID, nodeTagMap, err)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "获取节点ID失败"
		return res, nil
	}

	srcNode := nodeEntityMap[req.Request.Edge.SrcNodeName]
	dstNode := nodeEntityMap[req.Request.Edge.DstNodeName]
	if srcNode == nil || srcNode.TagID != req.Request.Edge.SrcTagID {
		logs.WarnContextf(ctx, "CreateNodeEdge: source node '%s' with tag_id %v not found", req.Request.Edge.SrcNodeName, req.Request.Edge.SrcTagID)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_src_node_not_found"
		return res, nil
	}
	if dstNode == nil || dstNode.TagID != req.Request.Edge.DstTagID {
		logs.WarnContextf(ctx, "CreateNodeEdge: destination node '%s' with tag_id %v not found", req.Request.Edge.DstNodeName, req.Request.Edge.DstTagID)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_dst_node_not_found"
		return res, nil
	}

	if graph.AlreadyHasEdgeByTagID(ctx, graphInfo.ID, graphInfo.VersionID,
		edgeTag.ID, srcNode.ID, dstNode.ID) {
		logs.WarnContextf(ctx, "CreateNodeEdge.AlreadyHasEdge(graphID:%v, versionID:%v, srcNodeName:%s, dstNodeName:%s) already exist", graphInfo.ID, graphInfo.VersionID, req.Request.Edge.SrcNodeName, req.Request.Edge.DstNodeName)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_edge_exists"
		return res, nil
	}

	var (
		edge = &foresttype.GraphEdge{
			GraphID:          graphInfo.ID,
			GraphVersionID:   graphInfo.VersionID,
			TagID:            edgeTag.ID,
			DstID:            dstNode.ID,
			SrcID:            srcNode.ID,
			DstTagID:         dstNode.TagID,
			SrcTagID:         srcNode.TagID,
			PropertiesValues: req.Request.Edge.PropertiesValues,
			FileIDList:       req.Request.Edge.FileIDList,
			ChunkIDList:      req.Request.Edge.ChunkIDList,
		}
		edgeTagRelation *foresttype.GraphEdgeTag
	)
	if !graph.AlreadyHasTagRelation(ctx, graphInfo.ID, graphInfo.VersionID, srcNode.TagID, dstNode.TagID) {
		// create tag relation
		edgeTagRelation = &foresttype.GraphEdgeTag{
			GraphID:        graphInfo.ID,
			GraphVersionID: graphInfo.VersionID,
			EdgeTypeID:     edgeTag.ID,
			SrcTagID:       srcNode.TagID,
			DstTagID:       dstNode.TagID,
		}
	}
	cli, err := nebulagraph.NewNebulaCLI(ctx, graphInfo.SpaceName)
	if err != nil {
		logs.ErrorContextf(ctx, "CreateNodeEdge.NewNebulaCLI(spaceName:%s) error: %v", graphInfo.SpaceName, err)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_create_graph_cli_failed"
		return res, nil
	}
	defer cli.Release()

	if err := dbutil.Knownow().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// ? mysql action
		if newTag {
			if err := tx.Save(edgeTag).Error; err != nil {
				logs.ErrorContextf(ctx, "CreateNodeEdge.Save(edgeTag:%v) error: %v", edgeTag, err)
				return err
			}
			edge.TagID = edgeTag.ID
			if edgeTagRelation != nil {
				edgeTagRelation.EdgeTypeID = edgeTag.ID
			}

			// 检查 NebulaGraph 中边类型是否存在
			if !cli.EdgeTypeExists(edgeTag.TagName) {
				// 在 NebulaGraph 中创建边类型
				if err := cli.CreateGraphTag(tx, edgeTag); err != nil {
					logs.ErrorContextf(ctx, "CreateNodeEdge.CreateGraphTag(edgeTag:%v) error: %v", edgeTag, err)
					return err
				}
			}
		}

		// 保存边
		if err := tx.Save(edge).Error; err != nil {
			logs.ErrorContextf(ctx, "CreateNodeEdge.Save(edge:%v) error: %v", edge, err)
			return err
		}

		// 如果不存在 tag relation，则创建
		if edgeTagRelation != nil {
			if err := tx.Save(edgeTagRelation).Error; err != nil {
				logs.ErrorContextf(ctx, "CreateNodeEdge.Save(edgeTagRelation:%v) error: %v", edgeTagRelation, err)
				return err
			}
		}

		// ? nebula action
		// 插入边到 Nebula
		if err := cli.InsertEdge(&foresttype.EdgeInfo{
			GraphID:          graphInfo.ID,
			GraphVersionID:   graphInfo.VersionID,
			SrcID:            srcNode.ID,
			DstID:            dstNode.ID,
			TagID:            edgeTag.ID,
			SrcTagID:         srcNode.TagID,
			DstTagID:         dstNode.TagID,
			EdgeTypeName:     edgeTag.TagName,
			SrcNodeName:      req.Request.Edge.SrcNodeName,
			DstNodeName:      req.Request.Edge.DstNodeName,
			PropertiesValues: edge.PropertiesValues,
			FileIDList:       edge.FileIDList,
			ChunkIDList:      edge.ChunkIDList,
		}); err != nil {
			logs.ErrorContextf(ctx, "CreateNodeEdge.InsertEdge(edge:%v) error: %v", edge, err)
			return err
		}

		return nil
	}); err != nil {
		logs.ErrorContextf(ctx, "CreateNodeEdge.Transaction error: %v", err)
		return res, err
	}

	return res, nil
}

// DeleteNode 删除节点
// 如果传入了TagID，则删除对应的tag节点；如果TagID为0，则删除该节点名称的所有tag节点
func DeleteNode(ctx *gin.Context, req *dtographnode.DeleteNodeRequest) (res *dtographnode.DeleteNodeResponse, err error) {
	res = &dtographnode.DeleteNodeResponse{}

	graphInfo, err := graph.GetGraph(ctx, req.Request.GraphID)
	if err != nil {
		logs.WarnContextf(ctx, "DeleteNode.GetGraph(graphID:%v) error: %v", req.Request.GraphID, err)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_graph_not_found"
		return res, nil
	}

	var nodesToDelete []*foresttype.GraphTagNode
	var edges []*foresttype.EdgeInfo

	// 如果传入了TagID，删除对应的tag节点；否则删除该节点名称的所有tag节点
	if req.Request.TagID > 0 {
		// 查询指定的tag节点
		node, err := graph.GetTNodeByName(ctx, graphInfo.ID, graphInfo.VersionID, req.Request.NodeName, req.Request.TagID)
		if err != nil {
			logs.WarnContextf(ctx, "DeleteNode.GetTNodeByName(graphID:%v, versionID:%v, nodeName:%s, tagID:%v) error: %v", graphInfo.ID, graphInfo.VersionID, req.Request.NodeName, req.Request.TagID, err)
			res.Code = errcode.ErrCode_BadRequest
			res.Message = "kecore_node_not_found"
			return res, nil
		}
		nodesToDelete = []*foresttype.GraphTagNode{node}

		// 获取该tag节点的所有边
		edges, err = graph.GetEdgesByNodeName(ctx, graphInfo.ID, graphInfo.VersionID, req.Request.NodeName, req.Request.TagID)
		if err != nil {
			logs.WarnContextf(ctx, "DeleteNode.GetEdgesByNodeName(graphID:%v, versionID:%v, nodeName:%s, tagID:%v) error: %v", graphInfo.ID, graphInfo.VersionID, req.Request.NodeName, req.Request.TagID, err)
			res.Code = errcode.ErrCode_BadRequest
			res.Message = "kecore_get_edge_failed"
			return res, nil
		}
	} else {
		// 查询该节点名称的所有tag节点
		nodesToDelete, err = graph.GetTNodesByName(ctx, graphInfo.ID, graphInfo.VersionID, req.Request.NodeName)
		if err != nil {
			logs.WarnContextf(ctx, "DeleteNode.GetTNodesByName(graphID:%v, versionID:%v, nodeName:%s) error: %v", graphInfo.ID, graphInfo.VersionID, req.Request.NodeName, err)
			res.Code = errcode.ErrCode_BadRequest
			res.Message = "kecore_get_node_failed"
			return res, nil
		}
		if len(nodesToDelete) == 0 {
			logs.WarnContextf(ctx, "DeleteNode.GetTNodesByName(graphID:%v, versionID:%v, nodeName:%s) no nodes found", graphInfo.ID, graphInfo.VersionID, req.Request.NodeName)
			res.Code = errcode.ErrCode_BadRequest
			res.Message = "kecore_node_not_found"
			return res, nil
		}

		// 获取该节点名称的所有边（不区分tagID）
		edges, err = graph.GetEdgesByNodeNameAll(ctx, graphInfo.ID, graphInfo.VersionID, req.Request.NodeName)
		if err != nil {
			logs.WarnContextf(ctx, "DeleteNode.GetEdgesByNodeNameAll(graphID:%v, versionID:%v, nodeName:%s) error: %v", graphInfo.ID, graphInfo.VersionID, req.Request.NodeName, err)
			res.Code = errcode.ErrCode_BadRequest
			res.Message = "kecore_get_edge_failed"
			return res, nil
		}
	}

	// check if the edge's tag is the last tag relation
	edgeTagIDs := utils.Map(edges, func(edge *foresttype.EdgeInfo) uint {
		return edge.TagID
	})
	edgeIDs := utils.Map(edges, func(edge *foresttype.EdgeInfo) uint {
		return edge.EdgeID
	})

	lastTagRelationEdgeTagIDs, err := graph.GetLastTagRelationEdgeTagIDs(ctx, graphInfo.ID, graphInfo.VersionID, edgeTagIDs)
	if err != nil {
		logs.WarnContextf(ctx, "DeleteNode.GetLastTagRelationEdgeTagIDs(graphID:%v, versionID:%v, edgeTagIDs:%v) error: %v", graphInfo.ID, graphInfo.VersionID, edgeTagIDs, err)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_get_last_tag_relation_edge_tag_ids_failed"
		return res, nil
	}

	cli, err := nebulagraph.NewNebulaCLI(ctx, graphInfo.SpaceName)
	if err != nil {
		logs.ErrorContextf(ctx, "DeleteNode.NewNebulaCLI(spaceName:%s) error: %v", graphInfo.SpaceName, err)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_create_graph_cli_failed"
		return res, nil
	}
	defer cli.Release()

	// 获取要删除的节点ID列表
	nodeIDs := utils.Map(nodesToDelete, func(node *foresttype.GraphTagNode) uint {
		return node.ID
	})

	if err := dbutil.Knownow().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// ? mysql action
		// delete nodes
		if len(nodeIDs) > 0 {
			if err := tx.
				Table(foresttype.GraphTagNode{}.TableName()).
				Where("id in ?", nodeIDs).
				Delete(&foresttype.GraphTagNode{}).Error; err != nil {
				logs.WarnContextf(ctx, "DeleteNode.Delete(nodeIDs:%v) error: %v", nodeIDs, err)
				return err
			}
		}

		if len(edgeIDs) > 0 {
			if err := tx.
				Table(foresttype.TableNameKeGraphEdge).
				Where("id in ?", edgeIDs).
				Delete(&foresttype.GraphEdge{}).Error; err != nil {
				logs.WarnContextf(ctx, "DeleteNode.Delete(edgeIDs:%v) error: %v", edgeIDs, err)
				return err
			}
		}

		if len(lastTagRelationEdgeTagIDs) > 0 {
			if err := tx.
				Table(foresttype.TableNameKeGraphEdgeTag).
				Where("id in ?", lastTagRelationEdgeTagIDs).
				Delete(&foresttype.GraphEdgeTag{}).Error; err != nil {
				logs.WarnContextf(ctx, "DeleteNode.Delete(lastTagRelationEdgeTagIDs:%v) error: %v", lastTagRelationEdgeTagIDs, err)
				return err
			}
		}

		// ? nebula action
		if req.Request.TagID > 0 {
			// 只删除指定的tag，使用 DELETE TAG tag_name FROM vertex_id
			// 获取tag名称
			tag, err := graph.GetTagByID(ctx, req.Request.TagID)
			if err != nil {
				logs.WarnContextf(ctx, "DeleteNode.GetTagByID(tagID:%v) error: %v", req.Request.TagID, err)
				return err
			}
			if err := cli.DeleteTag(req.Request.NodeName, tag.TagName); err != nil {
				logs.WarnContextf(ctx, "DeleteNode.DeleteTag(nodeName:%s, tagName:%s) error: %v", req.Request.NodeName, tag.TagName, err)
				return err
			}
		} else {
			// 删除整个节点的所有tag，使用 DELETE VERTEX
			if err := cli.DeleteNode(req.Request.NodeName); err != nil {
				logs.WarnContextf(ctx, "DeleteNode.DeleteNode(nodeName:%s) error: %v", req.Request.NodeName, err)
				return err
			}
		}

		return nil
	}); err != nil {
		logs.WarnContextf(ctx, "DeleteNode.Transaction(graphID:%v, versionID:%v, nodeName:%s) error: %v", graphInfo.ID, graphInfo.VersionID, req.Request.NodeName, err)
		return res, err
	}

	return res, nil
}

func GetNodeReference(ctx *gin.Context, req *dtographnode.GetNodeReferenceRequest) (res *dtographnode.GetNodeReferenceResponse, err error) {
	res = &dtographnode.GetNodeReferenceResponse{}

	graphInfo, err := graph.GetGraph(ctx, req.Request.GraphID)
	if err != nil {
		logs.WarnContextf(ctx, "GetNodeReference.GetGraph(graphID:%v) error: %v", req.Request.GraphID, err)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_graph_not_found"
		return res, nil
	}

	tnode, err := graph.GetTNodeByName(ctx, graphInfo.ID, graphInfo.VersionID, req.Request.NodeName, req.Request.TagID)
	if err != nil {
		logs.WarnContextf(ctx, "GetNodeReference.GetTNodeByName(graphID:%v, versionID:%v, nodeName:%s, tagID:%v) error: %v", graphInfo.ID, graphInfo.VersionID, req.Request.NodeName, req.Request.TagID, err)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "获取节点失败"
		return res, nil
	}

	tag, err := graph.GetTagByID(ctx, tnode.TagID)
	if err != nil {
		logs.WarnContextf(ctx, "GetNodeReference.GetTagByID(tagID:%v) error: %v", tnode.TagID, err)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_get_tag_failed"
		return res, nil
	}

	// fill tags
	res.Response.Tags = []dtographnode.TagObject{
		{
			TagID:            tnode.TagID,
			TagName:          tag.TagName,
			Properties:       tag.Properties,
			PropertiesValues: tnode.PropertiesValues,
		},
	}

	// 使用 tag_id + name 查询边
	edges, err := graph.GetEdgesByNodeName(ctx, graphInfo.ID, graphInfo.VersionID, req.Request.NodeName, req.Request.TagID)
	if err != nil {
		logs.WarnContextf(ctx, "GetNodeReference.GetEdgesByNodeName(graphID:%v, versionID:%v, nodeName:%s, tagID:%v) error: %v", graphInfo.ID, graphInfo.VersionID, req.Request.NodeName, req.Request.TagID, err)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "获取边失败"
		return res, nil
	}

	// fill edges
	res.Response.Edges = make([]dtographnode.EdgeObject, 0, len(edges))
	for _, v := range edges {
		res.Response.Edges = append(res.Response.Edges, dtographnode.EdgeObject{
			EdgeID:           v.EdgeID,
			SrcNodeID:        v.SrcID,
			SrcNodeName:      v.SrcNodeName,
			SrcTagID:         v.SrcTagID,
			DstNodeID:        v.DstID,
			DstNodeName:      v.DstNodeName,
			DstTagID:         v.DstTagID,
			EdgeName:         v.EdgeTypeName,
			FileIDList:       v.FileIDList,
			ChunkIDList:      v.ChunkIDList,
			Properties:       v.Properties,
			PropertiesValues: v.PropertiesValues,
		})
	}

	var chunkIDs []string
	var fileIDs []uint
	fileIDs = append(fileIDs, tnode.FileIDList.Slice()...)
	chunkIDs = append(chunkIDs, tnode.ChunkIDList.Slice()...)
	for _, v := range edges {
		chunkIDs = append(chunkIDs, v.ChunkIDList.Slice()...)
	}

	fIDs, err := graph.GetFileIDsByChunkIDs(ctx, chunkIDs)
	if err != nil {
		logs.WarnContextf(ctx, "GetNodeReference.GetFileIDsByChunkIDs(chunkIDs:%v) error: %v", chunkIDs, err)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_get_file_id_failed"
		return res, nil
	}
	fileIDs = append(fileIDs, fIDs...)
	files, err := forest.GetForestFileByIDs(fileIDs)
	if err != nil {
		logs.WarnContextf(ctx, "GetNodeReference.GetForestFileByIDs(fileIDs:%v) error: %v", fileIDs, err)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "获取文件失败"
		return res, nil
	}

	// fill files
	res.Response.Files = make([]dtographnode.FileObject, 0, len(files))
	for _, v := range files {
		res.Response.Files = append(res.Response.Files, dtographnode.FileObject{
			FileID:   v.ID,
			FileName: v.Name,
		})
	}

	return res, nil
}

func GetGraphEdges(ctx *gin.Context, req *dtographnode.GetGraphEdgesRequest) (res *dtographnode.GetGraphEdgesResponse, err error) {
	res = &dtographnode.GetGraphEdgesResponse{}

	graphInfo, err := graph.GetGraph(ctx, req.Request.GraphID)
	if err != nil {
		logs.WarnContextf(ctx, "GetGraphEdges.GetGraph(graphID:%v) error: %v", req.Request.GraphID, err)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_graph_not_found"
		return res, nil
	}

	edges, err := graph.GetEdgeByGraphID(ctx, graphInfo.ID, graphInfo.VersionID)
	if err != nil {
		logs.WarnContextf(ctx, "GetGraphEdges.GetEdgeByGraphID(graphID:%v, versionID:%v) error: %v", graphInfo.ID, graphInfo.VersionID, err)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "获取边失败"
		return res, nil
	}

	res.Response.Edges = make([]dtographnode.EdgeObject, 0, len(edges))
	for _, v := range edges {
		res.Response.Edges = append(res.Response.Edges, dtographnode.EdgeObject{
			EdgeID:     v.ID,
			EdgeName:   v.TagName,
			Properties: v.Properties,
		})
	}
	return res, nil
}

// RenameNode 重命名单个 tag 节点
func RenameNode(ctx *gin.Context, req *dtographnode.RenameNodeRequest) (res *dtographnode.RenameNodeResponse, err error) {
	res = &dtographnode.RenameNodeResponse{}

	// 1. 获取图谱信息
	graphInfo, err := graph.GetGraph(ctx, req.Request.GraphID)
	if err != nil {
		logs.WarnContextf(ctx, "RenameNode.GetGraph(graphID:%v) error: %v", req.Request.GraphID, err)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_graph_not_found"
		return res, nil
	}

	// 2. 检查旧节点是否存在
	oldNode, err := graph.GetTNodeByName(ctx, graphInfo.ID, graphInfo.VersionID, req.Request.OldNodeName, req.Request.TagID)
	if err != nil {
		logs.WarnContextf(ctx, "RenameNode.GetTNodeByName(graphID:%v, versionID:%v, nodeName:%s, tagID:%v) error: %v",
			graphInfo.ID, graphInfo.VersionID, req.Request.OldNodeName, req.Request.TagID, err)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_old_node_not_found"
		return res, nil
	}

	// 3. 检查新节点名称是否已存在（避免重名冲突）
	if graph.ExistNodeName(ctx, graphInfo.ID, graphInfo.VersionID, req.Request.NodeName, req.Request.TagID) {
		logs.WarnContextf(ctx, "RenameNode.ExistNodeName(graphID:%v, versionID:%v, nodeName:%s, tagID:%v) already exist",
			graphInfo.ID, graphInfo.VersionID, req.Request.NodeName, req.Request.TagID)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_new_node_name_already_exists"
		return res, nil
	}

	// 4. 获取 tag 信息
	tag, err := graph.GetTagByID(ctx, req.Request.TagID)
	if err != nil {
		logs.WarnContextf(ctx, "RenameNode.GetTagByID(tagID:%v) error: %v", req.Request.TagID, err)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_tag_not_found"
		return res, nil
	}

	// 5. 查询该节点的所有出入边
	edges, err := graph.GetEdgesByNodeName(ctx, graphInfo.ID, graphInfo.VersionID, req.Request.OldNodeName, req.Request.TagID)
	if err != nil {
		logs.WarnContextf(ctx, "RenameNode.GetEdgesByNodeName(graphID:%v, versionID:%v, nodeName:%s, tagID:%v) error: %v",
			graphInfo.ID, graphInfo.VersionID, req.Request.OldNodeName, req.Request.TagID, err)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_get_node_edges_failed"
		return res, nil
	}

	// 6. 创建 Nebula CLI
	cli, err := nebulagraph.NewNebulaCLI(ctx, graphInfo.SpaceName)
	if err != nil {
		logs.ErrorContextf(ctx, "RenameNode.NewNebulaCLI(spaceName:%s) error: %v", graphInfo.SpaceName, err)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_create_graph_cli_failed"
		return res, nil
	}
	defer cli.Release()

	// 7. 执行事务操作
	if err := dbutil.Knownow().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// ===== MySQL 操作 =====
		// 更新 GraphTagNode 表中的 name 字段
		oldNode.Name = req.Request.NodeName
		if err := tx.Save(oldNode).Error; err != nil {
			logs.ErrorContextf(ctx, "RenameNode.Save(oldNode:%v) error: %v", oldNode, err)
			return err
		}

		// ===== Nebula 操作 =====
		// 7.1 先删除 Nebula 中所有相关的边
		for _, edge := range edges {
			if err := cli.DeleteEdge(edge.EdgeTypeName, edge.SrcNodeName, edge.DstNodeName); err != nil {
				logs.WarnContextf(ctx, "RenameNode.DeleteEdge(edgeType:%s, src:%s, dst:%s) error: %v",
					edge.EdgeTypeName, edge.SrcNodeName, edge.DstNodeName, err)
				return err
			}
		}

		// 7.2 删除旧节点的 tag
		if err := cli.DeleteTag(req.Request.OldNodeName, tag.TagName); err != nil {
			logs.WarnContextf(ctx, "RenameNode.DeleteTag(nodeName:%s, tagName:%s) error: %v",
				req.Request.OldNodeName, tag.TagName, err)
			return err
		}

		// 7.3 插入新节点（使用新名称，保留原有属性）
		if err := cli.InsertNode(&foresttype.TagNodeInfo{
			GraphID:          graphInfo.ID,
			GraphVersionID:   graphInfo.VersionID,
			Uin:              graphInfo.Uin,
			CompanyID:        graphInfo.CompanyID,
			TagID:            tag.ID,
			TagName:          tag.TagName,
			Properties:       tag.Properties,
			PropertiesValues: oldNode.PropertiesValues,
			Name:             req.Request.NodeName,
		}); err != nil {
			logs.ErrorContextf(ctx, "RenameNode.InsertNode(tagName:%s, nodeName:%s) error: %v",
				tag.TagName, req.Request.NodeName, err)
			return err
		}

		// 7.4 重建所有边（将边中的旧节点名替换为新节点名）
		if len(edges) > 0 {
			edgeInfos := make([]*foresttype.EdgeInfo, 0, len(edges))
			for _, edge := range edges {
				newEdge := &foresttype.EdgeInfo{
					GraphID:          edge.GraphID,
					GraphVersionID:   edge.GraphVersionID,
					SrcID:            edge.SrcID,
					DstID:            edge.DstID,
					TagID:            edge.TagID,
					SrcTagID:         edge.SrcTagID,
					DstTagID:         edge.DstTagID,
					EdgeTypeName:     edge.EdgeTypeName,
					SrcNodeName:      edge.SrcNodeName,
					DstNodeName:      edge.DstNodeName,
					PropertiesValues: edge.PropertiesValues,
				}
				// 替换边中的旧节点名为新节点名
				if edge.SrcNodeName == req.Request.OldNodeName && edge.SrcTagID == req.Request.TagID {
					newEdge.SrcNodeName = req.Request.NodeName
				}
				if edge.DstNodeName == req.Request.OldNodeName && edge.DstTagID == req.Request.TagID {
					newEdge.DstNodeName = req.Request.NodeName
				}
				edgeInfos = append(edgeInfos, newEdge)
			}

			// 批量插入边
			if err := cli.InsertEdges(edgeInfos); err != nil {
				logs.ErrorContextf(ctx, "RenameNode.InsertEdges(edges:%v) error: %v", edgeInfos, err)
				return err
			}
		}

		return nil
	}); err != nil {
		logs.ErrorContextf(ctx, "RenameNode.Transaction error: %v", err)
		return res, err
	}

	return res, nil
}
