package rbac

// Role defines a set of permissions
type Role struct {
	Name        string        `json:"name" yaml:"name"`
	Description string        `json:"description,omitempty" yaml:"description,omitempty"`
	Permissions []*Permission `json:"permissions" yaml:"permissions"`
}

// Permission defines an allowed operation
type Permission struct {
	Resource string   `json:"resource" yaml:"resource"`
	Actions  []Action `json:"actions" yaml:"actions"`
}

// Action represents an operation that can be performed
type Action string

const (
	ActionCreate Action = "create"
	ActionRead   Action = "read"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
	ActionStart  Action = "start"
	ActionStop   Action = "stop"
	ActionPull   Action = "pull"
	ActionPush   Action = "push"
	ActionExec   Action = "exec"
	ActionAll    Action = "*"
)

// HasPermission checks if role has a specific permission
func (r *Role) HasPermission(resource string, action Action) bool {
	for _, perm := range r.Permissions {
		// Check wildcard resource
		if perm.Resource == "*" || perm.Resource == resource {
			// Check actions
			for _, a := range perm.Actions {
				if a == "*" || a == action {
					return true
				}
			}
		}
	}
	return false
}

// Built-in role constants
const (
	RoleAdmin     = "admin"
	RoleDeveloper = "developer"
	RoleOperator  = "operator"
	RoleViewer    = "viewer"
)

// GetBuiltinRole returns a built-in role definition
func GetBuiltinRole(name string) *Role {
	switch name {
	case RoleAdmin:
		return &Role{
			Name:        RoleAdmin,
			Description: "Full administrative access",
			Permissions: []*Permission{
				{
					Resource: "*",
					Actions:  []Action{ActionAll},
				},
			},
		}
	case RoleDeveloper:
		return &Role{
			Name:        RoleDeveloper,
			Description: "Developer access for container management",
			Permissions: []*Permission{
				{
					Resource: "container",
					Actions:  []Action{ActionCreate, ActionRead, ActionUpdate, ActionDelete, ActionStart, ActionStop, ActionExec},
				},
				{
					Resource: "image",
					Actions:  []Action{ActionRead, ActionPull},
				},
				{
					Resource: "volume",
					Actions:  []Action{ActionCreate, ActionRead, ActionDelete},
				},
				{
					Resource: "network",
					Actions:  []Action{ActionRead},
				},
			},
		}
	case RoleOperator:
		return &Role{
			Name:        RoleOperator,
			Description: "Operator access for container operations",
			Permissions: []*Permission{
				{
					Resource: "container",
					Actions:  []Action{ActionRead, ActionStart, ActionStop},
				},
				{
					Resource: "image",
					Actions:  []Action{ActionRead},
				},
				{
					Resource: "volume",
					Actions:  []Action{ActionRead},
				},
			},
		}
	case RoleViewer:
		return &Role{
			Name:        RoleViewer,
			Description: "Read-only access",
			Permissions: []*Permission{
				{
					Resource: "*",
					Actions:  []Action{ActionRead},
				},
			},
		}
	default:
		return nil
	}
}
