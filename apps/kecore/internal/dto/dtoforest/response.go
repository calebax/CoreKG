package dtoforest

import (
	"fmt"
	"sort"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type GetResourceBaseInfoResponse struct {
	apiobj.BaseResponse
	Response GetResourceBaseInfoEmbedResponse
}

type GetResourceBaseInfoEmbedResponse struct {
	// Tree 资源树
	Tree []*ForestResourceTreeNode `json:"tree"`
}

type ResourceBasInfoItem struct {
	// ID 资源 id
	ID string `json:"id"`
	// Name 资源名称
	Name string `json:"name"`
}

type ForestResourceTreeNode struct {
	// ForestID 知识库 id
	ForestID uint `json:"forest_id"`
	// ID id
	ID uint `json:"id"`
	// ParentID 父节点 id
	ParentID uint `json:"parent_id"`
	// Key 节点唯一 Key
	Key string `json:"key"`
	// ParentKey 父节点key
	ParentKey string `json:"parent_key"`
	// Name 名称
	Name string `json:"name"`
	// NodeType 节点类型，forest、dir、file、mysql_table、excel_sheet
	NodeType string `json:"node_type"`
	// ForestType 知识库类型，mysql、excel
	ForestType foresttype.ForestType `json:"forest_type,omitempty"`
	// ForestDataSourceType 知识库数据源类型
	ForestDataSourceType foresttype.ForestDataSourceType `json:"forest_data_source_type,omitempty"`
	// ForestDataSourceSubtype 知识库数据源子类型
	ForestDataSourceSubtype foresttype.ForestDataSourceSubtype `json:"forest_data_source_subtype,omitempty"`
	// Children 子节点
	Children []*ForestResourceTreeNode `json:"children"`
}

func BuildTree(nodes []*ForestResourceTreeNode) []*ForestResourceTreeNode {
	nodeMap := make(map[string]*ForestResourceTreeNode) // 创建一个映射，便于查找节点
	var roots []*ForestResourceTreeNode                 // 存储所有的根节点

	// 创建节点映射，方便查找
	for i := range nodes {
		node := nodes[i]
		node.Children = []*ForestResourceTreeNode{} // 初始化子节点列表
		nodeMap[node.Key] = node                    // 将节点添加到映射中
	}

	// 构建树结构
	for i := range nodes {
		node := nodes[i]
		if node.ParentKey == "" {
			// 该节点是根节点
			roots = append(roots, node)
		} else {
			// 该节点有父节点，将其添加到父节点的子节点列表中
			if parent, ok := nodeMap[node.ParentKey]; ok {
				parent.Children = append(parent.Children, node)
			} else {
				// 处理父节点不存在的情况
				fmt.Printf("警告: 节点 ID %s 的父节点 ID %s 未找到\n", node.Key, node.ParentKey)
			}
		}
	}

	// 对根节点按 ID 排序
	sort.Slice(roots, func(i, j int) bool {
		return roots[i].ID < roots[j].ID
	})

	// 递归对每个节点的子节点按 ID 排序
	sortChildren(roots)

	return roots
}

// sortChildren 递归对所有节点的子节点按 ID 排序
func sortChildren(nodes []*ForestResourceTreeNode) {
	for _, node := range nodes {
		if len(node.Children) > 0 {
			// 对当前节点的子节点按 ID 排序
			sort.Slice(node.Children, func(i, j int) bool {
				return node.Children[i].ID < node.Children[j].ID
			})
			// 递归处理子节点
			sortChildren(node.Children)
		}
	}
}

type SetResourceEnableResponse struct {
	apiobj.BaseResponse
	Response SetResourceEnableEmbedResponse
}
type SetResourceEnableEmbedResponse struct {
}

type UpdateForestDescriptionResponse struct {
	apiobj.BaseResponse
	Response UpdateForestDescriptionEmbedResponse
}
type UpdateForestDescriptionEmbedResponse struct {
}

type GetOriginResourceResponse struct {
	apiobj.BaseResponse
	Response GetResourceURLListEmbedResponse
}

type Resource struct {
	// ID 资源 id
	ID uint `json:"id"`
	//Meta 资源元数据
	Meta map[string]interface{} `json:"meta"`
	//	ResourceType 资源类型
	ResourceType foresttype.ResourceType `json:"resource_type"`
}

type GetResourceURLListEmbedResponse struct {
	Data []*Resource `json:"data"`
}
