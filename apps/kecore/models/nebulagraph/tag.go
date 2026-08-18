package nebulagraph

import (
	"fmt"
	"strings"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

// CreateGraphTag 创建图数据库Tag或edge节点
func (cli *NebulaCli) CreateGraphTag(db *gorm.DB, tag *foresttype.GraphTag) error {
	if tag.TagStatus == foresttype.TagStatusSynced {
		return nil
	}
	logs.InfoContextf(cli.ctx, "create tag:%s", tag.CreateTagString())
	res, err := cli.ExecuteAndCheck(tag.CreateTagString())
	if err != nil {
		logs.ErrorContextf(cli.ctx, "fail to create tag:%s,err:%v", tag.CreateTagString(), err)
		return fmt.Errorf("CreateGraphTag ExecuteAndCheck err:%v,%v", tag.CreateTagString(), err)
	}
	logs.InfoContextf(cli.ctx, "create tag:%s", logs.JSON(res))
	tag.TagStatus = foresttype.TagStatusSynced
	err = db.Save(tag).Error
	if err != nil {
		logs.ErrorContextf(cli.ctx, "save tag:%s err:%v", tag.CreateTagString(), err)
		return err
	}
	return nil
}

// DropGraphTag 删除图数据库Tag或edge节点
func (cli *NebulaCli) DropGraphTag(tag *foresttype.GraphTag) error {
	res, err := cli.ExecuteAndCheck(fmt.Sprintf("DROP %s %s;", tag.TagType, tag.TagName))
	if err != nil {
		logs.ErrorContextf(cli.ctx, "fail to drop tag:%s,err:%v", tag.CreateTagString(), err)
		return fmt.Errorf("DropGraphTag ExecuteAndCheck err:%v,%v", tag.CreateTagString(), err)
	}
	logs.InfoContextf(cli.ctx, "drop tag:%s", logs.JSON(res))
	return nil
}

// CreateGraphTag 创建图数据库Tag或edge节点
func (cli *NebulaCli) AlterTagAdd(tag *foresttype.GraphTag, p *foresttype.Property) error {
	res, err := cli.ExecuteAndCheck(fmt.Sprintf("ALTER %s %s ADD (%s);", tag.TagType, tag.TagName, p.TagStr()))
	if err != nil {
		logs.ErrorContextf(cli.ctx, "fail to alter add tag:%s,err:%v", tag.CreateTagString(), err)
		return fmt.Errorf("AlterTagAdd ExecuteAndCheck err:%v,%v", tag.CreateTagString(), err)
	}
	logs.InfoContextf(cli.ctx, "alter add tag:%s", logs.JSON(res))
	tag.Properties = append(tag.Properties, p)
	return nil
}

// TODO tag属性的删除操作
// AlterTag 全量修改tag的属性
func (cli *NebulaCli) AlterTag(tag *foresttype.GraphTag, drop bool) error {
	pdesc, err := cli.GetTagDesc(tag)
	if err != nil {
		logs.ErrorContextf(cli.ctx, " GetTagDesc fail to alter tag:%s,err:%v", tag.TagName, err)
		return err
	}
	pdescMap := pdesc.NameMap()
	newMap := tag.Properties.NameMap()
	dropOps := []string{}
	change := []string{}
	add := []string{}

	for _, p := range tag.Properties {
		if field, ok := pdescMap[p.Name]; ok {
			// 判断是否一样修改属性
			if field.Type == p.Type && field.Default == p.Defaults && field.Comment == p.Comment {
				continue
			}
			change = append(change, p.TagStr())
		} else {
			add = append(add, p.TagStr())
		}
	}
	for oldName := range pdescMap {
		if _, ok := newMap[oldName]; !ok {
			dropOps = append(dropOps, fmt.Sprintf("`%s`", oldName))
		}
	}
	if len(change) > 0 {
		// 修改
		ngql := fmt.Sprintf("ALTER %s `%s` CHANGE (%s);", tag.TagType, tag.TagName, strings.Join(change, ", "))
		logs.InfoContextf(cli.ctx, "alter tag:%s change:%s", tag.TagName, ngql)
		_, err := cli.ExecuteAndCheck(ngql)
		if err != nil {
			logs.ErrorContextf(cli.ctx, "fail to alter CHANGE tag:%s,err:%v", ngql, err)
			return fmt.Errorf("AlterTag ExecuteAndCheck err:%v,%v", ngql, err)
		}
	}
	if len(add) > 0 {
		// 新增
		ngql := fmt.Sprintf("ALTER %s `%s` ADD (%s);", tag.TagType, tag.TagName, strings.Join(add, ", "))
		logs.InfoContextf(cli.ctx, "alter tag:%s add:%s", tag.TagName, ngql)
		_, err := cli.ExecuteAndCheck(ngql)
		if err != nil {
			logs.ErrorContextf(cli.ctx, "fail to alter tag:%s,err:%v", ngql, err)
			return fmt.Errorf("AlterTagAdd ExecuteAndCheck err:%v,%v", ngql, err)
		}
	}
	if drop && len(dropOps) > 0 {
		// 删除
		ngql := fmt.Sprintf("ALTER %s `%s` DROP (%s);", tag.TagType, tag.TagName, strings.Join(dropOps, ", "))
		logs.InfoContextf(cli.ctx, "alter tag:%s drop:%s", tag.TagName, ngql)
		_, err := cli.ExecuteAndCheck(ngql)
		if err != nil {
			logs.ErrorContextf(cli.ctx, "fail to alter tag:%s,err:%v", ngql, err)
			return fmt.Errorf("AlterTagAdd ExecuteAndCheck err:%v,%v", ngql, err)
		}
	}

	// time.Sleep(22 * time.Second)
	logs.InfoContextf(cli.ctx, "alter tag:%s", tag.TagName)
	return nil
}

// GetTagDesc 获取图数据库中的tag的属性信息
func (cli *NebulaCli) GetTagDesc(tag *foresttype.GraphTag) (TagDescList, error) {
	res, err := cli.ExecuteAndCheck(fmt.Sprintf("DESC %s `%s`;", tag.TagType, tag.TagName))
	if err != nil {
		logs.ErrorContextf(cli.ctx, "fail to alter add tag:%s,err:%v", tag.TagName, err)
		return nil, fmt.Errorf("AlterTagAdd ExecuteAndCheck err:%v,%v", tag.TagName, err)
	}
	properties := TagDescList{}
	err = res.Scan(&properties)
	if err != nil {
		logs.ErrorContextf(cli.ctx, "fail to alter add tag:%s,err:%v", tag.TagName, err)
		return nil, fmt.Errorf("AlterTagAdd ExecuteAndCheck err:%v,%v", tag.TagName, err)
	}
	logs.InfoContextf(cli.ctx, "alter add tag:%+v", properties)
	return properties, nil
}

// EdgeTypeExists 检查边类型是否存在
func (cli *NebulaCli) EdgeTypeExists(edgeName string) bool {
	checkSQL := fmt.Sprintf("DESC EDGE `%s`", edgeName)
	res, err := cli.ExecuteAndCheck(checkSQL)
	logs.InfoContextf(cli.ctx, "EdgeTypeExists: %s, res: %s", checkSQL, logs.JSON(res))
	return err == nil
}

// TagExists 检查节点标签类型是否存在
func (cli *NebulaCli) TagExists(tagName string) bool {
	checkSQL := fmt.Sprintf("DESC TAG `%s`", tagName)
	res, err := cli.ExecuteAndCheck(checkSQL)
	logs.InfoContextf(cli.ctx, "TagExists: %s, res: %s", checkSQL, logs.JSON(res))
	return err == nil
}
