package persist

import (
	"context"

	"github.com/casbin/casbin/v2/model"
)

// FilteredContextAdapter is the context-aware interface for Casbin adapters supporting filtered policies.
type FilteredContextAdapter interface {
	ContextAdapter

	// LoadFilteredPolicyCtx loads only policy rules that match the filter.
	LoadFilteredPolicyCtx(ctx context.Context, model model.Model, filter interface{}) error
	// IsFilteredCtx returns true if the loaded policy has been filtered.
	IsFilteredCtx(ctx context.Context) bool
}
