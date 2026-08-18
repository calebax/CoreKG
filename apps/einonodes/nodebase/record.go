package nodebase

// 极简版本 - 如果你坚持最简单的实现
type Record struct {
	Key   string         `json:"key"`
	Value string         `json:"value"`
	Extra map[string]any `json:"extra,omitempty"` // 至少添加 omitempty
}

type RecordList []*Record

func (rl *RecordList) Add(record *Record) {
	if record == nil { // 至少检查 nil
		return
	}
	if *rl == nil {
		*rl = make([]*Record, 0)
	}
	*rl = append(*rl, record)
}

func (rl *RecordList) Get(key string) *Record {
	if *rl == nil {
		return &Record{}
	}
	for _, v := range *rl {
		if v.Key == key {
			return v
		}
	}
	return &Record{}
}
