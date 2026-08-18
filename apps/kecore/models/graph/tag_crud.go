package graph

import (
	"context"
	"fmt"
	"strings"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/nebulagraph"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

// CreateTag 创建tag 暂不创建图数据库tag
func CreateTag(ctx context.Context, spaceName string, tag *foresttype.GraphTag) error {
	// cli, err := nebulagraph.NewNebulaCLI(ctx, spaceName)
	// if err != nil {
	// 	logs.ErrorContextf(ctx, "NewNebulaCLI error: %v", err)
	// 	return err
	// }
	// defer cli.Release()
	tag.Properties.Deduplicate()
	// err = cli.CreateGraphTag(tag)
	// if err != nil {
	// 	logs.ErrorContextf(ctx, "CreateGraphTag error: %v", err)
	// 	return err
	// }
	err := dbutil.Knownow().WithContext(ctx).Create(tag).Error
	if err != nil {
		logs.ErrorContextf(ctx, "CreateTag error: %v", err)
		return err
	}
	return nil
}

// CreateTag 创建tag
func CreateTagTx(ctx context.Context, db *gorm.DB, spaceName string, tag *foresttype.GraphTag) error {
	cli, err := nebulagraph.NewNebulaCLI(ctx, spaceName)
	if err != nil {
		logs.ErrorContextf(ctx, "NewNebulaCLI error: %v", err)
		return err
	}
	defer cli.Release()
	err = cli.CreateGraphTag(db, tag)
	if err != nil {
		logs.ErrorContextf(ctx, "CreateGraphTag error: %v", err)
		return err
	}
	err = db.WithContext(ctx).Create(tag).Error
	if err != nil {
		logs.ErrorContextf(ctx, "CreateGraph error: %v", err)
		return err
	}
	return nil
}

// GetTagIDMapByGraphID 获取所有tag的id和tag的map
func GetTagIDMapByGraphID(ctx context.Context, graphID uint) (map[uint]*foresttype.GraphTag, error) {
	var tags []*foresttype.GraphTag
	err := dbutil.Knownow().WithContext(ctx).Where("graph_id = ?", graphID).
		Where("tag_type = ?", foresttype.TagTypeNode).Find(&tags).Error
	if err != nil {
		logs.ErrorContextf(ctx, "GetTagIDMap error: %v", err)
		return nil, err
	}
	tagMap := make(map[uint]*foresttype.GraphTag)
	for _, tag := range tags {
		tagMap[tag.ID] = tag
	}
	return tagMap, nil
}

// GetTagIDMapByGraphID 获取所有tag的name和tag的map
func GetTagNameMapByGraphID(ctx context.Context, graphID, graphVersion uint) (map[string]*foresttype.GraphTag, error) {
	var tags []*foresttype.GraphTag
	err := dbutil.Knownow().WithContext(ctx).Where("graph_id = ?", graphID).
		Where("graph_version_id = ?", graphVersion).
		Where("tag_type = ?", foresttype.TagTypeNode).
		Find(&tags).Error
	if err != nil {
		logs.ErrorContextf(ctx, "GetTagNameMapByGraphID error: %v", err)
		return nil, err
	}
	tagMap := make(map[string]*foresttype.GraphTag)
	for _, tag := range tags {
		tagMap[tag.TagName] = tag
	}
	return tagMap, nil
}

// GetTagIDMapByGraphID 获取所有tag的name和tag的map
func GetEdgeNameMapByGraphID(ctx context.Context, graphID, graphVersion uint) (map[string]*foresttype.GraphTag, error) {
	var tags []*foresttype.GraphTag
	err := dbutil.Knownow().WithContext(ctx).Where("graph_id = ?", graphID).
		Where("graph_version_id = ?", graphVersion).
		Where("tag_type = ?", foresttype.TagTypeEdge).
		Find(&tags).Error
	if err != nil {
		logs.ErrorContextf(ctx, "GetTagNameMapByGraphID error: %v", err)
		return nil, err
	}
	tagMap := make(map[string]*foresttype.GraphTag)
	for _, tag := range tags {
		tagMap[tag.TagName] = tag
	}
	return tagMap, nil
}

// GetEdgeIDMapByGraphID 获取所有edge的id和tag的map
func GetEdgeIDMapByGraphID(ctx context.Context, graphID, graphVersion uint) (map[uint]*foresttype.GraphTag, error) {
	var tags []*foresttype.GraphTag
	err := dbutil.Knownow().WithContext(ctx).Where("graph_id = ?", graphID).
		Where("graph_version_id = ?", graphVersion).
		Where("tag_type = ?", foresttype.TagTypeEdge).Find(&tags).Error
	if err != nil {
		logs.ErrorContextf(ctx, "GetTagIDMap error: %v", err)
		return nil, err
	}
	tagMap := make(map[uint]*foresttype.GraphTag)
	for _, tag := range tags {
		tagMap[tag.ID] = tag
	}
	return tagMap, nil
}

// UpdateTag 更新tag
func UpdateTag(ctx context.Context, tag *foresttype.GraphTag) error {
	err := dbutil.Knownow().WithContext(ctx).Model(tag).Save(tag).Error
	if err != nil {
		logs.ErrorContextf(ctx, "UpdateTag error: %v", err)
		return err
	}
	return nil
}

// DeleteTag 删除tag
func DeleteTag(ctx context.Context, tagID uint) error {
	err := dbutil.Knownow().WithContext(ctx).Where("id = ?", tagID).Delete(&foresttype.GraphTag{}).Error
	if err != nil {
		logs.ErrorContextf(ctx, "DeleteTag error: %v", err)
		return err
	}
	return nil
}

// GetGraphTags 获取当前图的所有tag
func GetGraphTags(ctx context.Context, graphID, graphVersion uint) ([]*foresttype.GraphTag, error) {
	var tags []*foresttype.GraphTag
	err := dbutil.Knownow().WithContext(ctx).Where("graph_id = ?", graphID).
		Where("graph_version_id = ?", graphVersion).
		Find(&tags).Error
	if err != nil {
		logs.ErrorContextf(ctx, "GetGraphTags error: %v", err)
		return nil, err
	}
	return tags, nil
}

// GetTagByName 获取tag
func GetTagByName(ctx context.Context, graphID, graphVersion uint, name string) (*foresttype.GraphTag, error) {
	var tag foresttype.GraphTag
	err := dbutil.Knownow().WithContext(ctx).Where("graph_id = ?", graphID).
		Where("graph_version_id = ?", graphVersion).
		Where("tag_name = ?", name).
		Where("tag_type = ?", foresttype.TagTypeNode).
		First(&tag).Error
	if err != nil {
		logs.WarnContextf(ctx, "GetTagByName error: %v", err)
		return nil, err
	}
	return &tag, nil
}

// GetEdgeByName 获取边
func GetEdgeByName(ctx context.Context, graphID, graphVersion uint, name string) (*foresttype.GraphTag, error) {
	var tag foresttype.GraphTag
	err := dbutil.Knownow().WithContext(ctx).Where("graph_id = ?", graphID).
		Where("tag_name = ?", name).
		Where("graph_version_id = ?", graphVersion).
		Where("tag_type = ?", foresttype.TagTypeEdge).
		First(&tag).Error
	if err != nil {
		logs.WarnContextf(ctx, "GetTagByName error: %v", err)
		return nil, err
	}
	return &tag, nil
}

// GetEdgesByNames 批量获取边 tag，返回 map[edgeName]*GraphTag
func GetEdgesByNames(ctx context.Context, graphID, graphVersion uint, names []string) (map[string]*foresttype.GraphTag, error) {
	if len(names) == 0 {
		return make(map[string]*foresttype.GraphTag), nil
	}
	var tags []*foresttype.GraphTag
	err := dbutil.Knownow().WithContext(ctx).
		Where("graph_id = ?", graphID).
		Where("graph_version_id = ?", graphVersion).
		Where("tag_type = ?", foresttype.TagTypeEdge).
		Where("tag_name IN ?", names).
		Find(&tags).Error
	if err != nil {
		logs.WarnContextf(ctx, "GetEdgesByNames error: %v", err)
		return nil, err
	}
	result := make(map[string]*foresttype.GraphTag, len(tags))
	for _, tag := range tags {
		result[tag.TagName] = tag
	}
	return result, nil
}

// GetEdgeTag 获取边关系的TAG
func GetEdgeTag(ctx context.Context, edgeID, srcID, dstID uint) (*foresttype.GraphEdgeTag, error) {
	var edgeTag foresttype.GraphEdgeTag
	err := dbutil.Knownow().WithContext(ctx).
		Where("edge_type_id = ?", edgeID).
		Where("src_tag_id = ?", srcID).
		Where("dst_tag_id = ?", dstID).
		First(&edgeTag).Error
	if err != nil {
		logs.WarnContextf(ctx, "GetEdgeTag error: %v", err)
		return nil, err
	}
	return &edgeTag, nil
}

func UpdateEdgeTag(ctx context.Context, edgeTag *foresttype.GraphEdgeTag) error {
	err := dbutil.Knownow().WithContext(ctx).Model(edgeTag).Save(edgeTag).Error
	if err != nil {
		logs.ErrorContextf(ctx, "UpdateEdgeTag error: %v", err)
		return err
	}
	return nil
}

// GetEdgeTagByID 获取边关系的TAG
func GetEdgeTagByID(ctx context.Context, id uint) (*foresttype.GraphEdgeTag, error) {
	var edgeTag foresttype.GraphEdgeTag
	err := dbutil.Knownow().WithContext(ctx).
		Where("id = ?", id).
		First(&edgeTag).Error
	if err != nil {
		logs.WarnContextf(ctx, "GetEdgeTagByID error: %v", err)
		return nil, err
	}
	return &edgeTag, nil
}

func DeleteEdgeTag(ctx context.Context, id uint) error {
	err := dbutil.Knownow().WithContext(ctx).
		Where("id = ?", id).
		Delete(&foresttype.GraphEdgeTag{}).Error
	if err != nil {
		logs.WarnContextf(ctx, "DeleteEdgeTag error: %v", err)
		return err
	}
	return nil
}

func ListEdgeTag(ctx context.Context, graphID, graphVersion uint) ([]*foresttype.GraphEdgeTag, error) {
	var edgeTags []*foresttype.GraphEdgeTag
	err := dbutil.Knownow().WithContext(ctx).Where("graph_id = ?", graphID).
		Where("graph_version_id = ?", graphVersion).
		Find(&edgeTags).Error
	if err != nil {
		logs.ErrorContextf(ctx, "ListEdgeTag error: %v", err)
		return nil, err
	}
	return edgeTags, nil
}

func ListEdgeTagByEdgeID(ctx context.Context, edgeID uint) ([]*foresttype.GraphEdgeTag, error) {
	var edgeTags []*foresttype.GraphEdgeTag
	err := dbutil.Knownow().WithContext(ctx).Where("edge_type_id = ?", edgeID).
		Find(&edgeTags).Error
	if err != nil {
		logs.ErrorContextf(ctx, "ListEdgeTagByEdgeID error: %v", err)
		return nil, err
	}
	return edgeTags, nil
}

// ListTag
func ListTag(ctx context.Context, graphID, graphVersion uint, opt apiobj.PageQuery) (*TagInfoList, error) {
	queryList := &TagInfoList{}
	sql := dbutil.Knownow().WithContext(ctx).
		Table(foresttype.GraphTag{}.TableName()+" AS tag").
		// Select("tag.*").
		Where("tag.deleted_at IS NULL").
		Where("tag.graph_id = ?", graphID).
		Where("tag.graph_version_id = ?", graphVersion).
		Where("tag.company_id = ?", opt.CompanyID)
	for _, filter := range opt.Filters {
		switch filter.Field {
		case "uin":
			sql = sql.Where("tag.uin = ?", filter.Value[0])
		case "tag_type":
			sql = sql.Where("tag.tag_type = ?", filter.Value[0])
		case "name":
			sql = sql.Where("tag.tag_name LIKE ?", fmt.Sprintf("%%%s%%", filter.Value[0]))
		default:
			logs.ErrorContextf(ctx, "ListTag invalid filtering field: %s", filter.Field)
			return nil, fmt.Errorf("invalid filtering field: %s", filter.Field)
		}
	}
	// 处理 BeginTime 和 EndTime
	if !opt.BeginTime.IsZero() {
		sql = sql.Where("tag.created_at >= ?", opt.BeginTime)
	}
	if !opt.EndTime.IsZero() {
		sql = sql.Where("tag.created_at <= ?", opt.EndTime)
	}

	if err := sql.Count(&queryList.Total).Error; err != nil {
		logs.ErrorContextf(ctx, "ListTag Statistical project failed: %v", err)
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
		logs.ErrorContextf(ctx, "ListTag Retrieval project failed: %v", err)
		return nil, err
	}
	queryList.Limit = opt.Limit
	queryList.Offset = opt.Offset

	return queryList, nil
}

// TagInfoList
type TagInfoList struct {
	apiobj.QueryResponse
	Data []*TagInfo
}

// TagInfo TagInfo
type TagInfo struct {
	foresttype.GraphTag
}

// GetTagByID 获取tag
func GetTagByID(ctx context.Context, tagID uint) (*foresttype.GraphTag, error) {
	var tag foresttype.GraphTag
	err := dbutil.Knownow().WithContext(ctx).
		Where("id = ?", tagID).First(&tag).Error
	if err != nil {
		logs.WarnContextf(ctx, "GetTagByID error: %v", err)
		return nil, err
	}
	return &tag, nil
}
