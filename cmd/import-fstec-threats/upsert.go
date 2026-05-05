package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type importStats struct {
	threatsUpserted int
	sourceLinks     int
	vlLinks         int
	withTypes       int
	withoutTypes    int
}

// upsertAll loads in-memory rows into Postgres in a single transaction.
// We:
//  1. cache the asset_types name → id map and threat_categories name → id map
//  2. cache the threat_sources code → id map (S1..S4)
//  3. for each threat row, INSERT … ON CONFLICT (bdu_id) DO UPDATE
//  4. rebuild source_threats(threat_id, threat_source_id) for that threat.
func upsertAll(ctx context.Context, pool *pgxpool.Pool, rows []threatRow) (importStats, error) {
	var stats importStats

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return stats, err
	}
	defer tx.Rollback(ctx)

	atypes, err := loadIDMap(ctx, tx, `SELECT id, name FROM asset_types`)
	if err != nil {
		return stats, fmt.Errorf("load asset_types: %w", err)
	}
	tcats, err := loadIDMap(ctx, tx, `SELECT id, name FROM threat_categories`)
	if err != nil {
		return stats, fmt.Errorf("load threat_categories: %w", err)
	}
	tsources, err := loadIDMap(ctx, tx, `SELECT id, code FROM threat_sources`)
	if err != nil {
		return stats, fmt.Errorf("load threat_sources: %w", err)
	}
	vlcats, err := loadIDMap(ctx, tx, `SELECT id, code FROM vl_categories`)
	if err != nil {
		return stats, fmt.Errorf("load vl_categories: %w", err)
	}

	const upsertSQL = `
INSERT INTO threats (
    name, threat_category_id, source_type, description,
    q_threat, q_severity, bdu_id,
    applies_to_targets, applies_to_asset_types,
    impact_c, impact_i, impact_a, status,
    updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13, COALESCE($14, now()))
ON CONFLICT (bdu_id) DO UPDATE SET
    name                 = EXCLUDED.name,
    threat_category_id   = EXCLUDED.threat_category_id,
    source_type          = EXCLUDED.source_type,
    description          = EXCLUDED.description,
    q_threat             = EXCLUDED.q_threat,
    q_severity           = EXCLUDED.q_severity,
    applies_to_targets   = EXCLUDED.applies_to_targets,
    applies_to_asset_types = EXCLUDED.applies_to_asset_types,
    impact_c             = EXCLUDED.impact_c,
    impact_i             = EXCLUDED.impact_i,
    impact_a             = EXCLUDED.impact_a,
    status               = EXCLUDED.status,
    updated_at           = EXCLUDED.updated_at
RETURNING id`

	// We need bdu_id to be UNIQUE for the ON CONFLICT to work.
	if _, err := tx.Exec(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS uniq_threats_bdu_id ON threats(bdu_id)`); err != nil {
		return stats, fmt.Errorf("ensure unique bdu_id index: %w", err)
	}

	for _, r := range rows {
		var assetTypeIDs []int16
		for _, name := range r.AssetTypes {
			if id, ok := atypes[name]; ok {
				assetTypeIDs = append(assetTypeIDs, int16(id))
			}
		}
		if len(assetTypeIDs) > 0 {
			stats.withTypes++
		} else {
			stats.withoutTypes++
		}

		var categoryID *int16
		if r.CategoryName != "" {
			if id, ok := tcats[r.CategoryName]; ok {
				v := int16(id)
				categoryID = &v
			}
		}

		var updatedAt *time.Time
		if !r.UpdatedAt.IsZero() {
			updatedAt = &r.UpdatedAt
		}

		var threatID int64
		err := tx.QueryRow(ctx, upsertSQL,
			r.Name,
			categoryID,
			r.Source.SourceType,
			r.Description,
			deriveQThreat(r.Source.ThreatPower),
			deriveQSeverity(r.ImpactC, r.ImpactI, r.ImpactA),
			r.BDUID,
			r.TargetText,
			assetTypeIDs,
			r.ImpactC, r.ImpactI, r.ImpactA,
			r.Status,
			updatedAt,
		).Scan(&threatID)
		if err != nil {
			return stats, fmt.Errorf("upsert threat %s: %w", r.BDUID, err)
		}
		stats.threatsUpserted++

		// Rebuild source_threats edges for this threat.
		if _, err := tx.Exec(ctx, `DELETE FROM source_threats WHERE threat_id = $1`, threatID); err != nil {
			return stats, err
		}
		for _, code := range r.Source.SourceCodes {
			sid, ok := tsources[code]
			if !ok {
				continue
			}
			if _, err := tx.Exec(ctx,
				`INSERT INTO source_threats (threat_source_id, threat_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
				int16(sid), threatID,
			); err != nil {
				return stats, fmt.Errorf("link source %s → %d: %w", code, threatID, err)
			}
			stats.sourceLinks++
		}

		// Rebuild threat_vulnerable_links edges for this threat (ST → VL_category).
		if _, err := tx.Exec(ctx, `DELETE FROM threat_vulnerable_links WHERE threat_id = $1`, threatID); err != nil {
			return stats, err
		}
		for _, code := range r.VLCodes {
			vid, ok := vlcats[code]
			if !ok {
				continue
			}
			if _, err := tx.Exec(ctx,
				`INSERT INTO threat_vulnerable_links (threat_id, vl_category_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
				threatID, int16(vid),
			); err != nil {
				return stats, fmt.Errorf("link VL %s → %d: %w", code, threatID, err)
			}
			stats.vlLinks++
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return stats, err
	}
	return stats, nil
}

// loadIDMap runs `SELECT id, key FROM …` and returns a key→id map.
// Designed for tiny reference tables (≤10 rows).
func loadIDMap(ctx context.Context, tx pgx.Tx, query string) (map[string]int, error) {
	rows, err := tx.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string]int{}
	for rows.Next() {
		var id int
		var key string
		if err := rows.Scan(&id, &key); err != nil {
			return nil, err
		}
		out[key] = id
	}
	return out, rows.Err()
}
