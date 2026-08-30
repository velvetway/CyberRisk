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
// для каждой — контроли, закрывающие эту категорию И внедрённые на активе,
// плюс presence-индикатор: сколько активных asset_vulnerabilities этой
// категории сейчас открыто на активе (по сути — счётчик CVE).
//
// VL-категории берутся из таблицы vl_categories (6 шт из диплома). Связь
// «угроза → VL-категория» хранится в threat_vulnerable_links.
//
// Поле Uncovered = «у угрозы есть эта VL-категория, на активе нет ни
// одного контроля, который её закрывает». P6 добавил presence_count
// (число CVE/БДУ-записей соответствующей категории на активе) — это
// поле использует UI для предупреждения «найдено N свидетельств».
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
),
presence AS (
    SELECT vl_category_id, COUNT(*) AS cnt
    FROM asset_vulnerabilities
    WHERE asset_id = $1
      AND vl_category_id IS NOT NULL
      AND status IN ('open','in_progress','mitigated')
    GROUP BY vl_category_id
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
       ) AS uncovered,
       COALESCE((SELECT cnt FROM presence p WHERE p.vl_category_id = tv.id), 0) AS presence_count
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
		if err := rows.Scan(&v.CategoryID, &v.Code, &v.Name, &v.Description, &raw, &v.Uncovered, &v.PresenceCount); err != nil {
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
