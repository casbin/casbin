package rbac

import (
	"context"
)

// ContextRoleManager provides a context-aware interface to define the operations for managing roles.
// Prefer this over RoleManager interface for context propagation, which is useful for things like handling
// request timeouts.
type ContextRoleManager interface {
	RoleManager
	// ClearCtx clears all stored data and resets the role manager to the initial state with context.
	ClearCtx(ctx context.Context) error
	// AddLinkCtx adds the inheritance link between two roles. role: name1 and role: name2 with context.
	// domain is a prefix to the roles (can be used for other purposes).
	AddLinkCtx(ctx context.Context, name1 string, name2 string, domain ...string) error
	// DeleteLinkCtx deletes the inheritance link between two roles. role: name1 and role: name2 with context.
	// domain is a prefix to the roles (can be used for other purposes).
	DeleteLinkCtx(ctx context.Context, name1 string, name2 string, domain ...string) error
	// HasLinkCtx determines whether a link exists between two roles. role: name1 inherits role: name2 with context.
	// domain is a prefix to the roles (can be used for other purposes).
	HasLinkCtx(ctx context.Context, name1 string, name2 string, domain ...string) (bool, error)
	// GetRolesCtx gets the roles that a user inherits with context.
	// domain is a prefix to the roles (can be used for other purposes).
	GetRolesCtx(ctx context.Context, name string, domain ...string) ([]string, error)
	// GetUsersCtx gets the users that inherits a role with context.
	// domain is a prefix to the users (can be used for other purposes).
	GetUsersCtx(ctx context.Context, name string, domain ...string) ([]string, error)
	// GetDomainsCtx gets domains that a user has with context.
	GetDomainsCtx(ctx context.Context, name string) ([]string, error)
	// GetAllDomainsCtx gets all domains with context.
	GetAllDomainsCtx(ctx context.Context) ([]string, error)
}
