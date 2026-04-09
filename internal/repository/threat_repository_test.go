package repository

import (
	"context"
	"testing"

	"Diplom/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestThreatRepository_CreateAndGet(t *testing.T) {
	pool := setupTestDB(t)
	truncateAll(t, pool)
	repo := NewThreatRepository(pool)
	ctx := context.Background()

	threat := &domain.Threat{Name: "SQL Injection", SourceType: domain.ThreatSourceExternal, BaseLikelihood: 4}
	require.NoError(t, repo.Create(ctx, threat))
	assert.Greater(t, threat.ID, int64(0))

	found, err := repo.GetByID(ctx, threat.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "SQL Injection", found.Name)
}

func TestThreatRepository_List(t *testing.T) {
	pool := setupTestDB(t)
	truncateAll(t, pool)
	repo := NewThreatRepository(pool)
	ctx := context.Background()

	_ = repo.Create(ctx, &domain.Threat{Name: "T1", SourceType: domain.ThreatSourceExternal, BaseLikelihood: 3})
	_ = repo.Create(ctx, &domain.Threat{Name: "T2", SourceType: domain.ThreatSourceInternal, BaseLikelihood: 2})

	threats, err := repo.List(ctx, ThreatFilter{Limit: 10, Offset: 0})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(threats), 2)
}

func TestThreatRepository_Delete(t *testing.T) {
	pool := setupTestDB(t)
	truncateAll(t, pool)
	repo := NewThreatRepository(pool)
	ctx := context.Background()

	threat := &domain.Threat{Name: "ToDelete", SourceType: domain.ThreatSourceExternal, BaseLikelihood: 1}
	_ = repo.Create(ctx, threat)
	require.NoError(t, repo.Delete(ctx, threat.ID))

	found, _ := repo.GetByID(ctx, threat.ID)
	assert.Nil(t, found)
}
