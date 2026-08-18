package essearch

import "github.com/insmtx/corekg/apps/ketask/models/ragtypes"

// SearchResult es 查询所用结构体
type SearchResult struct {
	ScrollID string `json:"_scroll_id"`
	Hits     struct {
		MaxScore float64 `json:"max_score"`
		Total    struct {
			Value int64 `json:"value"`
		} `json:"total"`
		Hits []Hits `json:"hits"`
	} `json:"hits"`
}

type Hits struct {
	Source    ragtypes.Chunk `json:"_source"`
	HighLight struct {
		Description []string `json:"description"`
	} `json:"highlight"`
	ID    string  `json:"_id"`
	Score float64 `json:"_score"`
}

// FileDescResult .
type FileDescResult struct {
	Hits struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		Hits []struct {
			Source *ragtypes.FileDescription `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}
