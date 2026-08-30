package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// VLCategoryReader предоставляет минимальный read-only доступ к таблице
// vl_categories. Используется сервисами, которым нужен mapping code → id
// (например, asset_vulnerability при автодетекции CVE через CWE).
type VLCategoryReader interface {
	LoadVLCategoryIDs(ctx context.Context) (map[string]int16, error)
}

type vlCategoryRepo struct {
	pool *pgxpool.Pool
}

func NewVLCategoryReader(pool *pgxpool.Pool) VLCategoryReader {
	return &vlCategoryRepo{pool: pool}
}

func (r *vlCategoryRepo) LoadVLCategoryIDs(ctx context.Context) (map[string]int16, error) {
	rows, err := r.pool.Query(ctx, `SELECT code, id FROM vl_categories`)
	if err != nil {
		return nil, fmt.Errorf("load vl_categories: %w", err)
	}
	defer rows.Close()

	out := make(map[string]int16, 6)
	for rows.Next() {
		var code string
		var id int16
		if err := rows.Scan(&code, &id); err != nil {
			return nil, err
		}
		out[code] = id
	}
	return out, rows.Err()
}
