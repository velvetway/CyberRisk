package repository

import (
	"context"
	"fmt"
	"testing"

	"Diplom/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssetRepository_Create(t *testing.T) {
	pool := setupTestDB(t)
	truncateAll(t, pool)
	repo := NewAssetRepository(pool)
	ctx := context.Background()

	asset := &domain.Asset{
		Name:        "Test Server",
		Environment: domain.AssetEnvProd,
		IsIsolated:  false,
	}
	err := repo.Create(ctx, asset)
	require.NoError(t, err)
	assert.Greater(t, asset.ID, int64(0))
}

func TestAssetRepository_GetByID(t *testing.T) {
	pool := setupTestDB(t)
	truncateAll(t, pool)
	repo := NewAssetRepository(pool)
	ctx := context.Background()

	asset := &domain.Asset{
		Name:        "Test DB",
		Environment: domain.AssetEnvProd,
	}
	require.NoError(t, repo.Create(ctx, asset))

	found, err := repo.GetByID(ctx, asset.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "Test DB", found.Name)
}

func TestAssetRepository_List(t *testing.T) {
	pool := setupTestDB(t)
	truncateAll(t, pool)
	repo := NewAssetRepository(pool)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		require.NoError(t, repo.Create(ctx, &domain.Asset{
			Name:        fmt.Sprintf("Asset %d", i),
			Environment: domain.AssetEnvDev,
		}))
	}

	assets, err := repo.List(ctx, AssetFilter{Limit: 10, Offset: 0})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(assets), 3)
}

func TestAssetRepository_Update(t *testing.T) {
	pool := setupTestDB(t)
	truncateAll(t, pool)
	repo := NewAssetRepository(pool)
	ctx := context.Background()

	asset := &domain.Asset{
		Name:        "Original",
		Environment: domain.AssetEnvDev,
	}
	require.NoError(t, repo.Create(ctx, asset))
	asset.Name = "Updated"
	asset.IsIsolated = true
	require.NoError(t, repo.Update(ctx, asset))

	found, _ := repo.GetByID(ctx, asset.ID)
	assert.Equal(t, "Updated", found.Name)
	assert.True(t, found.IsIsolated)
}

func TestAssetRepository_Delete(t *testing.T) {
	pool := setupTestDB(t)
	truncateAll(t, pool)
	repo := NewAssetRepository(pool)
	ctx := context.Background()

	asset := &domain.Asset{
		Name:        "ToDelete",
		Environment: domain.AssetEnvTest,
	}
	require.NoError(t, repo.Create(ctx, asset))
	require.NoError(t, repo.Delete(ctx, asset.ID))

	found, _ := repo.GetByID(ctx, asset.ID)
	assert.Nil(t, found)
}
