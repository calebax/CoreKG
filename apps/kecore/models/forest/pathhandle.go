package forest

import (
	"strings"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
)

// SplitPath 拆解输入的path路径,返回parents列表和当前文件名称（目录包含/）
func SplitPath(p string) ([]string, string) {
	var (
		parents []string
		current string
	)
	if p == "" {
		return parents, current
	}

	items := strings.Split(p, "/")
	if !IsDirPath(p) {
		parents = items[:len(items)-1]
		current = items[len(items)-1]
	} else {
		parents = items[:len(items)-2]
		current = items[len(items)-2] + "/"
	}
	for i, parent := range parents {
		parents[i] = parent + "/"
	}

	return parents, current
}

// GetForestParentFileByPath 获取路径的父级目录，如果没有，返回nil
func GetForestParentFileByPath(forestID uint, p string) (*foresttype.KnownowForestFile, error) {
	parents, _ := SplitPath(p)
	if len(parents) == 0 {
		return nil, nil
	}
	var (
		parent        = &foresttype.KnownowForestFile{}
		err           error
		queryParentID uint = 0
	)

	for _, pname := range parents {
		parent, err = GetForestDirByParentAndName(forestID, queryParentID, pname)
		if err != nil {
			return nil, err
		}
		queryParentID = parent.ID
	}
	return parent, nil
}

// GetForestFileByPath 通过路径查询文件
func GetForestFileByPath(forestID uint, p string) (*foresttype.KnownowForestFile, error) {
	_, fname := SplitPath(p)
	parent, err := GetForestParentFileByPath(forestID, p)
	if err != nil {
		return nil, err
	}
	// 有父级目录
	var parentID uint = 0
	if parent != nil {
		parentID = parent.ID
	}
	return GetForestFileByName(forestID, parentID, fname)
}
