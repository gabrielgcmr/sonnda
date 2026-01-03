package usersvc

import (
	"context"
	"errors"
	"testing"
	"time"

	"sonnda-api/internal/domain/model/identity"
	"sonnda-api/internal/domain/model/rbac"
	"sonnda-api/internal/domain/model/user"

	"github.com/google/uuid"
)

type fakeUserRepo struct {
	findByEmailRes *user.User
	findByEmailErr error

	findByAuthRes *user.User
	findByAuthErr error

	findByCPFRes *user.User
	findByCPFErr error

	findByIDRes *user.User
	findByIDErr error

	saveErr error

	updateErr error

	softDeleteErr error

	calledFindByEmail bool
	calledFindByAuth  bool
	calledFindByCPF   bool
	calledFindByID    bool
	calledSave        bool
	calledUpdate      bool
	calledSoftDelete  bool

	saved       []*user.User
	updated     []*user.User
	softDeleted []uuid.UUID
}

func (r *fakeUserRepo) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	r.calledFindByEmail = true
	return r.findByEmailRes, r.findByEmailErr
}

func (r *fakeUserRepo) FindByAuthIdentity(ctx context.Context, provider, subject string) (*user.User, error) {
	r.calledFindByAuth = true
	return r.findByAuthRes, r.findByAuthErr
}

func (r *fakeUserRepo) FindByCPF(ctx context.Context, cpf string) (*user.User, error) {
	r.calledFindByCPF = true
	return r.findByCPFRes, r.findByCPFErr
}

func (r *fakeUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	r.calledFindByID = true
	return r.findByIDRes, r.findByIDErr
}

func (r *fakeUserRepo) Save(ctx context.Context, u *user.User) error {
	r.calledSave = true
	r.saved = append(r.saved, u)
	return r.saveErr
}

func (r *fakeUserRepo) Update(ctx context.Context, u *user.User) error {
	r.calledUpdate = true
	r.updated = append(r.updated, u)
	return r.updateErr
}

func (r *fakeUserRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	r.calledSoftDelete = true
	r.softDeleted = append(r.softDeleted, id)
	return r.softDeleteErr
}

type fakeIdentitySvc struct {
	providerName string
	disableErr   error

	calledDisable bool
	disabledSubj  []string
}

func (s *fakeIdentitySvc) ProviderName() string {
	if s.providerName != "" {
		return s.providerName
	}
	return "firebase"
}

func (s *fakeIdentitySvc) VerifyToken(ctx context.Context, tokenStr string) (*identity.Identity, error) {
	return nil, errors.New("not implemented")
}

func (s *fakeIdentitySvc) DisableUser(ctx context.Context, subject string) error {
	s.calledDisable = true
	s.disabledSubj = append(s.disabledSubj, subject)
	return s.disableErr
}

func validRegisterInput() RegisterInput {
	return RegisterInput{
		Provider:  "firebase",
		Subject:   "sub-123",
		Email:     "person@example.com",
		Role:      rbac.RoleCaregiver,
		FullName:  "Pessoa Teste",
		BirthDate: time.Date(1990, 1, 2, 0, 0, 0, 0, time.UTC),
		CPF:       "52998224725",
		Phone:     "11999998888",
	}
}

func TestService_Register_EmailAlreadyExists(t *testing.T) {
	repo := &fakeUserRepo{
		findByEmailRes: &user.User{ID: uuid.Must(uuid.NewV7())},
	}
	svc := New(repo, nil, nil)

	u, err := svc.Register(context.Background(), validRegisterInput())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, user.ErrEmailAlreadyExists) {
		t.Fatalf("expected ErrEmailAlreadyExists, got %v", err)
	}
	if u != nil {
		t.Fatalf("expected nil user, got %v", u)
	}
	if repo.calledSave {
		t.Fatalf("expected Save not to be called")
	}
}

func TestService_Register_FindByEmailError(t *testing.T) {
	sentinel := errors.New("db down")
	repo := &fakeUserRepo{findByEmailErr: sentinel}
	svc := New(repo, nil, nil)

	u, err := svc.Register(context.Background(), validRegisterInput())
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	if u != nil {
		t.Fatalf("expected nil user, got %v", u)
	}
	if repo.calledSave {
		t.Fatalf("expected Save not to be called")
	}
}

func TestService_Register_AuthIdentityAlreadyExists(t *testing.T) {
	repo := &fakeUserRepo{
		findByEmailRes: nil,
		findByAuthRes:  &user.User{ID: uuid.Must(uuid.NewV7())},
	}
	svc := New(repo, nil, nil)

	u, err := svc.Register(context.Background(), validRegisterInput())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, user.ErrAuthIdentityAlreadyExists) {
		t.Fatalf("expected ErrAuthIdentityAlreadyExists, got %v", err)
	}
	if u != nil {
		t.Fatalf("expected nil user, got %v", u)
	}
	if repo.calledSave {
		t.Fatalf("expected Save not to be called")
	}
}

func TestService_Register_CreateError(t *testing.T) {
	sentinel := errors.New("insert failed")
	repo := &fakeUserRepo{saveErr: sentinel}
	svc := New(repo, nil, nil)

	u, err := svc.Register(context.Background(), validRegisterInput())
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	if u != nil {
		t.Fatalf("expected nil user, got %v", err)
	}
	if !repo.calledSave {
		t.Fatalf("expected Save to be called")
	}
	if len(repo.saved) != 1 {
		t.Fatalf("expected 1 Save call, got %d", len(repo.saved))
	}
}

func TestService_Register_Success(t *testing.T) {
	repo := &fakeUserRepo{}
	svc := New(repo, nil, nil)

	u, err := svc.Register(context.Background(), validRegisterInput())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if u == nil {
		t.Fatalf("expected user, got nil")
	}
	if !repo.calledSave {
		t.Fatalf("expected Save to be called")
	}
	if len(repo.saved) != 1 {
		t.Fatalf("expected 1 Save call, got %d", len(repo.saved))
	}
	if repo.saved[0] != u {
		t.Fatalf("expected returned user to be the same instance passed to repo.Save")
	}
	if u.Email != "person@example.com" {
		t.Fatalf("expected email to match input")
	}
}

func baseUserForUpdate() *user.User {
	return &user.User{
		ID:           uuid.Must(uuid.NewV7()),
		AuthProvider: "firebase",
		AuthSubject:  "sub-123",
		Email:        "person@example.com",
		Role:         "caregiver",
		FullName:     "Nome Antigo",
		BirthDate:    time.Date(1990, 1, 2, 0, 0, 0, 0, time.UTC),
		CPF:          "52998224725",
		Phone:        "11999999999",
		CreatedAt:    time.Now().Add(-24 * time.Hour),
		UpdatedAt:    time.Now().Add(-1 * time.Hour),
	}
}

func TestService_Update_UserNotFound(t *testing.T) {
	repo := &fakeUserRepo{findByIDRes: nil}
	svc := New(repo, nil, nil)

	out, err := svc.Update(context.Background(), UpdateInput{UserID: uuid.Nil})
	if !errors.Is(err, user.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
	if out != nil {
		t.Fatalf("expected nil user, got %v", out)
	}
	if repo.calledUpdate {
		t.Fatalf("expected Update not to be called")
	}
}

func TestService_Update_FindByIDError(t *testing.T) {
	sentinel := errors.New("db down")
	repo := &fakeUserRepo{findByIDErr: sentinel}
	svc := New(repo, nil, nil)

	userID := uuid.Must(uuid.NewV7())
	out, err := svc.Update(context.Background(), UpdateInput{UserID: userID})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	if out != nil {
		t.Fatalf("expected nil user, got %v", out)
	}
	if repo.calledUpdate {
		t.Fatalf("expected Update not to be called")
	}
}

func TestService_Update_InvalidFullName(t *testing.T) {
	u := baseUserForUpdate()
	repo := &fakeUserRepo{findByIDRes: u}
	svc := New(repo, nil, nil)

	bad := "   "
	userID := uuid.Must(uuid.NewV7())
	out, err := svc.Update(context.Background(), UpdateInput{UserID: userID, FullName: &bad})
	if !errors.Is(err, user.ErrInvalidFullName) {
		t.Fatalf("expected ErrInvalidFullName, got %v", err)
	}
	if out != nil {
		t.Fatalf("expected nil user, got %v", out)
	}
	if repo.calledUpdate {
		t.Fatalf("expected Update not to be called")
	}
}

func TestService_Update_InvalidBirthDate(t *testing.T) {
	u := baseUserForUpdate()
	repo := &fakeUserRepo{findByIDRes: u}
	svc := New(repo, nil, nil)

	future := time.Now().Add(24 * time.Hour)
	userID := uuid.Must(uuid.NewV7())
	out, err := svc.Update(context.Background(), UpdateInput{UserID: userID, BirthDate: &future})
	if !errors.Is(err, user.ErrInvalidBirthDate) {
		t.Fatalf("expected ErrInvalidBirthDate, got %v", err)
	}
	if out != nil {
		t.Fatalf("expected nil user, got %v", out)
	}
	if repo.calledUpdate {
		t.Fatalf("expected Update not to be called")
	}
}

func TestService_Update_InvalidCPF(t *testing.T) {
	u := baseUserForUpdate()
	repo := &fakeUserRepo{findByIDRes: u}
	svc := New(repo, nil, nil)

	badCPF := "123"
	userID := uuid.Must(uuid.NewV7())
	out, err := svc.Update(context.Background(), UpdateInput{UserID: userID, CPF: &badCPF})
	if !errors.Is(err, user.ErrInvalidCPF) {
		t.Fatalf("expected ErrInvalidCPF, got %v", err)
	}
	if out != nil {
		t.Fatalf("expected nil user, got %v", out)
	}
	if repo.calledUpdate {
		t.Fatalf("expected Update not to be called")
	}
}

func TestService_Update_InvalidPhone(t *testing.T) {
	u := baseUserForUpdate()
	repo := &fakeUserRepo{findByIDRes: u}
	svc := New(repo, nil, nil)

	bad := "\t\n"
	userID := uuid.Must(uuid.NewV7())
	out, err := svc.Update(context.Background(), UpdateInput{UserID: userID, Phone: &bad})
	if !errors.Is(err, user.ErrInvalidPhone) {
		t.Fatalf("expected ErrInvalidPhone, got %v", err)
	}
	if out != nil {
		t.Fatalf("expected nil user, got %v", out)
	}
	if repo.calledUpdate {
		t.Fatalf("expected Update not to be called")
	}
}

func TestService_Update_UpdateError(t *testing.T) {
	sentinel := errors.New("update failed")
	u := baseUserForUpdate()
	repo := &fakeUserRepo{findByIDRes: u, updateErr: sentinel}
	svc := New(repo, nil, nil)

	newName := "Nome Novo"
	userID := uuid.Must(uuid.NewV7())
	out, err := svc.Update(context.Background(), UpdateInput{UserID: userID, FullName: &newName})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	if out != nil {
		t.Fatalf("expected nil user, got %v", out)
	}
	if !repo.calledUpdate {
		t.Fatalf("expected Update to be called")
	}
	if len(repo.updated) != 1 {
		t.Fatalf("expected 1 Update call, got %d", len(repo.updated))
	}
	if repo.updated[0] != u {
		t.Fatalf("expected Update to receive the existing user pointer")
	}
}

func TestService_Update_Success(t *testing.T) {
	u := baseUserForUpdate()
	repo := &fakeUserRepo{findByIDRes: u}
	svc := New(repo, nil, nil)

	newName := "  Nome Novo  "
	newBirth := time.Date(1991, 2, 3, 0, 0, 0, 0, time.UTC)
	newCPF := "529.982.247-25"
	newPhone := "  11888887777  "
	userID := uuid.Must(uuid.NewV7())

	out, err := svc.Update(context.Background(), UpdateInput{
		UserID:    userID,
		FullName:  &newName,
		BirthDate: &newBirth,
		CPF:       &newCPF,
		Phone:     &newPhone,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if out == nil {
		t.Fatalf("expected user, got nil")
	}
	if out != u {
		t.Fatalf("expected same pointer returned")
	}
	if out.FullName != "Nome Novo" {
		t.Fatalf("expected trimmed full name, got %q", out.FullName)
	}
	if !out.BirthDate.Equal(newBirth) {
		t.Fatalf("expected updated birth date")
	}
	if out.CPF != "52998224725" {
		t.Fatalf("expected normalized cpf, got %q", out.CPF)
	}
	if out.Phone != "11888887777" {
		t.Fatalf("expected trimmed phone, got %q", out.Phone)
	}
	if !repo.calledUpdate {
		t.Fatalf("expected Update to be called")
	}
	if len(repo.updated) != 1 {
		t.Fatalf("expected 1 Update call, got %d", len(repo.updated))
	}
	if repo.updated[0] != u {
		t.Fatalf("expected Update to receive the existing user pointer")
	}
}

func TestService_Delete_UserNotFound(t *testing.T) {
	repo := &fakeUserRepo{findByIDRes: nil}
	auth := &fakeIdentitySvc{}
	svc := New(repo, nil, auth)

	err := svc.Delete(context.Background(), uuid.Nil)
	if !errors.Is(err, user.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
	if repo.calledSoftDelete {
		t.Fatalf("expected SoftDelete not to be called")
	}
	if auth.calledDisable {
		t.Fatalf("expected DisableUser not to be called")
	}
}

func TestService_Delete_FindByIDError(t *testing.T) {
	sentinel := errors.New("db down")
	repo := &fakeUserRepo{findByIDErr: sentinel}
	auth := &fakeIdentitySvc{}
	svc := New(repo, nil, auth)

	userID := uuid.Must(uuid.NewV7())
	err := svc.Delete(context.Background(), userID)
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	if repo.calledSoftDelete {
		t.Fatalf("expected SoftDelete not to be called")
	}
	if auth.calledDisable {
		t.Fatalf("expected DisableUser not to be called")
	}
}

func TestService_Delete_SoftDeleteError(t *testing.T) {
	sentinel := errors.New("soft delete failed")
	userID := uuid.Must(uuid.NewV7())
	repo := &fakeUserRepo{
		findByIDRes:   &user.User{ID: userID, AuthProvider: "custom"},
		softDeleteErr: sentinel,
	}
	auth := &fakeIdentitySvc{}
	svc := New(repo, nil, auth)

	err := svc.Delete(context.Background(), userID)
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	if !repo.calledSoftDelete {
		t.Fatalf("expected SoftDelete to be called")
	}
	if len(repo.softDeleted) != 1 || repo.softDeleted[0] != userID {
		t.Fatalf("expected SoftDelete to be called with correct user ID")
	}
	if auth.calledDisable {
		t.Fatalf("expected DisableUser not to be called")
	}
}

func TestService_Delete_DisableError(t *testing.T) {
	sentinel := errors.New("disable failed")
	userID := uuid.Must(uuid.NewV7())
	repo := &fakeUserRepo{
		findByIDRes: &user.User{ID: userID, AuthProvider: "firebase", AuthSubject: "subj"},
	}
	auth := &fakeIdentitySvc{disableErr: sentinel}
	svc := New(repo, nil, auth)

	err := svc.Delete(context.Background(), userID)
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	if !repo.calledSoftDelete {
		t.Fatalf("expected SoftDelete to be called")
	}
	if len(repo.softDeleted) != 1 || repo.softDeleted[0] != userID {
		t.Fatalf("expected SoftDelete to be called with correct user ID")
	}
	if !auth.calledDisable || len(auth.disabledSubj) != 1 || auth.disabledSubj[0] != "subj" {
		t.Fatalf("expected DisableUser to be called with subject")
	}
}

func TestService_Delete_Success_Firebase(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	repo := &fakeUserRepo{findByIDRes: &user.User{ID: userID, AuthProvider: "firebase", AuthSubject: "subj"}}
	auth := &fakeIdentitySvc{}
	svc := New(repo, nil, auth)

	err := svc.Delete(context.Background(), userID)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !repo.calledSoftDelete {
		t.Fatalf("expected SoftDelete to be called")
	}
	if len(repo.softDeleted) != 1 || repo.softDeleted[0] != userID {
		t.Fatalf("expected SoftDelete to be called with correct user ID")
	}
	if !auth.calledDisable || len(auth.disabledSubj) != 1 || auth.disabledSubj[0] != "subj" {
		t.Fatalf("expected DisableUser to be called for firebase user")
	}
}

func TestService_Delete_Success_NonFirebase(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	repo := &fakeUserRepo{findByIDRes: &user.User{ID: userID, AuthProvider: "custom"}}
	auth := &fakeIdentitySvc{}
	svc := New(repo, nil, auth)

	err := svc.Delete(context.Background(), userID)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !repo.calledSoftDelete {
		t.Fatalf("expected SoftDelete to be called")
	}
	if len(repo.softDeleted) != 1 || repo.softDeleted[0] != userID {
		t.Fatalf("expected SoftDelete to be called with correct user ID")
	}
	if auth.calledDisable {
		t.Fatalf("expected DisableUser not to be called for non-firebase provider")
	}
}
