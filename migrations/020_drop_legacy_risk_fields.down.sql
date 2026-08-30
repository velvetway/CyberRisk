-- Reverse the legacy purge. This is best-effort (does not restore row data
-- of the dropped columns) but rebuilds enough schema to let the older
-- migrations replay against an empty database.

CREATE TYPE recommendation_status AS ENUM ('planned', 'in_progress', 'implemented', 'rejected');
CREATE TYPE risk_level AS ENUM ('low', 'medium', 'high', 'critical');
CREATE TYPE risk_status AS ENUM ('open', 'in_review', 'mitigated', 'accepted');

CREATE TYPE kii_category AS ENUM ('none', 'cat3', 'cat2', 'cat1');
CREATE TYPE protection_level AS ENUM ('uz1', 'uz2', 'uz3', 'uz4');
CREATE TYPE data_category_type AS ENUM (
    'public', 'internal', 'confidential',
    'personal_data', 'personal_data_special', 'personal_data_biometric',
    'kii', 'state_secret', 'banking_secret', 'medical_secret', 'commercial_secret'
);

ALTER TABLE asset_controls
    ADD COLUMN effectiveness NUMERIC(3,2) NOT NULL DEFAULT 1.0
        CHECK (effectiveness BETWEEN 0 AND 1);

ALTER TABLE controls
    ADD COLUMN reduces_likelihood_by NUMERIC(3,2) NOT NULL DEFAULT 0.0
        CHECK (reduces_likelihood_by BETWEEN 0 AND 1),
    ADD COLUMN reduces_impact_by     NUMERIC(3,2) NOT NULL DEFAULT 0.0
        CHECK (reduces_impact_by BETWEEN 0 AND 1);

ALTER TABLE threats
    ADD COLUMN base_likelihood        SMALLINT NOT NULL DEFAULT 3
        CHECK (base_likelihood BETWEEN 1 AND 5),
    ADD COLUMN attack_vector          TEXT,
    ADD COLUMN impact_confidentiality BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN impact_integrity       BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN impact_availability    BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE assets
    ADD COLUMN type                   VARCHAR(64),
    ADD COLUMN location               TEXT,
    ADD COLUMN business_criticality   SMALLINT NOT NULL DEFAULT 3
        CHECK (business_criticality BETWEEN 1 AND 5),
    ADD COLUMN confidentiality        SMALLINT NOT NULL DEFAULT 3
        CHECK (confidentiality BETWEEN 1 AND 5),
    ADD COLUMN integrity              SMALLINT NOT NULL DEFAULT 3
        CHECK (integrity BETWEEN 1 AND 5),
    ADD COLUMN availability           SMALLINT NOT NULL DEFAULT 3
        CHECK (availability BETWEEN 1 AND 5),
    ADD COLUMN data_category          data_category_type DEFAULT 'internal',
    ADD COLUMN protection_level       protection_level,
    ADD COLUMN kii_category           kii_category DEFAULT 'none',
    ADD COLUMN has_personal_data      BOOLEAN DEFAULT FALSE,
    ADD COLUMN personal_data_volume   VARCHAR(32),
    ADD COLUMN has_internet_access    BOOLEAN DEFAULT TRUE;

CREATE INDEX IF NOT EXISTS idx_assets_type             ON assets(type);
CREATE INDEX IF NOT EXISTS idx_assets_data_category    ON assets(data_category);
CREATE INDEX IF NOT EXISTS idx_assets_kii              ON assets(kii_category);
CREATE INDEX IF NOT EXISTS idx_assets_protection_level ON assets(protection_level);

CREATE TABLE risk_scenarios (
    id                  BIGSERIAL PRIMARY KEY,
    asset_id            BIGINT   NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    threat_id           BIGINT   NOT NULL REFERENCES threats(id),
    vulnerability_id    BIGINT   REFERENCES vulnerabilities(id),
    title               TEXT     NOT NULL,
    description         TEXT,
    impact              NUMERIC(4,2),
    likelihood          NUMERIC(4,2),
    risk_score          NUMERIC(5,2),
    risk_level          risk_level,
    status              risk_status NOT NULL DEFAULT 'open',
    created_by_user_id  BIGINT REFERENCES users(id),
    updated_by_user_id  BIGINT REFERENCES users(id),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE recommendation_templates (
    id                        BIGSERIAL PRIMARY KEY,
    code                      VARCHAR(64) NOT NULL UNIQUE,
    title                     TEXT        NOT NULL,
    description               TEXT,
    asset_type_id             SMALLINT REFERENCES asset_types(id),
    threat_category_id        SMALLINT REFERENCES threat_categories(id),
    vulnerability_category_id SMALLINT REFERENCES vulnerability_categories(id),
    min_risk_level            risk_level,
    created_at                TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE risk_scenario_recommendations (
    id                          BIGSERIAL PRIMARY KEY,
    risk_scenario_id            BIGINT NOT NULL REFERENCES risk_scenarios(id) ON DELETE CASCADE,
    recommendation_template_id  BIGINT NOT NULL REFERENCES recommendation_templates(id),
    status                      recommendation_status NOT NULL DEFAULT 'planned',
    comment                     TEXT,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uniq_risk_recom UNIQUE (risk_scenario_id, recommendation_template_id)
);
