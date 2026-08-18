package types

import (
	"time"
)

type Model struct {
	ID        string     `gorm:"column:id;size:32;primarykey"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime"`
	DeletedAt *time.Time `gorm:"index"`
}

func GenerateModel() Model {
	return Model{
		ID: GenerateID(),
	}
}
