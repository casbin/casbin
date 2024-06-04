package persist

import "context"

// UpdatableContextAdapter is the context-aware interface for Casbin adapters with add update policy function.
type UpdatableContextAdapter interface {
	ContextAdapter
	// UpdatePolicyCtx updates a policy rule from storage.
	// This is part of the Auto-Save feature.
	UpdatePolicyCtx(ctx context.Context, sec string, ptype string, oldRule, newRule []string) error
	// UpdatePoliciesCtx updates some policy rules to storage, like db, redis.
	UpdatePoliciesCtx(ctx context.Context, sec string, ptype string, oldRules, newRules [][]string) error
	// UpdateFilteredPoliciesCtx deletes old rules and adds new rules.
	UpdateFilteredPoliciesCtx(ctx context.Context, sec string, ptype string, newRules [][]string, fieldIndex int, fieldValues ...string) ([][]string, error)
}
