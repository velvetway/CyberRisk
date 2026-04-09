package repository

import (
	"context"
	"testing"

	"Diplom/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_CreateAndGetByUsername(t *testing.T) {
	pool := setupTestDB(t)
	truncateAll(t, pool)
	repo := NewUserRepository(pool)
	ctx := context.Background()

	user := &domain.User{
		Username: "testadmin", PasswordHash: "$2a$10$fakehashhere",
		Role: domain.UserRoleAdmin, IsActive: true,
	}
	require.NoError(t, repo.Create(ctx, user))
	assert.Greater(t, user.ID, int64(0))

	found, err := repo.GetByUsername(ctx, "testadmin")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, domain.UserRoleAdmin, found.Role)
}

func TestUserRepository_GetByID(t *testing.T) {
	pool := setupTestDB(t)
	truncateAll(t, pool)
	repo := NewUserRepository(pool)
	ctx := context.Background()

	user := &domain.User{
		Username: "viewer1", PasswordHash: "$2a$10$fakehash",
		Role: domain.UserRoleViewer, IsActive: true,
	}
	_ = repo.Create(ctx, user)

	found, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "viewer1", found.Username)
}

func TestUserRepository_GetByUsername_NotFound(t *testing.T) {
	pool := setupTestDB(t)
	truncateAll(t, pool)
	repo := NewUserRepository(pool)
	ctx := context.Background()

	found, err := repo.GetByUsername(ctx, "nonexistent")
	require.NoError(t, err)
	assert.Nil(t, found)
}
