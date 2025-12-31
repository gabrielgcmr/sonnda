package patientaccess

type Permission string

const (
	PermPatientRead         Permission = "patient:read"
	PermPatientWrite        Permission = "patient:write"
	PermMedicalProblemWrite Permission = "medical_record:problem:write"
	PermMedicalPrevention   Permission = "medical_record:prevention:write"
	PermLabsRead            Permission = "medical_record:labs:read"
	PermLabsUpload          Permission = "medical_record:labs:upload"
)

func defaultPermissionsForRole(role MemberRole) []Permission {
	switch role {
	case RoleProfessional:
		return []Permission{
			PermPatientRead,
			PermPatientWrite,
			PermMedicalProblemWrite,
			PermMedicalPrevention,
			PermLabsRead,
			PermLabsUpload,
		}
	case RoleCaregiver:
		return []Permission{
			PermPatientRead,
			PermPatientWrite,
			PermLabsRead,
			PermLabsUpload,
		}
	default:
		return nil
	}
}
