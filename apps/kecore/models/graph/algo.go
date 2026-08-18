package graph

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/nebulagraph"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/types"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TODO 批量插入mysql，nebula
// ParseAlgoResault 解析算法传来结构体
func (wrapper *ParseAlgoWrapper) ParseAlgoResault() error {
	// 第一步：查询上一版本的手动创建节点并合并属性
	if err := wrapper.mergePreviousVersionManualNodes(); err != nil {
		logs.ErrorContextf(wrapper.ctx, "ParseAlgoResault mergePreviousVersionManualNodes err: %v", err)
		// 合并失败不影响算法节点的创建，只记录错误
	}

	// 用于本次解析过程中去重（同一次 resault.Nodes 内部重复）
	seenAlgoKeys := make(map[string]struct{}, len(wrapper.resault.Nodes))
	for _, v := range wrapper.resault.Nodes {
		// 如果该节点已由“上一版本手动节点复制”创建，则跳过算法节点创建
		algoKey := fmt.Sprintf("%s:%s", v.Node, v.TagName)
		if _, ok := wrapper.copiedManualNodeKeys[algoKey]; ok {
			continue
		}
		// 同一次算法结果内部去重
		if _, ok := seenAlgoKeys[algoKey]; ok {
			continue
		}
		seenAlgoKeys[algoKey] = struct{}{}

		tag, ok := wrapper.tMap[v.TagName]
		if !ok {
			logs.WarnContextf(wrapper.ctx, "tag: %s not found", v.TagName)
			continue
		}
		err := dbutil.Knownow().WithContext(wrapper.ctx).Transaction(func(tx *gorm.DB) error {
			nodeInfo, err := wrapper.insertMysqlNode(tx, tag, v)
			if err != nil {
				logs.ErrorContextf(wrapper.ctx, "ParseAlgoResault insertMysqlNode err: %v", err)
				return err
			}
			// 插入图谱
			err = wrapper.cli.InsertNode(nodeInfo)
			if err != nil {
				logs.ErrorContextf(wrapper.ctx, "ParseAlgoResault InsertNode err: %v", err)
				return err
			}
			return nil
		})
		if err != nil {
			logs.ErrorContextf(wrapper.ctx, "ParseAlgoResault Transaction insert node err: %v", err)
			// 一个节点失败不影响其他节点插入
			continue
		}

	}
	for _, v := range wrapper.resault.Edges {
		// 检查并获取值
		edgeType, edgeOk := wrapper.eMap[v.EdgeName]
		if !edgeOk {
			logs.WarnContextf(wrapper.ctx, "edgeType %s not found", v.EdgeName)
			continue
		}
		srctag, srcOk := wrapper.tMap[v.SrcTag]
		if !srcOk {
			logs.WarnContextf(wrapper.ctx, "srctag %s not found", v.SrcTag)
			continue
		}
		dsttag, dstOk := wrapper.tMap[v.DstTag]
		if !dstOk {
			logs.WarnContextf(wrapper.ctx, "dsttag %s not found", v.DstTag)
			continue
		}
		// 没有类型不插入
		v.SrcTagID = srctag.ID
		v.DstTagID = dsttag.ID

		_, err := GetEdgeTag(wrapper.ctx, edgeType.ID, srctag.ID, dsttag.ID)
		if err != nil {
			logs.WarnContextf(wrapper.ctx, "GetEdgeTag err: %v", err)
			continue
		}
		err = dbutil.Knownow().WithContext(wrapper.ctx).Transaction(func(tx *gorm.DB) error {
			edgeInfo, err := wrapper.insertMysqlEdge(tx, edgeType, srctag, dsttag, v)
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					return err
				}
				logs.ErrorContextf(wrapper.ctx, "ParseAlgoResault insertMysqlEdge err: %v", err)
				return err
			}
			// 插入图谱
			err = wrapper.cli.InsertEdge(edgeInfo)
			if err != nil {
				logs.ErrorContextf(wrapper.ctx, "ParseAlgoResault InsertEdge err: %v", err)
				return err
			}
			return nil
		})
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				continue
			}
			logs.ErrorContextf(wrapper.ctx, "ParseAlgoResault Transaction insert edge err: %v", err)
			continue
		}
	}
	return nil
}

// insertMysqlNode 向mysql插入节点
func (wrapper *ParseAlgoWrapper) insertMysqlNode(tx *gorm.DB, tag *foresttype.GraphTag, resNode Node) (*foresttype.TagNodeInfo, error) {
	// 校验节点值是否正确
	var (
		err error
	)
	resNode.PropertiesValuse, err = resNode.PropertiesValuse.ValidateAndComplete(tag)
	if err != nil {
		logs.ErrorContextf(wrapper.ctx, "ParseAlgoResault ValidateAndComplete err: %v,value: %s", err, logs.JSON(resNode))
		return nil, err
	}

	tNode, err := GetTNodeByTagIDNodeID(wrapper.ctx, tag.ID, resNode.Node)
	if err != nil && err != gorm.ErrRecordNotFound {
		logs.ErrorContextf(wrapper.ctx, "ParseAlgoResault GetNodeByName err: %v", err)
		return nil, err
	}
	rcDao := NewKeGraphResourceChunkDao()

	if err == gorm.ErrRecordNotFound {
		tNode = &foresttype.GraphTagNode{
			Uin:            tag.Uin,
			CompanyID:      tag.CompanyID,
			GraphID:        tag.GraphID,
			GraphVersionID: wrapper.graph.VersionID,
			Name:           resNode.Node,

			TagID:            tag.ID,
			FileIDList:       types.NewUintArray([]uint{wrapper.resault.FileID}),
			PropertiesValues: resNode.PropertiesValuse,
		}

		err = CreateTagNodeTX(wrapper.ctx, tx, tNode)
		if err != nil {
			logs.ErrorContextf(wrapper.ctx, "ParseAlgoResault CreateNewNode err: %v", err)
			return nil, err
		}
		chunklist := foresttype.KeGraphResourceChunkList{}
		for _, v := range resNode.ChunkIDs {
			// tNode.ChunkIDList.Add(v)
			chunklist = append(chunklist, foresttype.KeGraphResourceChunk{
				ChunkID:        v,
				GraphID:        wrapper.graph.ID,
				GraphVersionID: wrapper.graph.VersionID,
				ResourceID:     tNode.ID,
				ResourceType:   foresttype.KeGraphResourceChunkTypeNode,
			})
		}
		chunklist.DeduplicateByChunkID()
		err = rcDao.WithTx(tx).BatchReplace(wrapper.ctx, chunklist)
		if err != nil {
			logs.ErrorContextf(wrapper.ctx, "rcDao.BatchReplace err: %v", err)
			return nil, err
		}
		nodeInfo := &foresttype.TagNodeInfo{
			Name:             tNode.Name,
			NodeID:           tNode.ID,
			NodeTagID:        tNode.ID,
			Uin:              tNode.Uin,
			CompanyID:        tNode.CompanyID,
			TagID:            tag.ID,
			GraphID:          tag.GraphID,
			GraphVersionID:   wrapper.graph.VersionID,
			FileIDList:       tNode.FileIDList,
			ChunkIDList:      tNode.ChunkIDList,
			PropertiesValues: tNode.PropertiesValues,
			Properties:       tag.Properties,
			TagName:          tag.TagName,
		}
		return nodeInfo, nil
	}
	{
		// tNode, err := GetTNodeByTagIDNodeID(wrapper.ctx, tag.ID, node.ID)
		// if err != nil && err != gorm.ErrRecordNotFound {
		// 	logs.ErrorContextf(wrapper.ctx, "ParseAlgoResault GetTNodeByTagIDNodeID err: %v", err)
		// 	return nil, err
		// }
		// if err == gorm.ErrRecordNotFound {
		// 	tNode = &foresttype.GraphTagNode{
		// 		GraphID:        wrapper.graph.ID,
		// 		GraphVersionID: wrapper.graph.VersionID,
		// 		NodeID:         node.ID,
		// 		TagID:          tag.ID,
		// 		FileIDList:     types.NewUintArray([]uint{wrapper.resault.FileID}),
		// 		// ChunkIDList:      types.NewStringArray(resNode.ChunkIDs),
		// 		PropertiesValues: resNode.PropertiesValuse,
		// 	}
		// 	err = tx.Create(tNode).Error
		// 	if err != nil {
		// 		logs.ErrorContextf(wrapper.ctx, "ParseAlgoResault CreateNewNode err: %v", err)
		// 		return nil, err
		// 	}
		// 	chunklist := foresttype.KeGraphResourceChunkList{}
		// 	for _, v := range resNode.ChunkIDs {
		// 		// tNode.ChunkIDList.Add(v)
		// 		chunklist = append(chunklist, foresttype.KeGraphResourceChunk{
		// 			ChunkID:        v,
		// 			GraphID:        wrapper.graph.ID,
		// 			GraphVersionID: wrapper.graph.VersionID,
		// 			ResourceID:     tNode.ID,
		// 			ResourceType:   foresttype.KeGraphResourceChunkTypeNode,
		// 		})
		// 	}
		// 	chunklist.DeduplicateByChunkID()
		// 	err = rcDao.WithTx(tx).BatchReplace(wrapper.ctx, chunklist)
		// 	if err != nil {
		// 		logs.ErrorContextf(wrapper.ctx, "rcDao.BatchReplace err: %v", err)
		// 		return nil, err
		// 	}
		// 	nodeInfo := &foresttype.TagNodeInfo{
		// 		Name:             node.Name,
		// 		NodeID:           node.ID,
		// 		NodeTagID:        tNode.ID,
		// 		Uin:              node.Uin,
		// 		CompanyID:        node.CompanyID,
		// 		TagID:            tag.ID,
		// 		GraphID:          tag.GraphID,
		// 		GraphVersionID:   wrapper.graph.VersionID,
		// 		FileIDList:       tNode.FileIDList,
		// 		ChunkIDList:      tNode.ChunkIDList,
		// 		PropertiesValues: tNode.PropertiesValues,
		// 		Properties:       tag.Properties,
		// 		TagName:          tag.TagName,
		// 	}
		// 	return nodeInfo, nil
		// }
	}
	tNode.FileIDList.Append(wrapper.resault.FileID)
	tNode.FileIDList.RemoveDuplicates()

	chunklist, err := rcDao.GetListByCond(wrapper.ctx, &KeGraphResourceChunkCond{Filters: []apiobj.Filter{
		{Field: "resource_type", Value: []string{string(foresttype.KeGraphResourceChunkTypeNode)}},
		{Field: "resource_id", Value: []string{fmt.Sprintf("%d", tNode.ID)}},
	}})
	if err != nil {
		logs.ErrorContextf(wrapper.ctx, "rcDao.GetListByCond err: %v", err)
		return nil, err
	}

	for _, v := range resNode.ChunkIDs {
		// tNode.ChunkIDList.Add(v)
		chunklist = append(chunklist, foresttype.KeGraphResourceChunk{
			ChunkID:        v,
			GraphID:        wrapper.graph.ID,
			GraphVersionID: wrapper.graph.VersionID,
			ResourceID:     tNode.ID,
			ResourceType:   foresttype.KeGraphResourceChunkTypeNode,
		})
	}
	chunklist.DeduplicateByChunkID()
	err = rcDao.WithTx(tx).BatchReplace(wrapper.ctx, chunklist)
	if err != nil {
		logs.ErrorContextf(wrapper.ctx, "rcDao.BatchReplace err: %v", err)
		return nil, err
	}

	tNode.PropertiesValues.UpdateAndSyncProperties(tag.Properties, resNode.PropertiesValuse)

	err = tx.Save(tNode).Error
	if err != nil {
		logs.ErrorContextf(wrapper.ctx, "ParseAlgoResault UpdateTNode err: %v", err)
		return nil, err
	}
	nodeInfo := &foresttype.TagNodeInfo{
		Name:             tNode.Name,
		NodeID:           tNode.ID,
		NodeTagID:        tNode.ID,
		Uin:              tNode.Uin,
		CompanyID:        tNode.CompanyID,
		TagID:            tag.ID,
		GraphID:          tag.GraphID,
		GraphVersionID:   wrapper.graph.VersionID,
		FileIDList:       tNode.FileIDList,
		ChunkIDList:      tNode.ChunkIDList,
		PropertiesValues: tNode.PropertiesValues,
		Properties:       tag.Properties,
		TagName:          tag.TagName,
	}
	return nodeInfo, nil
}

// insertMysqlEdge 向mysql插入边
func (wrapper *ParseAlgoWrapper) insertMysqlEdge(tx *gorm.DB, edgeType, srctag, dsttag *foresttype.GraphTag, resEdge EdgeValue) (*foresttype.EdgeInfo, error) {
	srcNode, err := GetTNodeByTagIDNodeID(wrapper.ctx, resEdge.SrcTagID, resEdge.SrcNode)
	if err != nil {
		logs.WarnContextf(wrapper.ctx, "ParseAlgoResault GetNodeByName err: %v", err)
		return nil, err
	}
	dstNode, err := GetTNodeByTagIDNodeID(wrapper.ctx, resEdge.DstTagID, resEdge.DstNode)
	if err != nil {
		logs.WarnContextf(wrapper.ctx, "ParseAlgoResault GetNodeByName err: %v", err)
		return nil, err
	}
	edge, err := GetGraphEdge(wrapper.ctx, edgeType.GraphID, srcNode.ID, dstNode.ID)
	if err != nil && err != gorm.ErrRecordNotFound {
		logs.ErrorContextf(wrapper.ctx, "ParseAlgoResault GetGraphEdge err: %v", err)
		return nil, err
	}

	rcDao := NewKeGraphResourceChunkDao()

	if err == gorm.ErrRecordNotFound {
		edge = &foresttype.GraphEdge{
			GraphID:        edgeType.GraphID,
			GraphVersionID: wrapper.graph.VersionID,
			SrcID:          srcNode.ID,
			SrcTagID:       srctag.ID,
			DstID:          dstNode.ID,
			DstTagID:       dsttag.ID,
			TagID:          edgeType.ID,
			FileIDList:     types.NewUintArray([]uint{wrapper.resault.FileID}),
			// ChunkIDList:    types.NewStringArray(resEdge.ChunkIDs),
		}
		err = tx.Create(edge).Error
		if err != nil {
			logs.ErrorContextf(wrapper.ctx, "ParseAlgoResault CreateNewEdge err: %v", err)
			return nil, err
		}
		chunklist := foresttype.KeGraphResourceChunkList{}
		for _, v := range resEdge.ChunkIDs {
			// tNode.ChunkIDList.Add(v)
			chunklist = append(chunklist, foresttype.KeGraphResourceChunk{
				ChunkID:        v,
				GraphID:        wrapper.graph.ID,
				GraphVersionID: wrapper.graph.VersionID,
				ResourceID:     edge.ID,
				ResourceType:   foresttype.KeGraphResourceChunkTypeEdge,
			})
		}
		chunklist.DeduplicateByChunkID()
		err = rcDao.WithTx(tx).BatchReplace(wrapper.ctx, chunklist)
		if err != nil {
			logs.ErrorContextf(wrapper.ctx, "rcDao.BatchReplace err: %v", err)
			return nil, err
		}
		edgeInfo := &foresttype.EdgeInfo{
			EdgeID:           edge.ID,
			SrcID:            srcNode.ID,
			SrcTagID:         srctag.ID,
			DstID:            dstNode.ID,
			DstTagID:         dsttag.ID,
			TagID:            edgeType.ID,
			FileIDList:       edge.FileIDList,
			ChunkIDList:      edge.ChunkIDList,
			EdgeTypeName:     edgeType.TagName,
			SrcNodeName:      srcNode.Name,
			DstNodeName:      dstNode.Name,
			PropertiesValues: edge.PropertiesValues,
			Properties:       edgeType.Properties,
		}

		return edgeInfo, nil
	}
	edge.FileIDList.Append(wrapper.resault.FileID)
	edge.FileIDList.RemoveDuplicates()

	chunklist, err := rcDao.GetListByCond(wrapper.ctx, &KeGraphResourceChunkCond{Filters: []apiobj.Filter{
		{Field: "resource_type", Value: []string{string(foresttype.KeGraphResourceChunkTypeEdge)}},
		{Field: "resource_id", Value: []string{fmt.Sprintf("%d", edge.ID)}},
	}})
	if err != nil {
		logs.ErrorContextf(wrapper.ctx, "rcDao.GetListByCond err: %v", err)
		return nil, err
	}

	for _, v := range resEdge.ChunkIDs {
		// tNode.ChunkIDList.Add(v)
		chunklist = append(chunklist, foresttype.KeGraphResourceChunk{
			ChunkID:        v,
			GraphID:        wrapper.graph.ID,
			GraphVersionID: wrapper.graph.VersionID,
			ResourceID:     edgeType.ID,
			ResourceType:   foresttype.KeGraphResourceChunkTypeEdge,
		})
	}
	chunklist.DeduplicateByChunkID()
	err = rcDao.WithTx(tx).BatchReplace(wrapper.ctx, chunklist)
	if err != nil {
		logs.ErrorContextf(wrapper.ctx, "rcDao.BatchReplace err: %v", err)
		return nil, err
	}

	err = tx.Save(edge).Error
	if err != nil {
		logs.ErrorContextf(wrapper.ctx, "ParseAlgoResault UpdateEdge err: %v", err)
		return nil, err
	}
	edgeInfo := &foresttype.EdgeInfo{
		GraphID:          wrapper.graph.ID,
		GraphVersionID:   wrapper.graph.VersionID,
		EdgeID:           edge.ID,
		SrcID:            srcNode.ID,
		SrcTagID:         srctag.ID,
		DstID:            dstNode.ID,
		DstTagID:         dsttag.ID,
		TagID:            edgeType.ID,
		FileIDList:       edge.FileIDList,
		ChunkIDList:      edge.ChunkIDList,
		EdgeTypeName:     edgeType.TagName,
		SrcNodeName:      srcNode.Name,
		DstNodeName:      dstNode.Name,
		PropertiesValues: edge.PropertiesValues,
		Properties:       edgeType.Properties,
	}
	return edgeInfo, nil
}

func (wrapper *ParseAlgoWrapper) upsertTag() error {
	newTags := map[string]*foresttype.GraphTag{}
	alterTags := map[string]*foresttype.GraphTag{}
	for _, v := range wrapper.resault.Nodes {
		tag, ok := wrapper.tMap[v.TagName]
		if !ok {
			// 判断是否以及存在同名边了
			_, ok := wrapper.eMap[v.TagName]
			if ok {
				// 如果存在跳出节点
				logs.WarnContextf(wrapper.ctx, "ParseAlgoResault upsertTag tag: %s already exists", v.TagName)
				continue
			}
			// 如果不存在，创建一个新标签
			newTag := &foresttype.GraphTag{
				GraphID:        wrapper.graph.ID,
				GraphVersionID: wrapper.graph.VersionID,
				TagName:        v.TagName,
				TagType:        foresttype.TagTypeNode,
				Uin:            wrapper.graph.Uin,
				CompanyID:      wrapper.graph.CompanyID,
			}
			newTags[v.TagName] = newTag
			wrapper.tMap[v.TagName] = newTag
			tag = newTag
		}
		originalPropCount := len(tag.Properties)
		propMap := tag.Properties.NameMap()

		for _, pv := range v.PropertiesValuse {
			_, exists := propMap[pv.Name]
			if !exists {
				// 如果属性不存在，添加一个属性
				propMap[pv.Name] = &foresttype.Property{
					Name: pv.Name,
					Type: "string", // TODO 新增属性暂时都用string
				}
				tag.Properties = append(tag.Properties, propMap[pv.Name])
			}
		}
		if ok && len(tag.Properties) != originalPropCount {
			if _, exnew := newTags[v.TagName]; !exnew {
				if _, exal := alterTags[v.TagName]; !exal {
					alterTags[v.TagName] = tag
				}
			}
		}
	}
	for _, v := range newTags {
		err := wrapper.cli.CreateGraphTag(dbutil.Knownow(), v)
		if err != nil {
			logs.ErrorContextf(wrapper.ctx, "CreateGraphTag error: %v", err)
			return err
		}
		// err = CreateTag(wrapper.ctx, wrapper.graph.SpaceName, v)
		// if err != nil {
		// 	logs.ErrorContextf(wrapper.ctx, "upsertTag CreateTag err: %v", err)
		// 	return err
		// }
	}

	for _, v := range alterTags {
		v.Properties.Deduplicate()
		err := wrapper.cli.AlterTag(v, false)
		if err != nil {
			logs.ErrorContextf(wrapper.ctx, "upsertTag AlterTag err: %v", err)
			return err
		}
		err = UpdateTag(wrapper.ctx, v)
		if err != nil {
			logs.ErrorContextf(wrapper.ctx, "upsertTag UpdateTag err: %v", err)
			return err
		}
	}
	return nil
}

func (wrapper *ParseAlgoWrapper) upsertEdge() error {
	for _, v := range wrapper.resault.Edges {
		if _, ok := wrapper.tMap[v.SrcTag]; !ok {
			continue
		}
		if _, ok := wrapper.tMap[v.DstTag]; !ok {
			continue
		}
		edge, edgeOk := wrapper.eMap[v.EdgeName]
		if !edgeOk {
			_, ok := wrapper.tMap[v.EdgeName]
			if ok {
				// 如果存在跳出节点
				logs.WarnContextf(wrapper.ctx, "ParseAlgoResault upsertEdge tag: %s already exists", v.EdgeName)
				continue
			}
			newEdge := &foresttype.GraphTag{
				GraphID:        wrapper.graph.ID,
				GraphVersionID: wrapper.graph.VersionID,
				TagName:        v.EdgeName,
				TagType:        foresttype.TagTypeEdge,
				Uin:            wrapper.graph.Uin,
				CompanyID:      wrapper.graph.CompanyID,
			}
			err := wrapper.cli.CreateGraphTag(dbutil.Knownow(), newEdge)
			if err != nil {
				logs.ErrorContextf(wrapper.ctx, "CreateGraphTag error: %v", err)
				return err
			}
			// err = CreateTag(wrapper.ctx, wrapper.graph.SpaceName, newEdge)
			// if err != nil {
			// 	logs.ErrorContextf(wrapper.ctx, "upsertEdge CreateTag err: %v", err)
			// 	return err
			// }
			et := &foresttype.GraphEdgeTag{
				GraphID:        wrapper.graph.ID,
				GraphVersionID: wrapper.graph.VersionID,
				EdgeTypeID:     newEdge.ID,
				SrcTagID:       wrapper.tMap[v.SrcTag].ID,
				DstTagID:       wrapper.tMap[v.DstTag].ID,
			}
			err = dbutil.Knownow().Create(et).Error
			if err != nil {
				logs.ErrorContextf(wrapper.ctx, "upsertEdge CreateEdgeType err: %v", err)
				return err
			}
			wrapper.eMap[v.EdgeName] = newEdge
			edge = newEdge
			continue
		}
		_, err := GetEdgeTag(wrapper.ctx, edge.ID, wrapper.tMap[v.SrcTag].ID, wrapper.tMap[v.DstTag].ID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				et := &foresttype.GraphEdgeTag{
					GraphID:        wrapper.graph.ID,
					GraphVersionID: wrapper.graph.VersionID,
					EdgeTypeID:     edge.ID,
					SrcTagID:       wrapper.tMap[v.SrcTag].ID,
					DstTagID:       wrapper.tMap[v.DstTag].ID,
				}
				err = dbutil.Knownow().Create(et).Error
				if err != nil {
					logs.ErrorContextf(wrapper.ctx, "upsertEdge CreateEdgeType err: %v", err)
					return err
				}
				continue
			}
			logs.ErrorContextf(wrapper.ctx, "upsertEdge GetEdgeTag err: %v", err)
			return err
		}

	}
	return nil
}

// mergePreviousVersionManualNodes 查询上一版本的手动创建节点并合并属性（不插入新节点）
func (wrapper *ParseAlgoWrapper) mergePreviousVersionManualNodes() error {
	// 获取上一个版本
	previousVersion, err := GetPreviousVersion(wrapper.ctx, wrapper.graph.ID, wrapper.graph.VersionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logs.WarnContextf(wrapper.ctx, "mergePreviousVersionManualNodes: no previous version found")
			return nil // 没有上一版本不是错误
		}
		return err
	}

	// 查询上一版本的所有手动创建的节点
	var oldNodes []*foresttype.GraphTagNode
	err = dbutil.Knownow().WithContext(wrapper.ctx).
		Where("graph_id = ?", previousVersion.ID).
		Where("graph_version_id = ?", previousVersion.VersionID).
		Where("created_type = ?", foresttype.CreatedTypeManual).
		Find(&oldNodes).Error
	if err != nil {
		logs.ErrorContextf(wrapper.ctx, "mergePreviousVersionManualNodes query old nodes err: %v", err)
		return err
	}

	if len(oldNodes) == 0 {
		logs.WarnContextf(wrapper.ctx, "mergePreviousVersionManualNodes: no manual nodes found in previous version")
		return nil
	}

	// 查询上一版本的所有 tag
	var oldTags []*foresttype.GraphTag
	err = dbutil.Knownow().WithContext(wrapper.ctx).
		Where("graph_id = ?", previousVersion.ID).
		Where("graph_version_id = ?", previousVersion.VersionID).
		Find(&oldTags).Error
	if err != nil {
		logs.ErrorContextf(wrapper.ctx, "mergePreviousVersionManualNodes query old tags err: %v", err)
		return err
	}

	// 建立旧 tag ID 到旧 tag 的映射
	oldTagMap := make(map[uint]*foresttype.GraphTag)
	for _, tag := range oldTags {
		oldTagMap[tag.ID] = tag
	}

	// 建立算法结果中的节点 key 映射（nodeName:tagName）
	algoNodeMap := make(map[string]Node)
	for _, v := range wrapper.resault.Nodes {
		algoKey := fmt.Sprintf("%s:%s", v.Node, v.TagName)
		algoNodeMap[algoKey] = v
	}

	// 第一步：收集需要更新的节点信息（tagID + nodeName）
	type nodeKey struct {
		TagID    uint
		NodeName string
	}
	nodeKeysToQuery := make([]nodeKey, 0)
	nodeKeyToOldNode := make(map[nodeKey]*foresttype.GraphTagNode)
	nodeKeyToAlgoNode := make(map[nodeKey]Node)
	nodeKeyToNewTag := make(map[nodeKey]*foresttype.GraphTag)

	for _, oldNode := range oldNodes {
		oldTag, ok := oldTagMap[oldNode.TagID]
		if !ok {
			continue
		}

		// 在新版本中查找同名的 tag
		newTag, ok := wrapper.tMap[oldTag.TagName]
		if !ok {
			continue
		}

		// 检查算法结果中是否有同名同tag的节点
		algoKey := fmt.Sprintf("%s:%s", oldNode.Name, oldTag.TagName)
		algoNode, hasAlgoNode := algoNodeMap[algoKey]

		if hasAlgoNode {
			key := nodeKey{TagID: newTag.ID, NodeName: oldNode.Name}
			nodeKeysToQuery = append(nodeKeysToQuery, key)
			nodeKeyToOldNode[key] = oldNode
			nodeKeyToAlgoNode[key] = algoNode
			nodeKeyToNewTag[key] = newTag
		}
	}

	if len(nodeKeysToQuery) == 0 {
		return nil
	}

	// 第二步：批量查询当前版本中已存在的节点
	currentNodesMap := make(map[nodeKey]*foresttype.GraphTagNode)
	for _, key := range nodeKeysToQuery {
		currentNode, err := GetTNodeByTagIDNodeID(wrapper.ctx, key.TagID, key.NodeName)
		if err != nil && err != gorm.ErrRecordNotFound {
			logs.ErrorContextf(wrapper.ctx, "mergePreviousVersionManualNodes GetTNodeByTagIDNodeID err: %v", err)
			continue
		}
		if err == nil {
			currentNodesMap[key] = currentNode
		}
	}

	if len(currentNodesMap) == 0 {
		return nil
	}

	// 第三步：在内存中组装需要更新的数据
	type nodeUpdateData struct {
		node           *foresttype.GraphTagNode
		newTag         *foresttype.GraphTag
		validatedProps foresttype.PropertiesValues
		chunkIDs       []string
		nodeInfo       *foresttype.TagNodeInfo
	}
	updates := make([]nodeUpdateData, 0, len(currentNodesMap))
	allChunkList := make(foresttype.KeGraphResourceChunkList, 0)

	for key, currentNode := range currentNodesMap {
		algoNode := nodeKeyToAlgoNode[key]
		newTag := nodeKeyToNewTag[key]

		// 验证并完成属性值
		validatedProps, err := algoNode.PropertiesValuse.ValidateAndComplete(newTag)
		if err != nil {
			logs.ErrorContextf(wrapper.ctx, "mergePreviousVersionManualNodes ValidateAndComplete err: %v", err)
			continue
		}

		// 合并属性
		currentNode.PropertiesValues.UpdateAndSyncProperties(newTag.Properties, validatedProps)

		// 更新 FileIDList
		currentNode.FileIDList.Append(wrapper.resault.FileID)
		currentNode.FileIDList.RemoveDuplicates()

		// 准备 ChunkIDList
		for _, chunkID := range algoNode.ChunkIDs {
			allChunkList = append(allChunkList, foresttype.KeGraphResourceChunk{
				ChunkID:        chunkID,
				GraphID:        wrapper.graph.ID,
				GraphVersionID: wrapper.graph.VersionID,
				ResourceID:     currentNode.ID,
				ResourceType:   foresttype.KeGraphResourceChunkTypeNode,
			})
		}

		// 准备 Nebula 节点信息
		nodeInfo := &foresttype.TagNodeInfo{
			Name:             currentNode.Name,
			NodeID:           currentNode.ID,
			NodeTagID:        currentNode.ID,
			Uin:              currentNode.Uin,
			CompanyID:        currentNode.CompanyID,
			TagID:            newTag.ID,
			GraphID:          wrapper.graph.ID,
			GraphVersionID:   wrapper.graph.VersionID,
			FileIDList:       currentNode.FileIDList,
			ChunkIDList:      currentNode.ChunkIDList,
			PropertiesValues: currentNode.PropertiesValues,
			Properties:       newTag.Properties,
			TagName:          newTag.TagName,
		}

		updates = append(updates, nodeUpdateData{
			node:           currentNode,
			newTag:         newTag,
			validatedProps: validatedProps,
			chunkIDs:       algoNode.ChunkIDs,
			nodeInfo:       nodeInfo,
		})

		// 记录已合并的节点，避免后续重复处理
		algoKey := fmt.Sprintf("%s:%s", currentNode.Name, newTag.TagName)
		wrapper.copiedManualNodeKeys[algoKey] = struct{}{}
	}

	// 第四步：统一在一个事务中批量处理
	err = dbutil.Knownow().WithContext(wrapper.ctx).Transaction(func(tx *gorm.DB) error {
		// 批量更新节点：使用 CreateInBatches + ON CONFLICT
		nodesToUpsert := make([]*foresttype.GraphTagNode, 0, len(updates))
		for _, update := range updates {
			nodesToUpsert = append(nodesToUpsert, update.node)
		}

		if len(nodesToUpsert) > 0 {
			err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{
					{Name: "tag_id"},
					{Name: "name"},
				},
				DoUpdates: clause.AssignmentColumns([]string{"properties_values", "file_id_list", "updated_at"}),
			}).CreateInBatches(nodesToUpsert, 100).Error
			if err != nil {
				logs.ErrorContextf(wrapper.ctx, "mergePreviousVersionManualNodes batch upsert node err: %v", err)
				return err
			}
		}

		// 批量更新 ChunkIDList
		if len(allChunkList) > 0 {
			allChunkList.DeduplicateByChunkID()
			rcDao := NewKeGraphResourceChunkDao()
			err = rcDao.WithTx(tx).BatchReplace(wrapper.ctx, allChunkList)
			if err != nil {
				logs.ErrorContextf(wrapper.ctx, "mergePreviousVersionManualNodes BatchReplace err: %v", err)
				return err
			}
		}

		return nil
	})
	if err != nil {
		logs.ErrorContextf(wrapper.ctx, "mergePreviousVersionManualNodes transaction err: %v", err)
		return err
	}

	// 第五步：更新 Nebula 中的节点（在事务外，循环调用 InsertNode）
	for _, update := range updates {
		err = wrapper.cli.InsertNode(update.nodeInfo)
		if err != nil {
			logs.ErrorContextf(wrapper.ctx, "mergePreviousVersionManualNodes InsertNode err: %v", err)
			// 不返回错误，继续处理其他节点
		}
	}

	return nil
}

// CopyPreviousVersionManualNodes 复制上一版本的手动创建节点到新版本（独立函数，供 ParseGraph 接口调用）
func CopyPreviousVersionManualNodes(ctx context.Context, graphInfo *foresttype.ForestGraphInfo, cli *nebulagraph.NebulaCli, tMap map[string]*foresttype.GraphTag, eMap map[string]*foresttype.GraphTag) error {
	// 获取上一个版本
	previousVersion, err := GetPreviousVersion(ctx, graphInfo.ID, graphInfo.VersionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logs.WarnContextf(ctx, "CopyPreviousVersionManualNodes: no previous version found")
			return nil // 没有上一版本不是错误
		}
		return err
	}

	// 第一步：查询上一版本的所有手动创建的节点
	var oldNodes []*foresttype.GraphTagNode
	err = dbutil.Knownow().WithContext(ctx).
		Where("graph_id = ?", previousVersion.ID).
		Where("graph_version_id = ?", previousVersion.VersionID).
		Where("created_type = ?", foresttype.CreatedTypeManual).
		Find(&oldNodes).Error
	if err != nil {
		logs.ErrorContextf(ctx, "CopyPreviousVersionManualNodes query old nodes err: %v", err)
		return err
	}

	if len(oldNodes) == 0 {
		logs.WarnContextf(ctx, "CopyPreviousVersionManualNodes: no manual nodes found in previous version")
		return nil
	}

	// 第二步：查询上一版本的所有 tag
	var oldTags []*foresttype.GraphTag
	err = dbutil.Knownow().WithContext(ctx).
		Where("graph_id = ?", previousVersion.ID).
		Where("graph_version_id = ?", previousVersion.VersionID).
		Find(&oldTags).Error
	if err != nil {
		logs.ErrorContextf(ctx, "CopyPreviousVersionManualNodes query old tags err: %v", err)
		return err
	}

	// 建立旧 tag ID 到旧 tag 的映射
	oldTagMap := make(map[uint]*foresttype.GraphTag)
	for _, tag := range oldTags {
		oldTagMap[tag.ID] = tag
	}

	// 建立旧节点 ID 到节点的映射（用于判断节点是否手动创建）
	oldNodeIDMap := make(map[uint]*foresttype.GraphTagNode)
	for _, node := range oldNodes {
		oldNodeIDMap[node.ID] = node
	}

	// 第三步：查询手动节点之间的边（源节点和目标节点都是手动创建）
	oldNodeIDs := make([]uint, 0, len(oldNodes))
	for _, node := range oldNodes {
		oldNodeIDs = append(oldNodeIDs, node.ID)
	}

	var oldEdges []*foresttype.GraphEdge
	if len(oldNodeIDs) > 0 {
		err = dbutil.Knownow().WithContext(ctx).
			Where("graph_id = ?", previousVersion.ID).
			Where("graph_version_id = ?", previousVersion.VersionID).
			Where("src_id IN ?", oldNodeIDs).
			Where("dst_id IN ?", oldNodeIDs). // 只查询目标节点也在手动节点列表中的边
			Find(&oldEdges).Error
		if err != nil {
			logs.ErrorContextf(ctx, "CopyPreviousVersionManualNodes query old edges err: %v", err)
			return err
		}
	}

	// 第四步：构建旧节点名称到新节点 ID 的映射（用于边的复制）
	oldNodeNameToNewID := make(map[string]uint)

	// 第五步：在一个事务中批量处理所有节点
	err = dbutil.Knownow().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		nodesToInsertNebula := make([]foresttype.TagNodeInfo, 0, len(oldNodes))

		// 处理所有节点
		for _, oldNode := range oldNodes {
			oldTag, ok := oldTagMap[oldNode.TagID]
			if !ok {
				logs.WarnContextf(ctx, "CopyPreviousVersionManualNodes: old tag not found for node %s", oldNode.Name)
				continue
			}

			// 在新版本中查找同名的 tag
			newTag, ok := tMap[oldTag.TagName]
			if !ok {
				logs.WarnContextf(ctx, "CopyPreviousVersionManualNodes: new tag %s not found", oldTag.TagName)
				continue
			}

			var nodeInfo foresttype.TagNodeInfo
			// 仅复制旧手动节点数据：直接创建为 manual
			newNode := &foresttype.GraphTagNode{
				Uin:              graphInfo.Uin,
				CompanyID:        graphInfo.CompanyID,
				GraphID:          graphInfo.ID,
				GraphVersionID:   graphInfo.VersionID,
				Name:             oldNode.Name,
				TagID:            newTag.ID,
				FileIDList:       oldNode.FileIDList,
				ChunkIDList:      oldNode.ChunkIDList,
				PropertiesValues: oldNode.PropertiesValues,
				CreatedType:      foresttype.CreatedTypeManual,
			}

			err = CreateTagNodeTX(ctx, tx, newNode)
			if err != nil {
				logs.ErrorContextf(ctx, "CopyPreviousVersionManualNodes create node err: %v", err)
				return err
			}

			// 记录旧节点名到新节点 ID 的映射（用于复制边）
			oldNodeKey := fmt.Sprintf("%s:%s", oldNode.Name, oldTag.TagName)
			oldNodeNameToNewID[oldNodeKey] = newNode.ID

			nodeInfo = foresttype.TagNodeInfo{
				GraphID:          graphInfo.ID,
				GraphVersionID:   graphInfo.VersionID,
				Uin:              graphInfo.Uin,
				CompanyID:        graphInfo.CompanyID,
				TagID:            newTag.ID,
				TagName:          newTag.TagName,
				Properties:       newTag.Properties,
				PropertiesValues: newNode.PropertiesValues,
				Name:             newNode.Name,
			}

			nodesToInsertNebula = append(nodesToInsertNebula, nodeInfo)
		}

		// 第六步：批量处理边（只复制手动节点到手动节点的边）
		if len(oldEdges) > 0 {
			err = copyManualEdgesBatch(ctx, tx, cli, graphInfo, oldEdges, oldTagMap, oldNodeIDMap, oldNodeNameToNewID, tMap, eMap)
			if err != nil {
				logs.ErrorContextf(ctx, "CopyPreviousVersionManualNodes copy edges err: %v", err)
				return err
			}
		}

		// 第七步：批量插入节点到 Nebula
		for _, nodeInfo := range nodesToInsertNebula {
			err = cli.InsertNode(&nodeInfo)
			if err != nil {
				logs.ErrorContextf(ctx, "CopyPreviousVersionManualNodes insert node to nebula err: %v", err)
				// 不返回错误，继续处理其他节点
			}
		}

		return nil
	})

	if err != nil {
		logs.ErrorContextf(ctx, "CopyPreviousVersionManualNodes transaction err: %v", err)
		return err
	}

	return nil
}

// copyManualEdgesBatch 批量复制手动节点之间的边到新版本（独立函数版本）
func copyManualEdgesBatch(ctx context.Context, tx *gorm.DB, cli *nebulagraph.NebulaCli, graphInfo *foresttype.ForestGraphInfo, oldEdges []*foresttype.GraphEdge, oldTagMap map[uint]*foresttype.GraphTag, oldNodeIDMap map[uint]*foresttype.GraphTagNode, oldNodeNameToNewID map[string]uint, tMap map[string]*foresttype.GraphTag, eMap map[string]*foresttype.GraphTag) error {
	// 过滤出手动节点到手动节点的边
	manualToManualEdges := make([]*foresttype.GraphEdge, 0)
	for _, oldEdge := range oldEdges {
		// 检查源节点和目标节点是否都是手动创建
		_, srcIsManual := oldNodeIDMap[oldEdge.SrcID]
		_, dstIsManual := oldNodeIDMap[oldEdge.DstID]

		if srcIsManual && dstIsManual {
			manualToManualEdges = append(manualToManualEdges, oldEdge)
		}
	}

	if len(manualToManualEdges) == 0 {
		return nil
	}

	// 构建边数据
	newEdges := make([]*foresttype.GraphEdge, 0)
	edgeInfos := make([]*foresttype.EdgeInfo, 0)

	for _, oldEdge := range manualToManualEdges {
		// 获取旧边的 tag
		oldEdgeTag, ok := oldTagMap[oldEdge.TagID]
		if !ok {
			logs.WarnContextf(ctx, "copyManualEdgesBatch: old edge tag not found")
			continue
		}

		// 查找新版本对应的边 tag
		newEdgeTag, ok := eMap[oldEdgeTag.TagName]
		if !ok {
			logs.WarnContextf(ctx, "copyManualEdgesBatch: new edge tag %s not found", oldEdgeTag.TagName)
			continue
		}

		// 获取源节点和目标节点的 tag
		oldSrcTag, srcOk := oldTagMap[oldEdge.SrcTagID]
		oldDstTag, dstOk := oldTagMap[oldEdge.DstTagID]
		if !srcOk || !dstOk {
			logs.WarnContextf(ctx, "copyManualEdgesBatch: source or dest tag not found")
			continue
		}

		// 在新版本中查找源节点和目标节点的 tag
		newSrcTag, srcOk := tMap[oldSrcTag.TagName]
		newDstTag, dstOk := tMap[oldDstTag.TagName]
		if !srcOk || !dstOk {
			logs.WarnContextf(ctx, "copyManualEdgesBatch: new source or dest tag not found")
			continue
		}

		// 获取旧源节点和目标节点
		oldSrcNode, srcOk := oldNodeIDMap[oldEdge.SrcID]
		oldDstNode, dstOk := oldNodeIDMap[oldEdge.DstID]
		if !srcOk || !dstOk {
			logs.WarnContextf(ctx, "copyManualEdgesBatch: old source or dest node not found")
			continue
		}

		// 通过旧节点名查找新节点 ID
		srcKey := fmt.Sprintf("%s:%s", oldSrcNode.Name, oldSrcTag.TagName)
		dstKey := fmt.Sprintf("%s:%s", oldDstNode.Name, oldDstTag.TagName)

		newSrcID, srcOk := oldNodeNameToNewID[srcKey]
		newDstID, dstOk := oldNodeNameToNewID[dstKey]
		if !srcOk || !dstOk {
			logs.WarnContextf(ctx, "copyManualEdgesBatch: new source or dest node not found in mapping")
			continue
		}

		// 检查边是否已存在
		var existingEdge foresttype.GraphEdge
		err := tx.WithContext(ctx).Where("graph_id = ?", graphInfo.ID).
			Where("graph_version_id = ?", graphInfo.VersionID).
			Where("tag_id = ?", newEdgeTag.ID).
			Where("src_id = ?", newSrcID).
			Where("dst_id = ?", newDstID).
			First(&existingEdge).Error

		if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
			// 边不存在，创建新边
			newEdge := &foresttype.GraphEdge{
				GraphID:          graphInfo.ID,
				GraphVersionID:   graphInfo.VersionID,
				TagID:            newEdgeTag.ID,
				SrcID:            newSrcID,
				DstID:            newDstID,
				SrcTagID:         newSrcTag.ID,
				DstTagID:         newDstTag.ID,
				PropertiesValues: oldEdge.PropertiesValues,
				FileIDList:       oldEdge.FileIDList,
				ChunkIDList:      oldEdge.ChunkIDList,
			}

			newEdges = append(newEdges, newEdge)

			// 准备 Nebula 边信息
			edgeInfo := &foresttype.EdgeInfo{
				GraphID:          graphInfo.ID,
				GraphVersionID:   graphInfo.VersionID,
				SrcID:            newSrcID,
				DstID:            newDstID,
				TagID:            newEdgeTag.ID,
				SrcTagID:         newSrcTag.ID,
				DstTagID:         newDstTag.ID,
				EdgeTypeName:     newEdgeTag.TagName,
				SrcNodeName:      oldSrcNode.Name,
				DstNodeName:      oldDstNode.Name,
				PropertiesValues: oldEdge.PropertiesValues,
			}

			edgeInfos = append(edgeInfos, edgeInfo)
		}
		// 如果边已存在或查询出错，跳过
	}

	// 批量插入边到 MySQL
	if len(newEdges) > 0 {
		err := tx.CreateInBatches(newEdges, 50).Error
		if err != nil {
			logs.ErrorContextf(ctx, "copyManualEdgesBatch create edges err: %v", err)
			return err
		}

		// 批量插入边到 Nebula
		err = cli.InsertEdges(edgeInfos)
		if err != nil {
			logs.ErrorContextf(ctx, "copyManualEdgesBatch insert edges to nebula err: %v", err)
			return err
		}
	}

	return nil
}

// copyManualEdgesBatch 批量复制手动节点之间的边到新版本（wrapper 方法版本，保持向后兼容）
func (wrapper *ParseAlgoWrapper) copyManualEdgesBatch(tx *gorm.DB, oldEdges []*foresttype.GraphEdge, oldTagMap map[uint]*foresttype.GraphTag, oldNodeIDMap map[uint]*foresttype.GraphTagNode, oldNodeNameToNewID map[string]uint) error {
	return copyManualEdgesBatch(wrapper.ctx, tx, wrapper.cli, wrapper.graph, oldEdges, oldTagMap, oldNodeIDMap, oldNodeNameToNewID, wrapper.tMap, wrapper.eMap)
}

// replaceString 去除图中的非法字符
func ReplaceString(str string) string {
	// 创建一个 Replacer，替换四组字符为空
	replacer := strings.NewReplacer(
		"|", "",
		">", "",
		"<", "",
		"\"", "",
		"'", "",
		"&gt;", "",
		"\n", "",
		".", "",
		",", "",
		";", "",
		"`", "",
		" ", "_",
		"-", "_",
		"+", "",
		"=", "",
		"(", "",
		")", "",
		"[", "",
		"]", "",
		"{", "",
		"}", "",
		"*", "",
		"/", "",
		"\\", "",
		"?", "",
		"!", "",
		"@", "",
		"#", "",
		"$", "",
		"%", "",
		"^", "",
		"~", "",
		":", "",
		"\t", "",
		"\r", "",
	)
	return replacer.Replace(str)
}

// replaceString 去除图中的非法字符
func ReplaceValue(str string) string {
	// 创建一个 Replacer，替换四组字符为空
	replacer := strings.NewReplacer(
		"|", "",
		">", "",
		"<", "",
		"\"", "",
		"'", "",
		"\n", "",
		",", "",
		";", "",
		"`", "",
		" ", "_",
		"-", "_",
		"+", "",
		"=", "",
		"(", "",
		")", "",
		"[", "",
		"]", "",
		"{", "",
		"}", "",
		"*", "",
		"/", "",
		"\\", "",
		"^", "",
		"~", "",
		":", "",
	)
	return replacer.Replace(str)
}
