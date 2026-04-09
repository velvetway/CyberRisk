package risk

import (
	"testing"

	"Diplom/internal/domain"

	"github.com/stretchr/testify/assert"
)

func ptr[T any](v T) *T { return &v }

func baseAsset() *domain.Asset {
	return &domain.Asset{
		ID: 1, Name: "Test Server",
		BusinessCriticality: 3, Confidentiality: 3, Integrity: 3, Availability: 3,
		Environment: domain.AssetEnvProd, HasInternetAccess: true,
	}
}

func baseThreat() *domain.Threat {
	return &domain.Threat{
		ID: 1, Name: "Test Threat", BaseLikelihood: 3,
		ImpactConfidentiality: false, ImpactIntegrity: false, ImpactAvailability: false,
	}
}

func TestCalculate_BasicRisk(t *testing.T) {
	calc := NewCalculator()
	result := calc.Calculate(baseAsset(), baseThreat(), []domain.Vulnerability{{Severity: 3}})
	assert.GreaterOrEqual(t, result.Impact, int16(1))
	assert.LessOrEqual(t, result.Impact, int16(5))
	assert.GreaterOrEqual(t, result.Likelihood, int16(1))
	assert.LessOrEqual(t, result.Likelihood, int16(5))
	assert.Equal(t, result.Impact*result.Likelihood, result.Score)
}

func TestCalculate_HighCriticalityIncreasesImpact(t *testing.T) {
	calc := NewCalculator()
	threat := baseThreat()
	vulns := []domain.Vulnerability{{Severity: 3}}
	lowCrit := baseAsset(); lowCrit.BusinessCriticality = 1
	highCrit := baseAsset(); highCrit.BusinessCriticality = 5
	assert.Greater(t, calc.Calculate(highCrit, threat, vulns).Impact, calc.Calculate(lowCrit, threat, vulns).Impact)
}

func TestCalculate_HighSeverityIncreasesImpact(t *testing.T) {
	calc := NewCalculator()
	asset := baseAsset()
	threat := baseThreat()
	resultLow := calc.Calculate(asset, threat, []domain.Vulnerability{{Severity: 1}})
	resultHigh := calc.Calculate(asset, threat, []domain.Vulnerability{{Severity: 5}})
	assert.GreaterOrEqual(t, resultHigh.Impact, resultLow.Impact)
}

func TestCalculate_ProdEnvIncreasesLikelihood(t *testing.T) {
	calc := NewCalculator()
	threat := baseThreat()
	vulns := []domain.Vulnerability{{Severity: 3}}
	devAsset := baseAsset(); devAsset.Environment = domain.AssetEnvDev
	prodAsset := baseAsset(); prodAsset.Environment = domain.AssetEnvProd
	assert.GreaterOrEqual(t, calc.Calculate(prodAsset, threat, vulns).Likelihood, calc.Calculate(devAsset, threat, vulns).Likelihood)
}

func TestCalculate_IsolatedReducesLikelihood(t *testing.T) {
	calc := NewCalculator()
	threat := baseThreat(); threat.BaseLikelihood = 4
	vulns := []domain.Vulnerability{{Severity: 3}}
	open := baseAsset(); open.IsIsolated = false
	isolated := baseAsset(); isolated.IsIsolated = true
	assert.LessOrEqual(t, calc.Calculate(isolated, threat, vulns).Likelihood, calc.Calculate(open, threat, vulns).Likelihood)
}

func TestCalculate_ThreatCIABonus(t *testing.T) {
	calc := NewCalculator()
	vulns := []domain.Vulnerability{{Severity: 3}}
	asset := baseAsset(); asset.Confidentiality = 5; asset.Integrity = 5; asset.Availability = 5
	noCIA := baseThreat()
	fullCIA := baseThreat(); fullCIA.ImpactConfidentiality = true; fullCIA.ImpactIntegrity = true; fullCIA.ImpactAvailability = true
	assert.Greater(t, calc.Calculate(asset, fullCIA, vulns).Impact, calc.Calculate(asset, noCIA, vulns).Impact)
}

func TestCalculate_RiskLevels(t *testing.T) {
	tests := []struct{ score int16; level string }{
		{1, "Low"}, {5, "Low"}, {6, "Medium"}, {10, "Medium"},
		{11, "High"}, {15, "High"}, {16, "Critical"}, {25, "Critical"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.level, riskLevel(tt.score), "score=%d", tt.score)
	}
}

func TestCalculate_RegulatoryFactor_KII(t *testing.T) {
	calc := NewCalculator()
	asset := baseAsset(); asset.KIICategory = ptr(domain.KIICategoryCat1)
	result := calc.Calculate(asset, baseThreat(), []domain.Vulnerability{{Severity: 3}})
	assert.GreaterOrEqual(t, result.RegulatoryFactor, 2.0)
}

func TestCalculate_RegulatoryFactor_StateSecret(t *testing.T) {
	calc := NewCalculator()
	asset := baseAsset(); asset.DataCategory = ptr(domain.DataCategoryStateSecret)
	result := calc.Calculate(asset, baseThreat(), []domain.Vulnerability{{Severity: 3}})
	assert.GreaterOrEqual(t, result.RegulatoryFactor, 2.5)
}

func TestCalculate_RegulatoryFactor_PersonalDataVolume(t *testing.T) {
	calc := NewCalculator()
	asset := baseAsset(); asset.HasPersonalData = true; asset.PersonalDataVolume = ptr(">100000"); asset.DataCategory = ptr(domain.DataCategoryPersonalData)
	result := calc.Calculate(asset, baseThreat(), []domain.Vulnerability{{Severity: 3}})
	assert.Greater(t, result.RegulatoryFactor, 1.5)
}

func TestCalculate_RegulatoryFactor_NoRegulatory(t *testing.T) {
	calc := NewCalculator()
	result := calc.Calculate(baseAsset(), baseThreat(), []domain.Vulnerability{{Severity: 3}})
	assert.Equal(t, 1.0, result.RegulatoryFactor)
}

func TestCalculate_NoVulns(t *testing.T) {
	calc := NewCalculator()
	result := calc.Calculate(baseAsset(), baseThreat(), nil)
	assert.GreaterOrEqual(t, result.Impact, int16(1))
}

func TestCalculate_ClampValues(t *testing.T) {
	calc := NewCalculator()
	asset := baseAsset(); asset.BusinessCriticality = 5; asset.Confidentiality = 5; asset.Integrity = 5; asset.Availability = 5
	threat := baseThreat(); threat.BaseLikelihood = 5; threat.ImpactConfidentiality = true; threat.ImpactIntegrity = true; threat.ImpactAvailability = true
	result := calc.Calculate(asset, threat, []domain.Vulnerability{{Severity: 10}})
	assert.LessOrEqual(t, result.Impact, int16(5))
	assert.LessOrEqual(t, result.Likelihood, int16(5))
}

func TestAdjustedRiskLevel(t *testing.T) {
	tests := []struct{ score float64; level string }{
		{5.0, "Low"}, {12.0, "Medium"}, {22.0, "High"}, {32.0, "Critical"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.level, AdjustedRiskLevel(tt.score))
	}
}

func TestMaxVulnSeverity(t *testing.T) {
	assert.Equal(t, int16(1), maxVulnSeverity(nil))
	assert.Equal(t, int16(7), maxVulnSeverity([]domain.Vulnerability{{Severity: 3}, {Severity: 7}, {Severity: 2}}))
}
