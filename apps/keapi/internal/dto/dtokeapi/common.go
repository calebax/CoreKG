package dtokeapi

import "github.com/ygpkg/yg-go/apis/apiobj"

// NormalizePageQuery 归一化分页参数，并补齐默认排序字段。
func NormalizePageQuery(p *apiobj.PageQuery, defaultOrderBy string) {
	p.Fill(nil)
	if len(p.OrderBy) == 0 && defaultOrderBy != "" {
		p.OrderBy = []string{defaultOrderBy}
	}
}
