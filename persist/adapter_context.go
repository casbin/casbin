package persist

import (
	"context"

	"github.com/casbin/casbin/v2/model"
)

// ContextAdapter provides a context-aware interface for Casbin adapters.
type ContextAdapter interface {
	// LoadPolicy loads all policy rules from the storage.
	LoadPolicy(ctx context.Context, model model.Model) error
	// SavePolicy saves all policy rules to the storage.
	SavePolicy(ctx context.Context, model model.Model) error

	// AddPolicy adds a policy rule to the storage.
	// This is part of the Auto-Save feature.
	AddPolicy(ctx context.Context, sec string, ptype string, rule []string) error
	// RemovePolicy removes a policy rule from the storage.
	// This is part of the Auto-Save feature.
	RemovePolicy(ctx context.Context, sec string, ptype string, rule []string) error
	// RemoveFilteredPolicy removes policy rules that match the filter from the storage.
	// This is part of the Auto-Save feature.
	RemoveFilteredPolicy(ctx context.Context, sec string, ptype string, fieldIndex int, fieldValues ...string) error
}
