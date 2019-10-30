package casbin

import (
	"github.com/casbin/casbin/effect"
	"github.com/casbin/casbin/model"
	"github.com/casbin/casbin/persist"
	"github.com/casbin/casbin/rbac"
)

type IEnforcer interface {

	/* Enforcer API`s */
	InitWithFile(modelPath string, policyPath string)
	InitWithAdapter(modelPath string, adapter persist.Adapter)
	InitWithModelAndAdapter(m model.Model, adapter persist.Adapter)
	LoadModel()
	GetModel() model.Model
	SetModel(m model.Model)
	GetAdapter() persist.Adapter
	SetAdapter(adapter persist.Adapter)
	SetWatcher(watcher persist.Watcher)
	SetRoleManager(rm rbac.RoleManager)
	SetEffector(eft effect.Effector)
	ClearPolicy()
	LoadPolicy() error
	LoadFilteredPolicy(filter interface{}) error
	IsFiltered() bool
	SavePolicy() error
	EnableEnforce(enable bool)
	EnableLog(enable bool)
	EnableAutoSave(autoSave bool)
	EnableAutoBuildRoleLinks(autoBuildRoleLinks bool)
	BuildRoleLinks()
	Enforce(rvals ...interface{}) bool

	/* RBAC API`s */
	GetRolesForUser(name string) ([]string, error)
	GetUsersForRole(name string) ([]string, error)
	HasRoleForUser(name string, role string) (bool, error)
	AddRoleForUser(user string, role string) bool
	AddPermissionForUser(user string, permission ...string) bool
	DeletePermissionForUser(user string, permission ...string) bool
	DeletePermissionsForUser(user string) bool
	GetPermissionsForUser(user string) [][]string
	HasPermissionForUser(user string, permission ...string) bool
	GetImplicitRolesForUser(name string, domain ...string) []string
	GetImplicitPermissionsForUser(user string, domain ...string) [][]string
	GetImplicitUsersForPermission(permission ...string) []string
	DeleteRoleForUser(user string, role string) bool
	DeleteRolesForUser(user string) bool
	DeleteUser(user string) bool
	DeleteRole(role string)
	DeletePermission(permission ...string) bool

	/* management API`s */
	GetAllSubjects() []string
	GetAllNamedSubjects(ptype string) []string
	GetAllObjects() []string
	GetAllNamedObjects(ptype string) []string
	GetAllActions() []string
	GetAllNamedActions(ptype string) []string
	GetAllRoles() []string
	GetAllNamedRoles(ptype string) []string
	GetPolicy() [][]string
	GetFilteredPolicy(fieldIndex int, fieldValues ...string) [][]string
	GetNamedPolicy(ptype string) [][]string
	GetFilteredNamedPolicy(ptype string, fieldIndex int, fieldValues ...string) [][]string
	GetGroupingPolicy() [][]string
	GetFilteredGroupingPolicy(fieldIndex int, fieldValues ...string) [][]string
	GetNamedGroupingPolicy(ptype string) [][]string
	GetFilteredNamedGroupingPolicy(ptype string, fieldIndex int, fieldValues ...string) [][]string
	HasPolicy(params ...interface{}) bool
	HasNamedPolicy(ptype string, params ...interface{}) bool
	AddPolicy(params ...interface{}) bool
	AddNamedPolicy(ptype string, params ...interface{}) bool
	RemovePolicy(params ...interface{}) bool
	RemoveFilteredPolicy(fieldIndex int, fieldValues ...string) bool
	RemoveNamedPolicy(ptype string, params ...interface{}) bool
	RemoveFilteredNamedPolicy(ptype string, fieldIndex int, fieldValues ...string) bool
	HasGroupingPolicy(params ...interface{}) bool
	HasNamedGroupingPolicy(ptype string, params ...interface{}) bool
	AddGroupingPolicy(params ...interface{}) bool
	AddNamedGroupingPolicy(ptype string, params ...interface{}) bool
	RemoveGroupingPolicy(params ...interface{}) bool
	RemoveFilteredGroupingPolicy(fieldIndex int, fieldValues ...string) bool
	RemoveNamedGroupingPolicy(ptype string, params ...interface{}) bool
	RemoveFilteredNamedGroupingPolicy(ptype string, fieldIndex int, fieldValues ...string) bool
	AddFunction(name string, function func(args ...interface{}) (interface{}, error))
}
