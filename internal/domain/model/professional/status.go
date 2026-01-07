package professional

type VerificationStatus string

const (
	StatusPending  VerificationStatus = "pending"
	StatusVerified VerificationStatus = "verified"
	StatusRejected VerificationStatus = "rejected"
)

func (s VerificationStatus) IsValid() bool {
	switch s {
	case StatusPending, StatusVerified, StatusRejected:
		return true
	default:
		return false
	}
}
