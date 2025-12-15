-- ============================
-- ENUM TYPES
-- ============================

CREATE TYPE user_role AS ENUM ('admin', 'auditor', 'viewer');

CREATE TYPE asset_environment AS ENUM ('prod', 'test', 'dev', 'other');

CREATE TYPE asset_vuln_status AS ENUM ('open', 'in_progress', 'mitigated', 'accepted');

CREATE TYPE risk_status AS ENUM ('open', 'in_review', 'mitigated', 'accepted');

CREATE TYPE recommendation_status AS ENUM ('planned', 'in_progress', 'implemented', 'rejected');

CREATE TYPE risk_level AS ENUM ('low', 'medium', 'high', 'critical');

CREATE TYPE threat_source_type AS ENUM ('external', 'internal', 'third_party');


-- ============================
-- USERS
-- ============================

CREATE TABLE users (
                       id              BIGSERIAL PRIMARY KEY,
                       username        VARCHAR(64) NOT NULL UNIQUE,
                       password_hash   TEXT        NOT NULL,
                       role            user_role   NOT NULL,
                       is_active       BOOLEAN     NOT NULL DEFAULT TRUE,
                       created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
                       updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Триггеры updated_at можно добавить позже (или обновлять из приложения).


-- ============================
-- ASSET TYPES
-- ============================

CREATE TABLE asset_types (
                             id          SMALLSERIAL PRIMARY KEY,
                             name        VARCHAR(64) NOT NULL UNIQUE,
                             description TEXT
);


-- ============================
-- ASSETS
-- ============================

CREATE TABLE assets (
                        id                      BIGSERIAL PRIMARY KEY,
                        name                    TEXT        NOT NULL,
                        asset_type_id           SMALLINT    REFERENCES asset_types(id),
                        owner                   TEXT,
                        description             TEXT,
                        business_criticality    SMALLINT    NOT NULL CHECK (business_criticality BETWEEN 1 AND 5),
                        confidentiality         SMALLINT    NOT NULL CHECK (confidentiality    BETWEEN 1 AND 5),
                        integrity               SMALLINT    NOT NULL CHECK (integrity          BETWEEN 1 AND 5),
                        availability            SMALLINT    NOT NULL CHECK (availability       BETWEEN 1 AND 5),
                        environment             asset_environment NOT NULL DEFAULT 'prod',
                        tags                    JSONB,
                        created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
                        updated_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_assets_asset_type_id ON assets(asset_type_id);
CREATE INDEX idx_assets_env           ON assets(environment);


-- ============================
-- THREAT CATEGORIES
-- ============================

CREATE TABLE threat_categories (
                                   id          SMALLSERIAL PRIMARY KEY,
                                   name        VARCHAR(64) NOT NULL UNIQUE,
                                   description TEXT
);


-- ============================
-- THREATS
-- ============================

CREATE TABLE threats (
                         id                  BIGSERIAL PRIMARY KEY,
                         name                TEXT                 NOT NULL,
                         threat_category_id  SMALLINT             REFERENCES threat_categories(id),
                         source_type         threat_source_type   NOT NULL,
                         description         TEXT,
                         base_likelihood     SMALLINT             NOT NULL CHECK (base_likelihood BETWEEN 1 AND 5),
                         created_at          TIMESTAMPTZ          NOT NULL DEFAULT now(),
                         updated_at          TIMESTAMPTZ          NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX uniq_threats_name ON threats(name);
CREATE INDEX idx_threats_category ON threats(threat_category_id);


-- ============================
-- VULNERABILITY CATEGORIES
-- ============================

CREATE TABLE vulnerability_categories (
                                          id          SMALLSERIAL PRIMARY KEY,
                                          name        VARCHAR(64) NOT NULL UNIQUE,
                                          description TEXT
);


-- ============================
-- VULNERABILITIES
-- ============================

CREATE TABLE vulnerabilities (
                                 id                          BIGSERIAL PRIMARY KEY,
                                 name                        TEXT        NOT NULL,
                                 vulnerability_category_id   SMALLINT    REFERENCES vulnerability_categories(id),
                                 description                 TEXT,
                                 severity                    SMALLINT    NOT NULL CHECK (severity BETWEEN 1 AND 10),
                                 affects_asset_type_id       SMALLINT    REFERENCES asset_types(id),
                                 created_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
                                 updated_at                  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_vuln_category       ON vulnerabilities(vulnerability_category_id);
CREATE INDEX idx_vuln_aff_asset_type ON vulnerabilities(affects_asset_type_id);


-- ============================
-- ASSET VULNERABILITIES (связь asset ↔ vulnerability)
-- ============================

CREATE TABLE asset_vulnerabilities (
                                       id               BIGSERIAL PRIMARY KEY,
                                       asset_id         BIGINT     NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
                                       vulnerability_id BIGINT     NOT NULL REFERENCES vulnerabilities(id) ON DELETE CASCADE,
                                       status           asset_vuln_status NOT NULL DEFAULT 'open',
                                       created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
                                       updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
                                       CONSTRAINT uniq_asset_vulnerability UNIQUE (asset_id, vulnerability_id)
);

CREATE INDEX idx_asset_vuln_asset   ON asset_vulnerabilities(asset_id);
CREATE INDEX idx_asset_vuln_status  ON asset_vulnerabilities(status);


-- ============================
-- CONTROL TYPES (типы мер защиты)
-- ============================

CREATE TABLE control_types (
                               id          SMALLSERIAL PRIMARY KEY,
                               name        VARCHAR(64) NOT NULL UNIQUE,
                               description TEXT
);


-- ============================
-- CONTROLS (меры защиты)
-- ============================

CREATE TABLE controls (
                          id                      BIGSERIAL PRIMARY KEY,
                          name                    TEXT        NOT NULL,
                          control_type_id         SMALLINT    REFERENCES control_types(id),
                          description             TEXT,
                          reduces_likelihood_by   NUMERIC(3,2) NOT NULL DEFAULT 0.0 CHECK (reduces_likelihood_by BETWEEN 0 AND 1),
                          reduces_impact_by       NUMERIC(3,2) NOT NULL DEFAULT 0.0 CHECK (reduces_impact_by    BETWEEN 0 AND 1),
                          created_at              TIMESTAMPTZ  NOT NULL DEFAULT now(),
                          updated_at              TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX uniq_controls_name   ON controls(name);
CREATE INDEX idx_controls_type          ON controls(control_type_id);


-- ============================
-- ASSET CONTROLS (мера защиты, применённая к активу)
-- ============================

CREATE TABLE asset_controls (
                                id              BIGSERIAL PRIMARY KEY,
                                asset_id        BIGINT     NOT NULL REFERENCES assets(id)   ON DELETE CASCADE,
                                control_id      BIGINT     NOT NULL REFERENCES controls(id) ON DELETE CASCADE,
                                effectiveness   NUMERIC(3,2) NOT NULL DEFAULT 1.0 CHECK (effectiveness BETWEEN 0 AND 1),
                                created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
                                updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
                                CONSTRAINT uniq_asset_control UNIQUE (asset_id, control_id)
);

CREATE INDEX idx_asset_controls_asset   ON asset_controls(asset_id);
CREATE INDEX idx_asset_controls_control ON asset_controls(control_id);


-- ============================
-- RISK SCENARIOS
-- ============================

CREATE TABLE risk_scenarios (
                                id                  BIGSERIAL PRIMARY KEY,
                                asset_id            BIGINT   NOT NULL REFERENCES assets(id)      ON DELETE CASCADE,
                                threat_id           BIGINT   NOT NULL REFERENCES threats(id),
                                vulnerability_id    BIGINT   REFERENCES vulnerabilities(id),
                                title               TEXT     NOT NULL,
                                description         TEXT,
                                impact              NUMERIC(4,2),   -- 1–5 (после нормализации)
                                likelihood          NUMERIC(4,2),   -- 1–5 (после нормализации)
                                risk_score          NUMERIC(5,2),   -- ~1–25
                                risk_level          risk_level,
                                status              risk_status     NOT NULL DEFAULT 'open',
                                created_by_user_id  BIGINT  REFERENCES users(id),
                                updated_by_user_id  BIGINT  REFERENCES users(id),
                                created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
                                updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_risk_scenarios_asset      ON risk_scenarios(asset_id);
CREATE INDEX idx_risk_scenarios_threat     ON risk_scenarios(threat_id);
CREATE INDEX idx_risk_scenarios_vuln       ON risk_scenarios(vulnerability_id);
CREATE INDEX idx_risk_scenarios_level      ON risk_scenarios(risk_level);
CREATE INDEX idx_risk_scenarios_status     ON risk_scenarios(status);
CREATE INDEX idx_risk_scenarios_heatmap    ON risk_scenarios(impact, likelihood);


-- ============================
-- RECOMMENDATION TEMPLATES (справочник)
-- ============================

CREATE TABLE recommendation_templates (
                                          id                      BIGSERIAL PRIMARY KEY,
                                          code                    VARCHAR(64) NOT NULL UNIQUE,
                                          title                   TEXT        NOT NULL,
                                          description             TEXT,
                                          asset_type_id           SMALLINT    REFERENCES asset_types(id),
                                          threat_category_id      SMALLINT    REFERENCES threat_categories(id),
                                          vulnerability_category_id SMALLINT  REFERENCES vulnerability_categories(id),
                                          min_risk_level          risk_level,
                                          created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
                                          updated_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_recom_tpl_asset_type   ON recommendation_templates(asset_type_id);
CREATE INDEX idx_recom_tpl_threat_cat   ON recommendation_templates(threat_category_id);
CREATE INDEX idx_recom_tpl_vuln_cat     ON recommendation_templates(vulnerability_category_id);


-- ============================
-- RISK SCENARIO RECOMMENDATIONS (фактические рекомендации по сценарию)
-- ============================

CREATE TABLE risk_scenario_recommendations (
                                               id                          BIGSERIAL PRIMARY KEY,
                                               risk_scenario_id            BIGINT  NOT NULL REFERENCES risk_scenarios(id) ON DELETE CASCADE,
                                               recommendation_template_id  BIGINT  NOT NULL REFERENCES recommendation_templates(id),
                                               status                      recommendation_status NOT NULL DEFAULT 'planned',
                                               comment                     TEXT,
                                               created_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
                                               updated_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
                                               CONSTRAINT uniq_risk_recom UNIQUE (risk_scenario_id, recommendation_template_id)
);

CREATE INDEX idx_risk_recom_scenario ON risk_scenario_recommendations(risk_scenario_id);
CREATE INDEX idx_risk_recom_status   ON risk_scenario_recommendations(status);
