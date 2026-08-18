package graph

import (
	"context"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

func InsertTemplate(ctx context.Context, graph *foresttype.ForestGraphInfo, template *Template) error {
	newTags := map[string]*foresttype.GraphTag{}
	for _, v := range template.Tags {
		_, ok := newTags[v.TagName]
		if !ok {
			// 如果不存在，创建一个新标签
			newTag := &foresttype.GraphTag{
				GraphID:        graph.ID,
				GraphVersionID: graph.VersionID,
				Description:    graph.Description,
				TagName:        v.TagName,
				TagType:        foresttype.TagTypeNode,
				Uin:            graph.Uin,
				CompanyID:      graph.CompanyID,
				Properties:     v.Properties,
			}
			newTags[v.TagName] = newTag
			err := CreateTag(ctx, graph.SpaceName, newTag)
			if err != nil {
				logs.ErrorContextf(ctx, "upsertTag CreateTag err: %v", err)
				return err
			}
		}
	}
	eMap := map[string]*foresttype.GraphTag{}
	for _, v := range template.Edges {
		if _, ok := newTags[v.Source]; !ok {
			continue
		}
		if _, ok := newTags[v.Target]; !ok {
			continue
		}
		edge, edgeOk := eMap[v.Name]
		if !edgeOk {
			_, ok := newTags[v.Name]
			if ok {
				// 如果存在跳出节点
				logs.WarnContextf(ctx, "InsertTemplate upsertEdge tag: %s already exists", v.Name)
				continue
			}
			newEdge := &foresttype.GraphTag{
				GraphID:        graph.ID,
				GraphVersionID: graph.VersionID,
				TagName:        v.Name,
				TagType:        foresttype.TagTypeEdge,
				Uin:            graph.Uin,
				CompanyID:      graph.CompanyID,
			}
			err := CreateTag(ctx, graph.SpaceName, newEdge)
			if err != nil {
				logs.ErrorContextf(ctx, "upsertEdge CreateTag err: %v", err)
				return err
			}
			et := &foresttype.GraphEdgeTag{
				GraphID:        graph.ID,
				GraphVersionID: graph.VersionID,
				EdgeTypeID:     newEdge.ID,
				SrcTagID:       newTags[v.Source].ID,
				DstTagID:       newTags[v.Target].ID,
			}
			err = dbutil.Knownow().Create(et).Error
			if err != nil {
				logs.ErrorContextf(ctx, "upsertEdge CreateEdgeType err: %v", err)
				return err
			}
			eMap[v.Name] = newEdge
			edge = newEdge
		}
		_, err := GetEdgeTag(ctx, edge.ID, newTags[v.Source].ID, newTags[v.Target].ID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				et := &foresttype.GraphEdgeTag{
					GraphID:        graph.ID,
					GraphVersionID: graph.VersionID,
					EdgeTypeID:     edge.ID,
					SrcTagID:       newTags[v.Source].ID,
					DstTagID:       newTags[v.Target].ID,
				}
				err = dbutil.Knownow().Create(et).Error
				if err != nil {
					logs.ErrorContextf(ctx, "InsertTemplate CreateEdgeType err: %v", err)
					return err
				}
				continue
			}
			logs.ErrorContextf(ctx, "InsertTemplate GetEdgeTag err: %v", err)
			return err
		}
	}

	return nil
}

// DeleteGraphTempData 删除图谱结构的临时数据
func DeleteGraphTempData(ctx context.Context, graph *foresttype.ForestGraphInfo) error {
	err := dbutil.Knownow().Transaction(func(tx *gorm.DB) error {
		err := tx.Where("graph_id = ?", graph.ID).Where("graph_version_id = ?", graph.VersionID).Delete(&foresttype.GraphTag{}).Error
		if err != nil {
			logs.ErrorContextf(ctx, "DeleteGraphTempData DeleteTag err: %v", err)
			return err
		}
		err = tx.Where("graph_id = ?", graph.ID).Where("graph_version_id = ?", graph.VersionID).Delete(&foresttype.GraphEdgeTag{}).Error
		if err != nil {
			logs.ErrorContextf(ctx, "DeleteGraphTempData DeleteEdge err: %v", err)
			return err
		}
		return nil
	})
	if err != nil {
		logs.ErrorContextf(ctx, "DeleteGraphTempData error: %v", err)
		return err
	}
	return nil
}

type Template struct {
	Tags  []tempTag  `json:"tags"`
	Edges []tempEdge `json:"edges"`
}

type tempTag struct {
	TagName     string                `json:"tag_name"`
	Description string                `json:"description"`
	Properties  foresttype.Properties `json:"properties"`
}

type tempEdge struct {
	Name   string `json:"name"`
	Source string `json:"source"`
	Target string `json:"target"`
}
