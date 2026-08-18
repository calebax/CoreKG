package forestkeywords

import "github.com/insmtx/corekg/apps/kecore/models/foresttype"

type SynonymKeywordDetail struct {
	foresttype.Keywords
	UserName        string                `json:"user_name"`
	SynonymKeywords []foresttype.Keywords `json:"synonym_keywords"`
}

type MajorKeywordDetail struct {
	foresttype.Keywords
	UserName string `json:"user_name"`
}
