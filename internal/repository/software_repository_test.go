package repository

import (
	"context"
	"testing"

	"Diplom/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSoftwareRepository_CreateAndGet(t *testing.T) {
	pool := setupTestDB(t)
	truncateAll(t, pool)
	repo := NewSoftwareRepository(pool)
	ctx := context.Background()

	sw := &domain.Software{Name: "Kaspersky Endpoint Security", Vendor: "Kaspersky Lab", IsRussian: true}
	require.NoError(t, repo.Create(ctx, sw))
	assert.Greater(t, sw.ID, int64(0))

	found, err := repo.GetByID(ctx, sw.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "Kaspersky Endpoint Security", found.Name)
	assert.True(t, found.IsRussian)
}

func TestSoftwareRepository_ListWithFilter(t *testing.T) {
	pool := setupTestDB(t)
	truncateAll(t, pool)
	repo := NewSoftwareRepository(pool)
	ctx := context.Background()

	_ = repo.Create(ctx, &domain.Software{Name: "KES", Vendor: "Kaspersky", IsRussian: true})
	_ = repo.Create(ctx, &domain.Software{Name: "Norton", Vendor: "Symantec", IsRussian: false})

	russian := true
	list, err := repo.List(ctx, SoftwareFilter{Limit: 10, IsRussian: &russian})
	require.NoError(t, err)
	for _, sw := range list {
		assert.True(t, sw.IsRussian)
	}
}

func TestSoftwareRepository_Delete(t *testing.T) {
	pool := setupTestDB(t)
	truncateAll(t, pool)
	repo := NewSoftwareRepository(pool)
	ctx := context.Background()

	sw := &domain.Software{Name: "ToDelete", Vendor: "Test"}
	_ = repo.Create(ctx, sw)
	require.NoError(t, repo.Delete(ctx, sw.ID))

	found, _ := repo.GetByID(ctx, sw.ID)
	assert.Nil(t, found)
}
