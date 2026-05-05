package repository

import (
	"context"
	"encoding/json"

	"Diplom/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RiskGraphRepository interface {
	LoadVulnerableLinks(ctx context.Context, assetID, threatID int64) ([]domain.VLNode, error)
}

type riskGraphRepository struct {
	pool *pgxpool.Pool
}

func NewRiskGraphRepository(pool *pgxpool.Pool) RiskGraphRepository {
	return &riskGraphRepository{pool: pool}
}

// LoadVulnerableLinks возвращает VL-категории, прикреплённые к угрозе,
// а для каждой — контроли, закрывающие эту категорию И внедрённые на активе.
//
// VL-категории берутся из таблицы vl_categories (6 шт из диплома). Связь
// «угроза → VL-категория» хранится в threat_vulnerable_links (после
// миграции 034 — это FK на vl_categories.id, не на vulnerabilities.id).
//
// Поле Uncovered семантически = «эта VL-категория есть у угрозы, но на
// активе нет ни одного контроля, который её закрывает». В P6 это будет
// уточнено через asset_vulnerabilities.vl_category_id (presence facts).
func (r *riskGraphRepository) LoadVulnerableLinks(ctx context.Context, assetID, threatID int64) ([]domain.VLNode, error) {
	const q = `
WITH threat_vls AS (
    SELECT vlc.id, vlc.code, vlc.name, vlc.description
    FROM vl_categories vlc
    JOIN threat_vulnerable_links tvl ON tvl.vl_category_id = vlc.id
    WHERE tvl.threat_id = $2
),
covering_controls AS (
    SELECT vcc.vl_category_id,
           c.id   AS control_id,
           c.name AS control_name,
           vcc.coverage
    FROM vl_category_controls vcc
    JOIN controls c        ON c.id = vcc.control_id
    JOIN asset_controls ac ON ac.control_id = c.id AND ac.asset_id = $1
)
SELECT tv.id,
       tv.code,
       tv.name,
       COALESCE(tv.description, ''),
       COALESCE(
         json_agg(json_build_object(
           'id', cc.control_id,
           'name', cc.control_name,
           'coverage', cc.coverage
         )) FILTER (WHERE cc.control_id IS NOT NULL),
         '[]'::json
       ) AS controls_json,
       NOT EXISTS (
         SELECT 1 FROM covering_controls cc2 WHERE cc2.vl_category_id = tv.id
       ) AS uncovered
FROM threat_vls tv
LEFT JOIN covering_controls cc ON cc.vl_category_id = tv.id
GROUP BY tv.id, tv.code, tv.name, tv.description
ORDER BY tv.code`

	rows, err := r.pool.Query(ctx, q, assetID, threatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.VLNode, 0)
	for rows.Next() {
		var v domain.VLNode
		var raw []byte
		if err := rows.Scan(&v.CategoryID, &v.Code, &v.Name, &v.Description, &raw, &v.Uncovered); err != nil {
			return nil, err
		}
		if len(raw) > 0 {
			if err := json.Unmarshal(raw, &v.CoverageControls); err != nil {
				return nil, err
			}
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
