// internal/domain/entities/patientaccess/policy.go
package patientaccess

type AccessPolicy interface {
	// métodos que o serviço espera
}

type DefaultAccessPolicy struct{}

func New() AccessPolicy {
	return &DefaultAccessPolicy{}
}
