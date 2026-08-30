package external_catalog

import (
	"context"
	"errors"

	"Diplom/internal/domain"
	"Diplom/internal/repository"
)

type Service interface {
	SearchBDU(ctx context.Context, f repository.BDUSearchFilter) ([]domain.BDUVulnerability, error)
	SearchMinreestr(ctx context.Context, f repository.MinreestrFilter) ([]domain.Software, error)
	SearchSZI(ctx context.Context, f repository.SZISearchFilter) ([]domain.SZICertificate, error)
	SZIControlCoverage(ctx context.Context) ([]domain.SZIControlCoverage, error)
}

type service struct {
	bdu       repository.BDURepository
	minreestr repository.MinreestrRepository
	szi       repository.SZIRepository
}

// NewService принимает только каталоги: запись уязвимостей в актив
// делает автодетект в asset_vulnerability, а этот сервис лишь читает
// внешние справочники.
func NewService(
	bdu repository.BDURepository,
	minreestr repository.MinreestrRepository,
	szi repository.SZIRepository,
) Service {
	return &service{bdu: bdu, minreestr: minreestr, szi: szi}
}

// SearchSZI отдаёт сертифицированные средства защиты из реестра ФСТЭК.
func (s *service) SearchSZI(ctx context.Context, f repository.SZISearchFilter) ([]domain.SZICertificate, error) {
	if s.szi == nil || !s.szi.IsAvailable() {
		return nil, errors.New("szi sqlite catalog is not available")
	}
	return s.szi.Search(ctx, f)
}

// SZIControlCoverage показывает, для каких методов ПТСЗИ вообще есть
// сертифицированные средства, а где выбирать не из чего.
func (s *service) SZIControlCoverage(ctx context.Context) ([]domain.SZIControlCoverage, error) {
	if s.szi == nil || !s.szi.IsAvailable() {
		return nil, errors.New("szi sqlite catalog is not available")
	}
	return s.szi.ControlCoverage(ctx)
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
