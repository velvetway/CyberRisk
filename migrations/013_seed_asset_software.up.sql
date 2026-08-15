WITH pairs(asset_name, software_name, vendor, version) AS (
    VALUES
        ('Web Application Server', 'Windows Server', 'Microsoft', '2019'),
        ('Web Application Server', 'PostgreSQL', 'PostgreSQL Global', '15'),
        ('Web Application Server', 'Docker', 'Docker Inc', '24'),
        ('Customer Database', 'PostgreSQL', 'PostgreSQL Global', '14'),
        ('Employee Workstation Network', 'Kaspersky Endpoint Security', 'АО «Лаборатория Касперского»', NULL),
        ('Development Server', 'Red Hat Enterprise Linux', 'Red Hat', '9'),
        ('Development Server', 'Docker', 'Docker Inc', '24'),
        ('Mobile App Backend', 'Astra Linux Common Edition', 'ООО «РусБИТех-Астра»', '2.12')
)
INSERT INTO asset_software (asset_id, software_id, version, notes)
SELECT a.id, s.id, p.version, 'seeded for bdu sync demo'
FROM pairs p
JOIN assets a ON a.name = p.asset_name
JOIN LATERAL (
    SELECT id
    FROM software_catalog
    WHERE name = p.software_name AND vendor = p.vendor
    ORDER BY id
    LIMIT 1
) s ON true
ON CONFLICT (asset_id, software_id) DO NOTHING;
