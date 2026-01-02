package patientaccess

import "errors"

type RequestStatus string

const (
	RequestPending   RequestStatus = "pending"  // solicitado, aguardando ação
	RequestApproved  RequestStatus = "approved" // aprovado (opcional; pode ir direto para grant)
	RequestRejected  RequestStatus = "rejected"
	RequestCancelled RequestStatus = "cancelled"
	RequestExpired   RequestStatus = "expired"
)

func (s RequestStatus) IsValid() bool {
	switch s {
	case RequestPending, RequestApproved, RequestRejected, RequestCancelled, RequestExpired:
		return true
	default:
		return false
	}
}

var ErrInvalidRequestStatus = errors.New("invalid request status")
