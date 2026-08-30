package control

import (
	"context"
	"errors"
	"testing"

	"Diplom/internal/domain"
)

// Mock-репозиторий — захватывает аргументы и возвращает заранее заданное.
type mockControlRepo struct {
	listResult        []domain.Control
	getByIDResult     *domain.Control
	listAttachedFn    func(ctx context.Context, assetID int64) ([]domain.Control, error)
	attachCalls       []struct{ asset, ctrl int64 }
	detachCalls       []struct{ asset, ctrl int64 }
	attachErr         error
	detachErr         error
}

func (m *mockControlRepo) List(ctx context.Context) ([]domain.Control, error) {
	return m.listResult, nil
}
func (m *mockControlRepo) GetByID(ctx context.Context, id int64) (*domain.Control, error) {
	return m.getByIDResult, nil
}
func (m *mockControlRepo) ListAttached(ctx context.Context, assetID int64) ([]domain.Control, error) {
	if m.listAttachedFn != nil {
		return m.listAttachedFn(ctx, assetID)
	}
	return nil, nil
}
func (m *mockControlRepo) Attach(ctx context.Context, assetID, controlID int64) error {
	m.attachCalls = append(m.attachCalls, struct{ asset, ctrl int64 }{assetID, controlID})
	return m.attachErr
}
func (m *mockControlRepo) Detach(ctx context.Context, assetID, controlID int64) error {
	m.detachCalls = append(m.detachCalls, struct{ asset, ctrl int64 }{assetID, controlID})
	return m.detachErr
}

// Положительные пути на validate-инвариантах: assetID/controlID > 0.

func TestService_Attach_Valid(t *testing.T) {
	repo := &mockControlRepo{}
	svc := NewService(repo)

	if err := svc.Attach(context.Background(), 1, 7); err != nil {
		t.Fatalf("ожидалось nil, получили %v", err)
	}
	if len(repo.attachCalls) != 1 {
		t.Fatalf("Attach не вызван в репозитории")
	}
	if repo.attachCalls[0].asset != 1 || repo.attachCalls[0].ctrl != 7 {
		t.Errorf("repo получил неправильные id-шники: %+v", repo.attachCalls[0])
	}
}

func TestService_Detach_Valid(t *testing.T) {
	repo := &mockControlRepo{}
	svc := NewService(repo)

	if err := svc.Detach(context.Background(), 5, 11); err != nil {
		t.Fatalf("ожидалось nil, получили %v", err)
	}
	if len(repo.detachCalls) != 1 {
		t.Fatalf("Detach не вызван в репозитории")
	}
}

// Невалидные id — сервис должен не дать в репо упасть.
// Мы проверяем что repo.attachCalls/detachCalls остались пустыми.
func TestService_Attach_InvalidIDs(t *testing.T) {
	tests := []struct{ asset, ctrl int64 }{
		{0, 1},   // asset=0
		{-1, 1},  // asset<0
		{1, 0},   // control=0
		{1, -1},  // control<0
		{0, 0},   // оба нуля
	}
	for _, tt := range tests {
		repo := &mockControlRepo{}
		svc := NewService(repo)
		if err := svc.Attach(context.Background(), tt.asset, tt.ctrl); err == nil {
			t.Errorf("Attach(%d,%d) ожидали ошибку, получили nil", tt.asset, tt.ctrl)
		}
		if len(repo.attachCalls) > 0 {
			t.Errorf("Attach(%d,%d) дошёл до репо несмотря на валидацию", tt.asset, tt.ctrl)
		}
	}
}

func TestService_Detach_InvalidIDs(t *testing.T) {
	repo := &mockControlRepo{}
	svc := NewService(repo)
	if err := svc.Detach(context.Background(), 0, 1); err == nil {
		t.Errorf("Detach(0,1) ожидали ошибку")
	}
	if err := svc.Detach(context.Background(), 1, 0); err == nil {
		t.Errorf("Detach(1,0) ожидали ошибку")
	}
	if len(repo.detachCalls) > 0 {
		t.Errorf("Detach с невалидными id дошёл до репо")
	}
}

func TestService_ListForAsset_InvalidID(t *testing.T) {
	repo := &mockControlRepo{}
	svc := NewService(repo)
	if _, err := svc.ListForAsset(context.Background(), 0); err == nil {
		t.Errorf("ListForAsset(0) ожидали ошибку")
	}
	if _, err := svc.ListForAsset(context.Background(), -1); err == nil {
		t.Errorf("ListForAsset(-1) ожидали ошибку")
	}
}

// Если репозиторий возвращает ошибку — она должна пробрасываться.
func TestService_Attach_RepoError(t *testing.T) {
	repo := &mockControlRepo{attachErr: errors.New("db down")}
	svc := NewService(repo)
	if err := svc.Attach(context.Background(), 1, 1); err == nil {
		t.Errorf("ожидали ошибку из репо, получили nil")
	}
}

// List и ListForAsset просто пробрасывают результат.
func TestService_List_PassthroughFromRepo(t *testing.T) {
	repo := &mockControlRepo{listResult: []domain.Control{
		{ID: 1, Name: "Антивирус"},
		{ID: 2, Name: "МСЭ"},
	}}
	svc := NewService(repo)
	out, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(out) != 2 || out[0].Name != "Антивирус" {
		t.Errorf("ожидали 2 контроля с именами от репо, получили %v", out)
	}
}

func TestService_ListForAsset_PassthroughFromRepo(t *testing.T) {
	repo := &mockControlRepo{
		listAttachedFn: func(ctx context.Context, assetID int64) ([]domain.Control, error) {
			if assetID != 42 {
				t.Errorf("неверный assetID: %d", assetID)
			}
			return []domain.Control{{ID: 1, Name: "X"}}, nil
		},
	}
	svc := NewService(repo)
	out, err := svc.ListForAsset(context.Background(), 42)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(out) != 1 || out[0].Name != "X" {
		t.Errorf("результат не пробросился: %v", out)
	}
}
