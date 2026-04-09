package asset

import (
	"testing"

	"Diplom/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestCalculateCIA_Database(t *testing.T) {
	c, i, a := CalculateCIA("database", "prod", 3)
	assert.Equal(t, int16(5), c)
	assert.Equal(t, int16(5), i)
	assert.Equal(t, int16(4), a)
}

func TestCalculateCIA_DevReducesValues(t *testing.T) {
	cProd, iProd, aProd := CalculateCIA("server", "prod", 3)
	cDev, iDev, aDev := CalculateCIA("server", "dev", 3)
	assert.Greater(t, cProd, cDev)
	assert.Greater(t, iProd, iDev)
	assert.Greater(t, aProd, aDev)
}

func TestCalculateCIA_HighCriticalityIncreasesValues(t *testing.T) {
	cLow, iLow, aLow := CalculateCIA("application", "prod", 1)
	cHigh, iHigh, aHigh := CalculateCIA("application", "prod", 5)
	assert.Greater(t, cHigh, cLow)
	assert.Greater(t, iHigh, iLow)
	assert.Greater(t, aHigh, aLow)
}

func TestCalculateCIA_ClampedTo1_5(t *testing.T) {
	c, i, a := CalculateCIA("workstation", "dev", 1)
	assert.GreaterOrEqual(t, c, int16(1))
	assert.GreaterOrEqual(t, i, int16(1))
	assert.GreaterOrEqual(t, a, int16(1))

	c, i, a = CalculateCIA("cloud", "prod", 5)
	assert.LessOrEqual(t, c, int16(5))
	assert.LessOrEqual(t, i, int16(5))
	assert.LessOrEqual(t, a, int16(5))
}

func TestCalculateCIA_UnknownType(t *testing.T) {
	c, i, a := CalculateCIA("unknown_type", "prod", 3)
	assert.Equal(t, int16(4), c)
	assert.Equal(t, int16(4), i)
	assert.Equal(t, int16(4), a)
}

func TestCalculateCriticality_StateSecret(t *testing.T) {
	cat := "state_secret"
	result := CalculateCriticality(&cat, nil, nil, false, nil, false, false, "prod")
	assert.Equal(t, int16(5), result)
}

func TestCalculateCriticality_DevLowersScore(t *testing.T) {
	cat := "internal"
	prod := CalculateCriticality(&cat, nil, nil, false, nil, false, false, "prod")
	dev := CalculateCriticality(&cat, nil, nil, false, nil, false, false, "dev")
	assert.Greater(t, prod, dev)
}

func TestApplyCIA(t *testing.T) {
	assetType := "database"
	a := &domain.Asset{
		Type: &assetType,
		Environment: domain.AssetEnvProd,
		BusinessCriticality: 3,
	}
	ApplyCIA(a)
	assert.GreaterOrEqual(t, a.Confidentiality, int16(1))
	assert.LessOrEqual(t, a.Confidentiality, int16(5))
}
