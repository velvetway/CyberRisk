package auth

import (
	"context"
	"testing"

	"Diplom/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeUserRepo struct {
	users  map[string]*domain.User
	nextID int64
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{users: make(map[string]*domain.User), nextID: 1}
}

func (r *fakeUserRepo) Create(_ context.Context, user *domain.User) error {
	user.ID = r.nextID
	r.nextID++
	r.users[user.Username] = user
	return nil
}

func (r *fakeUserRepo) GetByUsername(_ context.Context, username string) (*domain.User, error) {
	u, ok := r.users[username]
	if !ok {
		return nil, nil
	}
	return u, nil
}

func (r *fakeUserRepo) GetByID(_ context.Context, id int64) (*domain.User, error) {
	for _, u := range r.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, nil
}

func TestRegister_Success(t *testing.T) {
	svc := NewService(newFakeUserRepo(), "test-secret")
	user, err := svc.Register(context.Background(), "admin", "password123", domain.UserRoleAdmin)
	require.NoError(t, err)
	assert.Equal(t, "admin", user.Username)
	assert.Equal(t, domain.UserRoleAdmin, user.Role)
	assert.True(t, user.IsActive)
	assert.NotEqual(t, "password123", user.PasswordHash)
}

func TestRegister_DuplicateUser(t *testing.T) {
	svc := NewService(newFakeUserRepo(), "test-secret")
	_, err := svc.Register(context.Background(), "admin", "password123", domain.UserRoleAdmin)
	require.NoError(t, err)
	_, err = svc.Register(context.Background(), "admin", "other", domain.UserRoleViewer)
	assert.ErrorIs(t, err, ErrUserExists)
}

func TestLogin_Success(t *testing.T) {
	svc := NewService(newFakeUserRepo(), "test-secret")
	_, _ = svc.Register(context.Background(), "user1", "pass123", domain.UserRoleViewer)
	token, err := svc.Login(context.Background(), "user1", "pass123")
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestLogin_WrongPassword(t *testing.T) {
	svc := NewService(newFakeUserRepo(), "test-secret")
	_, _ = svc.Register(context.Background(), "user1", "pass123", domain.UserRoleViewer)
	_, err := svc.Login(context.Background(), "user1", "wrong")
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestLogin_NonExistentUser(t *testing.T) {
	svc := NewService(newFakeUserRepo(), "test-secret")
	_, err := svc.Login(context.Background(), "nobody", "pass")
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestValidateToken_Success(t *testing.T) {
	svc := NewService(newFakeUserRepo(), "test-secret")
	_, _ = svc.Register(context.Background(), "user1", "pass123", domain.UserRoleAuditor)
	token, _ := svc.Login(context.Background(), "user1", "pass123")
	claims, err := svc.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, "user1", claims.Username)
	assert.Equal(t, "auditor", claims.Role)
}

func TestValidateToken_InvalidToken(t *testing.T) {
	svc := NewService(newFakeUserRepo(), "test-secret")
	_, err := svc.ValidateToken("invalid.token.here")
	assert.Error(t, err)
}

func TestValidateToken_WrongSecret(t *testing.T) {
	repo := newFakeUserRepo()
	svc1 := NewService(repo, "secret-1")
	_, _ = svc1.Register(context.Background(), "user1", "pass", domain.UserRoleViewer)
	token, _ := svc1.Login(context.Background(), "user1", "pass")
	svc2 := NewService(newFakeUserRepo(), "secret-2")
	_, err := svc2.ValidateToken(token)
	assert.Error(t, err)
}
