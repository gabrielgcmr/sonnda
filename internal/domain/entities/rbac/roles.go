package rbac

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleDoctor Role = "doctor"
	RoleNurse  Role = "nurse"
	RoleCommon Role = "common"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RoleDoctor, RoleNurse, RoleCommon:
		return true
	default:
		return false
	}
}
