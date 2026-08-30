// Package repository — control & asset_control DAO.
//
// Контроли в нашей модели — справочник из 11 канонических методов
// противодействия из диплома (Антивирус, МСЭ, Honeypot, …). Сидится
// миграцией 035; пользователи могут добавить свои, но обычно работают
// со существующими.
//
// asset_controls — фактические внедрения на конкретном активе. Это
// именно то, что использует formula W: для каждой VL-категории угрозы
// смотрим, есть ли control из vl_category_controls, который у этого
// актива «attached».
package repository

import (
	"context"
	"fmt"

	"Diplom/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ControlRepository interface {
	List(ctx context.Context) ([]domain.Control, error)
	GetByID(ctx context.Context, id int64) (*domain.Control, error)

	ListAttached(ctx context.Context, assetID int64) ([]domain.Control, error)
	Attach(ctx context.Context, assetID, controlID int64) error
	Detach(ctx context.Context, assetID, controlID int64) error
}

type controlRepository struct {
	pool *pgxpool.Pool
}

func NewControlRepository(pool *pgxpool.Pool) ControlRepository {
	return &controlRepository{pool: pool}
}

func (r *controlRepository) List(ctx context.Context) ([]domain.Control, error) {
	const q = `
SELECT id, name, control_type_id, description, created_at, updated_at
FROM controls
ORDER BY name`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list controls: %w", err)
	}
	defer rows.Close()

	out := make([]domain.Control, 0)
	for rows.Next() {
		var c domain.Control
		if err := rows.Scan(&c.ID, &c.Name, &c.ControlTypeID, &c.Description, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *controlRepository) GetByID(ctx context.Context, id int64) (*domain.Control, error) {
	const q = `
SELECT id, name, control_type_id, description, created_at, updated_at
FROM controls WHERE id = $1`
	var c domain.Control
	if err := r.pool.QueryRow(ctx, q, id).Scan(
		&c.ID, &c.Name, &c.ControlTypeID, &c.Description, &c.CreatedAt, &c.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("get control: %w", err)
	}
	return &c, nil
}

func (r *controlRepository) ListAttached(ctx context.Context, assetID int64) ([]domain.Control, error) {
	const q = `
SELECT c.id, c.name, c.control_type_id, c.description, c.created_at, c.updated_at
FROM asset_controls ac
JOIN controls c ON c.id = ac.control_id
WHERE ac.asset_id = $1
ORDER BY c.name`
	rows, err := r.pool.Query(ctx, q, assetID)
	if err != nil {
		return nil, fmt.Errorf("list attached controls: %w", err)
	}
	defer rows.Close()

	out := make([]domain.Control, 0)
	for rows.Next() {
		var c domain.Control
		if err := rows.Scan(&c.ID, &c.Name, &c.ControlTypeID, &c.Description, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *controlRepository) Attach(ctx context.Context, assetID, controlID int64) error {
	const q = `
INSERT INTO asset_controls (asset_id, control_id)
VALUES ($1, $2)
ON CONFLICT (asset_id, control_id) DO NOTHING`
	if _, err := r.pool.Exec(ctx, q, assetID, controlID); err != nil {
		return fmt.Errorf("attach control: %w", err)
	}
	return nil
}

func (r *controlRepository) Detach(ctx context.Context, assetID, controlID int64) error {
	if _, err := r.pool.Exec(ctx,
		`DELETE FROM asset_controls WHERE asset_id = $1 AND control_id = $2`,
		assetID, controlID,
	); err != nil {
		return fmt.Errorf("detach control: %w", err)
	}
	return nil
}
