package rbac

type Role string

const (
	RoleAdmin     Role = "admin"
	RoleDoctor    Role = "doctor"
	RoleNurse     Role = "nurse"
	RoleCaregiver Role = "caregiver"
	RoleFamily    Role = "family"
	RoleOwner     Role = "owner"
)
