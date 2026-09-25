// internal/features/account/onboarding_test.go
package account

import (
	"context"
	"testing"

	accountdomain "github.com/gabrielgcmr/sonnda/internal/features/account/domain"
)

type onboardingRepository struct {
	Repository
}

func (r onboardingRepository) FindByAuthIdentity(context.Context, string, string) (*accountdomain.User, error) {
	return nil, nil
}

type onboardingUserService struct {
	Service
	input UserCreateInput
}

func (s *onboardingUserService) Create(_ context.Context, input UserCreateInput) (*accountdomain.User, error) {
	s.input = input
	return &accountdomain.User{AccountType: input.AccountType, FullName: input.FullName}, nil
}

func TestOnboardingRegistersWithoutProfessionalProfile(t *testing.T) {
	for _, accountType := range []accountdomain.AccountType{
		accountdomain.AccountTypeProfessional,
		accountdomain.AccountTypeBasicCare,
	} {
		t.Run(string(accountType), func(t *testing.T) {
			service := &onboardingUserService{}
			onboarding := NewOnboarding(onboardingRepository{}, service)
			created, err := onboarding.Register(context.Background(), RegisterInput{
				Issuer: "issuer", Subject: "subject", Email: "person@example.com",
				FullName: "Pessoa Teste", AccountType: accountType,
			})
			if err != nil {
				t.Fatalf("register: %v", err)
			}
			if created.AccountType != accountType || created.FullName != "Pessoa Teste" {
				t.Fatalf("unexpected user: %+v", created)
			}
			if service.input.Issuer != "issuer" || service.input.Subject != "subject" || service.input.Email != "person@example.com" {
				t.Fatalf("identity was not forwarded: %+v", service.input)
			}
		})
	}
}
