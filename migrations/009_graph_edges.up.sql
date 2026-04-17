-- S → ST: какие источники могут применить какую угрозу
CREATE TABLE source_threats (
    threat_source_id SMALLINT NOT NULL REFERENCES threat_sources(id) ON DELETE CASCADE,
    threat_id        BIGINT   NOT NULL REFERENCES threats(id)        ON DELETE CASCADE,
    PRIMARY KEY (threat_source_id, threat_id)
);
CREATE INDEX idx_st_threat ON source_threats(threat_id);

-- ST → VL: через какие уязвимые звенья реализуется угроза
CREATE TABLE threat_vulnerable_links (
    threat_id        BIGINT NOT NULL REFERENCES threats(id)          ON DELETE CASCADE,
    vulnerability_id BIGINT NOT NULL REFERENCES vulnerabilities(id)  ON DELETE CASCADE,
    PRIMARY KEY (threat_id, vulnerability_id)
);
CREATE INDEX idx_tvl_vuln ON threat_vulnerable_links(vulnerability_id);

-- ST → DA: к каким деструктивным действиям ведёт угроза
CREATE TABLE threat_destructive_actions (
    threat_id             BIGINT   NOT NULL REFERENCES threats(id)              ON DELETE CASCADE,
    destructive_action_id SMALLINT NOT NULL REFERENCES destructive_actions(id)  ON DELETE CASCADE,
    PRIMARY KEY (threat_id, destructive_action_id)
);
CREATE INDEX idx_tda_da ON threat_destructive_actions(destructive_action_id);

-- VL → Control: какой метод противодействия закрывает какое уязвимое звено
CREATE TABLE vulnerability_controls (
    vulnerability_id BIGINT NOT NULL REFERENCES vulnerabilities(id) ON DELETE CASCADE,
    control_id       BIGINT NOT NULL REFERENCES controls(id)        ON DELETE CASCADE,
    coverage         NUMERIC(3,2) NOT NULL DEFAULT 1.0 CHECK (coverage BETWEEN 0 AND 1),
    PRIMARY KEY (vulnerability_id, control_id)
);
CREATE INDEX idx_vc_control ON vulnerability_controls(control_id);
