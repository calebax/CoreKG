package svcmigrate

import (
	"context"
)

type BusinessType string

type Migrator interface {
	Run(ctx context.Context) error
}
