package rbac

type Role string

const (
	RoleAdmin     Role = "admin"
	RoleDoctor    Role = "doctor"
	RoleNurse     Role = "nurse"
	RolePatient   Role = "patient"
	RoleCaregiver Role = "caregiver"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RoleDoctor, RoleNurse, RolePatient, RoleCaregiver:
		return true
	default:
		return false
	}
}
