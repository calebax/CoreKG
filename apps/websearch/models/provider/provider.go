package provider

import (
	"context"

	"github.com/insmtx/corekg/apps/websearch/models/domain"
)

type Provider interface {
	Name() domain.ProviderName
	Search(context.Context, domain.SearchRequest) (domain.SearchResponse, error)
}
