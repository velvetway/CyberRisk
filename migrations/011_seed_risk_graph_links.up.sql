-- Базовые связи для графовой модели ПТСЗИ:
-- S -> ST -> VL -> DA и VL -> Control.

INSERT INTO source_threats (threat_source_id, threat_id)
SELECT ts.id, t.id
FROM (VALUES
    ('S1', 'SQL Injection Attack'),
    ('S4', 'SQL Injection Attack'),
    ('S4', 'Ransomware Attack'),
    ('S1', 'DDoS Attack'),
    ('S4', 'DDoS Attack'),
    ('S2', 'Insider Data Leak'),
    ('S3', 'Insider Data Leak'),
    ('S4', 'Brute Force Attack'),
    ('S1', 'Phishing Campaign'),
    ('S4', 'Phishing Campaign')
) AS m(source_code, threat_name)
JOIN threat_sources ts ON ts.code = m.source_code
JOIN threats t ON t.name = m.threat_name
ON CONFLICT DO NOTHING;

INSERT INTO threat_vulnerable_links (threat_id, vulnerability_id)
SELECT t.id, v.id
FROM (VALUES
    ('SQL Injection Attack', 'Unpatched Web Framework'),
    ('SQL Injection Attack', 'Insufficient Input Validation'),
    ('Ransomware Attack', 'Missing Network Segmentation'),
    ('Ransomware Attack', 'Weak Database Authentication'),
    ('DDoS Attack', 'Missing Network Segmentation'),
    ('Insider Data Leak', 'Weak Database Authentication'),
    ('Insider Data Leak', 'Missing Network Segmentation'),
    ('Brute Force Attack', 'Weak Database Authentication'),
    ('Phishing Campaign', 'Weak Database Authentication'),
    ('Phishing Campaign', 'Missing Network Segmentation')
) AS m(threat_name, vulnerability_name)
JOIN threats t ON t.name = m.threat_name
JOIN vulnerabilities v ON v.name = m.vulnerability_name
ON CONFLICT DO NOTHING;

INSERT INTO threat_destructive_actions (threat_id, destructive_action_id)
SELECT t.id, da.id
FROM (VALUES
    ('SQL Injection Attack', 'DA1'),
    ('SQL Injection Attack', 'DA4'),
    ('Ransomware Attack', 'DA3'),
    ('Ransomware Attack', 'DA5'),
    ('Ransomware Attack', 'DA7'),
    ('DDoS Attack', 'DA5'),
    ('DDoS Attack', 'DA7'),
    ('Insider Data Leak', 'DA1'),
    ('Insider Data Leak', 'DA4'),
    ('Insider Data Leak', 'DA6'),
    ('Brute Force Attack', 'DA1'),
    ('Brute Force Attack', 'DA4'),
    ('Phishing Campaign', 'DA1'),
    ('Phishing Campaign', 'DA6'),
    ('Phishing Campaign', 'DA7')
) AS m(threat_name, action_code)
JOIN threats t ON t.name = m.threat_name
JOIN destructive_actions da ON da.code = m.action_code
ON CONFLICT DO NOTHING;

INSERT INTO vulnerability_controls (vulnerability_id, control_id, coverage)
SELECT v.id, c.id, m.coverage
FROM (VALUES
    ('Unpatched Web Framework', 'Regular Patch Management', 0.90::NUMERIC),
    ('Unpatched Web Framework', 'Web Application Firewall (WAF)', 0.60::NUMERIC),
    ('Weak Database Authentication', 'Database Access Control', 0.85::NUMERIC),
    ('Missing Network Segmentation', 'Network Segmentation', 0.80::NUMERIC),
    ('Missing Network Segmentation', 'Intrusion Detection System', 0.50::NUMERIC),
    ('Outdated SSL/TLS Configuration', 'Regular Patch Management', 0.70::NUMERIC),
    ('Outdated SSL/TLS Configuration', 'Intrusion Detection System', 0.30::NUMERIC),
    ('Insufficient Input Validation', 'Web Application Firewall (WAF)', 0.75::NUMERIC)
) AS m(vulnerability_name, control_name, coverage)
JOIN vulnerabilities v ON v.name = m.vulnerability_name
JOIN controls c ON c.name = m.control_name
ON CONFLICT DO NOTHING;
