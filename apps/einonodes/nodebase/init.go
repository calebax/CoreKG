package nodebase

import "github.com/cloudwego/eino/compose"

func init() {
	compose.RegisterValuesMergeFunc(func(inputs []RecordList) (RecordList, error) {
		var merged RecordList
		for _, rl := range inputs {
			merged = append(merged, rl...)
		}
		return merged, nil
	})
}
