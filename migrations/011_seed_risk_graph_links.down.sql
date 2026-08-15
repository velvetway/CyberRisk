DELETE FROM vulnerability_controls vc
USING vulnerabilities v, controls c, (VALUES
    ('Unpatched Web Framework', 'Regular Patch Management'),
    ('Unpatched Web Framework', 'Web Application Firewall (WAF)'),
    ('Weak Database Authentication', 'Database Access Control'),
    ('Missing Network Segmentation', 'Network Segmentation'),
    ('Missing Network Segmentation', 'Intrusion Detection System'),
    ('Outdated SSL/TLS Configuration', 'Regular Patch Management'),
    ('Outdated SSL/TLS Configuration', 'Intrusion Detection System'),
    ('Insufficient Input Validation', 'Web Application Firewall (WAF)')
) AS m(vulnerability_name, control_name)
WHERE vc.vulnerability_id = v.id
  AND vc.control_id = c.id
  AND v.name = m.vulnerability_name
  AND c.name = m.control_name;

DELETE FROM threat_destructive_actions tda
USING threats t, destructive_actions da, (VALUES
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
WHERE tda.threat_id = t.id
  AND tda.destructive_action_id = da.id
  AND t.name = m.threat_name
  AND da.code = m.action_code;

DELETE FROM threat_vulnerable_links tvl
USING threats t, vulnerabilities v, (VALUES
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
WHERE tvl.threat_id = t.id
  AND tvl.vulnerability_id = v.id
  AND t.name = m.threat_name
  AND v.name = m.vulnerability_name;

DELETE FROM source_threats st
USING threat_sources ts, threats t, (VALUES
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
WHERE st.threat_source_id = ts.id
  AND st.threat_id = t.id
  AND ts.code = m.source_code
  AND t.name = m.threat_name;
