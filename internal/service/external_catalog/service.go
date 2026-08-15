package external_catalog

import (
	"context"
	"errors"
	"math"

	"Diplom/internal/domain"
	"Diplom/internal/repository"
)

type Service interface {
	SearchBDU(ctx context.Context, f repository.BDUSearchFilter) ([]domain.BDUVulnerability, error)
	SearchMinreestr(ctx context.Context, f repository.MinreestrFilter) ([]domain.Software, error)
	SyncAssetBDUVulnerabilities(ctx context.Context, assetID int64, limitPerSoftware int) (*AssetBDUSyncResult, error)
}

type AssetBDUSyncResult struct {
	AssetID            int64                    `json:"asset_id"`
	SoftwareChecked    int                      `json:"software_checked"`
	Vulnerabilities    []SyncedBDUVulnerability `json:"vulnerabilities"`
	BDUAvailable       bool                     `json:"bdu_available"`
	MinreestrAvailable bool                     `json:"minreestr_available"`
}

type SyncedBDUVulnerability struct {
	BDU           domain.BDUVulnerability `json:"bdu"`
	Vulnerability domain.Vulnerability    `json:"vulnerability"`
	Software      domain.Software         `json:"software"`
}

type service struct {
	bdu           repository.BDURepository
	minreestr     repository.MinreestrRepository
	softwareRepo  repository.SoftwareRepository
	vulnRepo      repository.VulnerabilityRepository
	assetVulnRepo repository.AssetVulnerabilityRepository
}

func NewService(
	bdu repository.BDURepository,
	minreestr repository.MinreestrRepository,
	softwareRepo repository.SoftwareRepository,
	vulnRepo repository.VulnerabilityRepository,
	assetVulnRepo repository.AssetVulnerabilityRepository,
) Service {
	return &service{
		bdu:           bdu,
		minreestr:     minreestr,
		softwareRepo:  softwareRepo,
		vulnRepo:      vulnRepo,
		assetVulnRepo: assetVulnRepo,
	}
}

func (s *service) SearchBDU(ctx context.Context, f repository.BDUSearchFilter) ([]domain.BDUVulnerability, error) {
	if s.bdu == nil || !s.bdu.IsAvailable() {
		return nil, errors.New("bdu sqlite catalog is not available")
	}
	return s.bdu.Search(ctx, f)
}

func (s *service) SearchMinreestr(ctx context.Context, f repository.MinreestrFilter) ([]domain.Software, error) {
	if s.minreestr == nil || !s.minreestr.IsAvailable() {
		return nil, errors.New("minreestr sqlite catalog is not available")
	}
	return s.minreestr.Search(ctx, f)
}

func (s *service) SyncAssetBDUVulnerabilities(ctx context.Context, assetID int64, limitPerSoftware int) (*AssetBDUSyncResult, error) {
	if limitPerSoftware <= 0 || limitPerSoftware > 50 {
		limitPerSoftware = 10
	}

	result := &AssetBDUSyncResult{
		AssetID:            assetID,
		BDUAvailable:       s.bdu != nil && s.bdu.IsAvailable(),
		MinreestrAvailable: s.minreestr != nil && s.minreestr.IsAvailable(),
	}
	if s.bdu == nil || !s.bdu.IsAvailable() {
		return result, nil
	}

	assetSoftware, err := s.softwareRepo.ListAssetSoftware(ctx, assetID)
	if err != nil {
		return nil, err
	}
	result.SoftwareChecked = len(assetSoftware)

	seen := map[string]bool{}
	for _, item := range assetSoftware {
		sw := item.Software
		matches, err := s.bdu.Search(ctx, repository.BDUSearchFilter{
			Software: sw.Name,
			Vendor:   sw.Vendor,
			Limit:    limitPerSoftware,
		})
		if err != nil {
			return nil, err
		}
		for _, bduVuln := range matches {
			if seen[bduVuln.ID] {
				continue
			}
			seen[bduVuln.ID] = true

			v := vulnerabilityFromBDU(bduVuln)
			if err := s.vulnRepo.UpsertExternal(ctx, &v); err != nil {
				return nil, err
			}
			if err := s.assetVulnRepo.Add(ctx, &domain.AssetVulnerability{
				AssetID:         assetID,
				VulnerabilityID: v.ID,
			}); err != nil {
				return nil, err
			}

			result.Vulnerabilities = append(result.Vulnerabilities, SyncedBDUVulnerability{
				BDU:           bduVuln,
				Vulnerability: v,
				Software:      sw,
			})
		}
	}

	return result, nil
}

func vulnerabilityFromBDU(v domain.BDUVulnerability) domain.Vulnerability {
	externalID := v.ID
	severity := v.SeverityLevel
	if severity < 1 && v.CVSSScore != nil {
		severity = int16(math.Ceil(*v.CVSSScore / 2.0))
	}
	if severity < 1 {
		severity = 1
	}
	if severity > 10 {
		severity = 10
	}

	desc := v.Description
	if desc == nil {
		desc = v.Solution
	}

	return domain.Vulnerability{
		Name:          v.Name,
		Description:   desc,
		Severity:      severity,
		ExternalID:    &externalID,
		Source:        "bdu",
		CVSSScore:     v.CVSSScore,
		CVEs:          v.CVEs,
		Vendors:       v.Vendors,
		SoftwareNames: v.SoftwareNames,
	}
}
