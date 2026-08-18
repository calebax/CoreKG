package keqa

import (
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
)

func GetForestQA(id uint) (*foresttype.KnownowForestQA, error) {
	out := &foresttype.KnownowForestQA{}
	if err := dbutil.Knownow().
		First(out, id).Error; err != nil {
		return nil, err
	}
	return out, nil
}

type PureQAItem struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

const MaxUploadQAItem = 100
