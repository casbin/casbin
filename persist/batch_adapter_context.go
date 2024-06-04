package persist

import "context"

// BatchContextAdapter is the context-aware interface for Casbin adapters with multiple add and remove policy functions.
type BatchContextAdapter interface {
	ContextAdapter
	// AddPoliciesCtx adds policy rules to the storage.
	// This is part of the Auto-Save feature.
	AddPoliciesCtx(ctx context.Context, sec string, ptype string, rules [][]string) error
	// RemovePoliciesCtx removes policy rules from the storage.
	// This is part of the Auto-Save feature.
	RemovePoliciesCtx(ctx context.Context, sec string, ptype string, rules [][]string) error
}
