-- Seed data for demonstration purposes
-- This migration adds sample assets, threats, vulnerabilities for testing

-- ============================================
-- 1. ASSET TYPES (справочник)
-- ============================================

INSERT INTO asset_types (id, name, description) VALUES
    (1, 'Server', 'Physical or virtual server infrastructure'),
    (2, 'Database', 'Database management systems'),
    (3, 'Application', 'Software applications and services'),
    (4, 'Network', 'Network infrastructure and equipment'),
    (5, 'Workstation', 'Employee workstations and desktops'),
    (6, 'Mobile', 'Mobile devices'),
    (7, 'IoT', 'Internet of Things devices'),
    (8, 'Cloud', 'Cloud-based resources and services')
ON CONFLICT (name) DO NOTHING;

-- ============================================
-- 2. THREAT CATEGORIES (справочник)
-- ============================================

INSERT INTO threat_categories (id, name, description) VALUES
    (1, 'Exploit', 'Exploitation of software vulnerabilities'),
    (2, 'Malware', 'Malicious software including viruses, trojans, ransomware'),
    (3, 'DoS/DDoS', 'Denial of service attacks'),
    (4, 'Information Disclosure', 'Unauthorized access to sensitive information'),
    (5, 'Phishing', 'Social engineering through deceptive communications'),
    (6, 'Insider Threat', 'Threats from internal employees or contractors'),
    (7, 'Physical', 'Physical security threats')
ON CONFLICT (name) DO NOTHING;

-- ============================================
-- 3. VULNERABILITY CATEGORIES (справочник)
-- ============================================

INSERT INTO vulnerability_categories (id, name, description) VALUES
    (1, 'Injection', 'SQL, NoSQL, LDAP, OS command injection'),
    (2, 'Authentication', 'Broken authentication and session management'),
    (3, 'Access Control', 'Broken access control and privilege escalation'),
    (4, 'Configuration', 'Security misconfiguration'),
    (5, 'Cryptography', 'Cryptographic failures'),
    (6, 'Outdated Software', 'Vulnerable and outdated components'),
    (7, 'Logging', 'Insufficient logging and monitoring')
ON CONFLICT (name) DO NOTHING;

-- ============================================
-- 4. CONTROL TYPES (справочник)
-- ============================================

INSERT INTO control_types (id, name, description) VALUES
    (1, 'Preventive', 'Controls that prevent security incidents'),
    (2, 'Detective', 'Controls that detect security incidents'),
    (3, 'Corrective', 'Controls that correct security incidents'),
    (4, 'Compensating', 'Alternative controls when primary controls are not feasible')
ON CONFLICT (name) DO NOTHING;

-- ============================================
-- 5. ASSETS (демонстрационные данные)
-- ============================================

INSERT INTO assets (name, asset_type_id, type, owner, description, location, business_criticality, confidentiality, integrity, availability, environment) VALUES
    ('Web Application Server', 1, 'server', 'IT Department', 'Production web application server hosting customer-facing portal', 'Datacenter A, Rack 12', 5, 4, 5, 5, 'prod'),
    ('Customer Database', 2, 'database', 'Data Team', 'PostgreSQL database containing customer personal information and transaction history', 'Datacenter A, Rack 15', 5, 5, 5, 4, 'prod'),
    ('Employee Workstation Network', 4, 'network', 'IT Department', 'Internal network segment for employee workstations and office equipment', 'Office Building, VLAN 100', 3, 3, 3, 4, 'prod'),
    ('Development Server', 1, 'server', 'Development Team', 'Development and testing environment server', 'Datacenter B, Rack 5', 2, 2, 2, 3, 'dev'),
    ('Mobile App Backend', 3, 'application', 'Mobile Team', 'API backend for mobile applications', 'Cloud AWS', 4, 4, 4, 5, 'prod');

-- ============================================
-- 6. THREATS (демонстрационные данные)
-- ============================================

INSERT INTO threats (name, threat_category_id, source_type, description, base_likelihood) VALUES
    ('SQL Injection Attack', 1, 'external', 'Malicious SQL code injection through web application input fields to extract or modify database data', 4),
    ('Ransomware Attack', 2, 'external', 'Malware that encrypts critical data and demands ransom payment for decryption key', 3),
    ('DDoS Attack', 3, 'external', 'Distributed Denial of Service attack flooding servers with traffic to make services unavailable', 3),
    ('Insider Data Leak', 4, 'internal', 'Unauthorized disclosure of sensitive information by employees with legitimate access', 2),
    ('Brute Force Attack', 1, 'external', 'Automated attempts to guess passwords and gain unauthorized access', 4),
    ('Phishing Campaign', 5, 'external', 'Deceptive emails targeting employees to steal credentials or install malware', 4);

-- ============================================
-- 7. VULNERABILITIES (демонстрационные данные)
-- ============================================

INSERT INTO vulnerabilities (name, vulnerability_category_id, description, severity, affects_asset_type_id) VALUES
    ('Unpatched Web Framework', 6, 'Web application framework version contains known SQL injection vulnerability, patch available but not yet applied', 4, 1),
    ('Weak Database Authentication', 2, 'Database uses default credentials and lacks multi-factor authentication', 4, 2),
    ('Missing Network Segmentation', 4, 'Insufficient network segmentation allows lateral movement between office network and critical servers', 3, 4),
    ('Outdated SSL/TLS Configuration', 5, 'Server uses deprecated TLS 1.0/1.1 protocols vulnerable to downgrade attacks', 3, 1),
    ('Insufficient Input Validation', 1, 'Application does not properly validate and sanitize user inputs', 5, 3);

-- ============================================
-- 8. ASSET-VULNERABILITY RELATIONSHIPS
-- ============================================

INSERT INTO asset_vulnerabilities (asset_id, vulnerability_id, status) VALUES
    (1, 1, 'open'),
    (1, 4, 'in_progress'),
    (2, 2, 'open'),
    (3, 3, 'open'),
    (5, 5, 'mitigated');

-- ============================================
-- 9. CONTROLS (демонстрационные данные)
-- ============================================

INSERT INTO controls (name, control_type_id, description, reduces_likelihood_by, reduces_impact_by) VALUES
    ('Web Application Firewall (WAF)', 1, 'Deploy WAF to filter malicious SQL injection attempts at network perimeter', 0.40, 0.20),
    ('Regular Patch Management', 1, 'Establish automated patch management process with monthly security updates', 0.50, 0.10),
    ('Database Access Control', 1, 'Implement role-based access control and multi-factor authentication for database', 0.35, 0.25),
    ('Network Segmentation', 1, 'Redesign network architecture with proper VLAN segmentation and firewall rules', 0.30, 0.30),
    ('Backup and Recovery System', 2, 'Implement automated backup system with offsite storage for ransomware protection', 0.10, 0.50),
    ('Intrusion Detection System', 2, 'Deploy IDS to monitor network traffic for suspicious activity', 0.20, 0.15),
    ('Security Awareness Training', 1, 'Conduct quarterly security awareness training for all employees', 0.25, 0.05);

-- ============================================
-- 10. ASSET CONTROLS (демонстрационные данные)
-- ============================================

INSERT INTO asset_controls (asset_id, control_id, effectiveness) VALUES
    (1, 1, 0.80),
    (1, 2, 0.90),
    (2, 3, 0.85),
    (2, 5, 0.95),
    (3, 4, 0.70),
    (3, 6, 0.75);

-- ============================================
-- 11. RECOMMENDATION TEMPLATES (справочник)
-- ============================================

INSERT INTO recommendation_templates (code, title, description, asset_type_id, threat_category_id, vulnerability_category_id, min_risk_level) VALUES
    ('REC-001', 'Immediate Framework Patching', 'Apply security patch for web framework within 48 hours to address critical SQL injection vulnerability', 1, 1, 6, 'high'),
    ('REC-002', 'Strengthen Database Security', 'Update database credentials, enable MFA, and implement least-privilege access controls', 2, NULL, 2, 'medium'),
    ('REC-003', 'Deploy WAF Solution', 'Implement Web Application Firewall to provide defense-in-depth against SQL injection and other web attacks', 1, 1, 1, 'high'),
    ('REC-004', 'Implement Network Segmentation', 'Redesign network topology to isolate critical assets from general office network', 4, NULL, 4, 'medium'),
    ('REC-005', 'Security Awareness Training', 'Conduct quarterly security awareness training for all employees to reduce insider threat risks', NULL, 6, NULL, 'low'),
    ('REC-006', 'Backup Strategy Implementation', 'Implement 3-2-1 backup strategy with regular testing of restoration procedures', NULL, 2, NULL, 'high'),
    ('REC-007', 'TLS Configuration Update', 'Upgrade to TLS 1.3, disable deprecated protocols and weak cipher suites', 1, NULL, 5, 'medium');
