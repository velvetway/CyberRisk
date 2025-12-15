-- Миграция: Наполнение справочника ПО
-- Российское ПО из реестра Минцифры и сертифицированные средства защиты

-- =====================================================
-- ОПЕРАЦИОННЫЕ СИСТЕМЫ (category_id = 1)
-- =====================================================
INSERT INTO software_catalog (name, vendor, version, category_id, is_russian, registry_number, description, website, fstec_certified, fstec_certificate_num, fstec_protection_class)
VALUES
('Astra Linux Special Edition', 'ООО «РусБИТех-Астра»', '1.7', 1, TRUE, '1292', 'ОС специального назначения с усиленной защитой', 'https://astralinux.ru', TRUE, '2557', 'до 1 класса защищённости'),
('Astra Linux Common Edition', 'ООО «РусБИТех-Астра»', '2.12', 1, TRUE, '1293', 'ОС общего назначения на базе Debian', 'https://astralinux.ru', FALSE, NULL, NULL),
('ALT Linux', 'ООО «Базальт СПО»', '10', 1, TRUE, '1541', 'Семейство ОС на базе Linux', 'https://basealt.ru', TRUE, '3173', 'до 2 класса защищённости'),
('РЕД ОС', 'ООО «РЕД СОФТ»', '7.3', 1, TRUE, '3751', 'Российская ОС на базе RPM', 'https://redos.red-soft.ru', TRUE, '4060', 'до 2 класса защищённости'),
('ROSA Linux', 'ООО «НТЦ ИТ РОСА»', '12', 1, TRUE, '1543', 'Российский дистрибутив Linux', 'https://rosa.ru', TRUE, '3598', 'до 3 класса защищённости'),
('Windows Server', 'Microsoft', '2022', 1, FALSE, NULL, 'Серверная ОС Microsoft', 'https://microsoft.com', FALSE, NULL, NULL),
('Ubuntu Server', 'Canonical', '22.04 LTS', 1, FALSE, NULL, 'Серверная ОС Ubuntu', 'https://ubuntu.com', FALSE, NULL, NULL),
('CentOS', 'Red Hat', '7/Stream', 1, FALSE, NULL, 'Серверная ОС на базе RHEL', 'https://centos.org', FALSE, NULL, NULL);

-- =====================================================
-- СУБД (category_id = 2)
-- =====================================================
INSERT INTO software_catalog (name, vendor, version, category_id, is_russian, registry_number, description, website, fstec_certified, fstec_certificate_num, fstec_protection_class)
VALUES
('PostgresPro', 'ООО «Постгрес Профессиональный»', '15', 2, TRUE, '3334', 'Российская СУБД на базе PostgreSQL', 'https://postgrespro.ru', TRUE, '4063', 'до 1 класса защищённости'),
('PostgresPro Enterprise', 'ООО «Постгрес Профессиональный»', '15', 2, TRUE, '3335', 'Корпоративная версия PostgresPro', 'https://postgrespro.ru', TRUE, '4064', 'до 1 класса защищённости'),
('Jatoba', 'ООО «Газинформсервис»', '4', 2, TRUE, '6945', 'Российская СУБД', 'https://gaz-is.ru', TRUE, '4127', 'до 2 класса защищённости'),
('Tarantool', 'VK (Mail.ru Group)', '2.10', 2, TRUE, '4228', 'In-memory СУБД и сервер приложений', 'https://tarantool.io', FALSE, NULL, NULL),
('ClickHouse', 'Яндекс', '23', 2, TRUE, '5181', 'Колоночная СУБД для аналитики', 'https://clickhouse.com', FALSE, NULL, NULL),
('PostgreSQL', 'PostgreSQL Global', '15', 2, FALSE, NULL, 'Открытая СУБД', 'https://postgresql.org', FALSE, NULL, NULL),
('MySQL', 'Oracle', '8.0', 2, FALSE, NULL, 'Реляционная СУБД', 'https://mysql.com', FALSE, NULL, NULL),
('Microsoft SQL Server', 'Microsoft', '2022', 2, FALSE, NULL, 'Корпоративная СУБД Microsoft', 'https://microsoft.com', FALSE, NULL, NULL),
('Oracle Database', 'Oracle', '21c', 2, FALSE, NULL, 'Корпоративная СУБД Oracle', 'https://oracle.com', FALSE, NULL, NULL);

-- =====================================================
-- АНТИВИРУСЫ (category_id = 3)
-- =====================================================
INSERT INTO software_catalog (name, vendor, version, category_id, is_russian, registry_number, description, website, fstec_certified, fstec_certificate_num, fstec_protection_class)
VALUES
('Kaspersky Endpoint Security', 'АО «Лаборатория Касперского»', '12', 3, TRUE, '205', 'Антивирусная защита рабочих станций', 'https://kaspersky.ru', TRUE, '4068', 'до 1 класса защищённости'),
('Kaspersky Security Center', 'АО «Лаборатория Касперского»', '14', 3, TRUE, '206', 'Центр управления защитой', 'https://kaspersky.ru', TRUE, '4069', 'до 1 класса защищённости'),
('Dr.Web Enterprise Security Suite', 'ООО «Доктор Веб»', '13', 3, TRUE, '283', 'Корпоративное антивирусное решение', 'https://drweb.ru', TRUE, '3509', 'до 1 класса защищённости'),
('Dr.Web Desktop Security Suite', 'ООО «Доктор Веб»', '13', 3, TRUE, '284', 'Антивирус для рабочих станций', 'https://drweb.ru', TRUE, '3510', 'до 1 класса защищённости');

-- =====================================================
-- СРЕДСТВА ЗАЩИТЫ ИНФОРМАЦИИ (category_id = 4)
-- =====================================================
INSERT INTO software_catalog (name, vendor, version, category_id, is_russian, registry_number, description, website, fstec_certified, fstec_certificate_num, fstec_protection_class, fsb_certified, fsb_certificate_num, fsb_protection_class)
VALUES
('Secret Net Studio', 'ООО «Код Безопасности»', '8', 4, TRUE, '4012', 'СЗИ от НСД', 'https://securitycode.ru', TRUE, '3725', 'до 1 класса защищённости', FALSE, NULL, NULL),
('Secret Net LSP', 'ООО «Код Безопасности»', '1.10', 4, TRUE, '4013', 'СЗИ для Linux', 'https://securitycode.ru', TRUE, '3726', 'до 1 класса защищённости', FALSE, NULL, NULL),
('Dallas Lock', 'ООО «Конфидент»', '8', 4, TRUE, '3889', 'СЗИ от НСД', 'https://dallaslock.ru', TRUE, '2720', 'до 1 класса защищённости', FALSE, NULL, NULL),
('Соболь', 'ООО «Код Безопасности»', '4', 4, TRUE, '4014', 'Аппаратно-программный модуль доверенной загрузки', 'https://securitycode.ru', TRUE, '1967', 'до 1 класса защищённости', TRUE, 'СФ/124-3449', 'КС3'),
('VipNet Client', 'ОАО «ИнфоТеКС»', '4', 4, TRUE, '115', 'Клиент VPN и персональный МЭ', 'https://infotecs.ru', TRUE, '2706', 'до 3 класса защищённости', TRUE, 'СФ/124-2835', 'КС2'),
('VipNet Coordinator', 'ОАО «ИнфоТеКС»', '4', 4, TRUE, '116', 'VPN-шлюз и МЭ', 'https://infotecs.ru', TRUE, '2707', 'до 3 класса защищённости', TRUE, 'СФ/124-2836', 'КС3'),
('Континент', 'ООО «Код Безопасности»', '4', 4, TRUE, '4015', 'Комплекс сетевой безопасности', 'https://securitycode.ru', TRUE, '3727', 'до 3 класса защищённости', TRUE, 'СФ/525-3211', 'КС3'),
('КриптоПро CSP', 'ООО «КРИПТО-ПРО»', '5', 4, TRUE, '441', 'СКЗИ для формирования ЭП', 'https://cryptopro.ru', TRUE, NULL, NULL, TRUE, 'СФ/124-4061', 'КС1/КС2/КС3'),
('КриптоПро JCP', 'ООО «КРИПТО-ПРО»', '2', 4, TRUE, '442', 'СКЗИ для Java-приложений', 'https://cryptopro.ru', FALSE, NULL, NULL, TRUE, 'СФ/124-4062', 'КС1/КС2'),
('ViPNet SafeDisk', 'ОАО «ИнфоТеКС»', '4', 4, TRUE, '117', 'Шифрование дисков', 'https://infotecs.ru', TRUE, '2946', 'до 2 класса защищённости', TRUE, 'СФ/124-3001', 'КС2'),
('Аккорд-АМДЗ', 'ОКБ САПР', '5', 4, TRUE, '1851', 'АПМДЗ', 'https://okbsapr.ru', TRUE, '1025', 'до 1 класса защищённости', TRUE, 'СФ/027-1547', 'КС3'),
('Блокхост-Сеть', 'ООО «Газинформсервис»', '4', 4, TRUE, '6890', 'СЗИ от НСД', 'https://gaz-is.ru', TRUE, '3877', 'до 2 класса защищённости', FALSE, NULL, NULL);

-- =====================================================
-- ОФИСНЫЕ ПРИЛОЖЕНИЯ (category_id = 5)
-- =====================================================
INSERT INTO software_catalog (name, vendor, version, category_id, is_russian, registry_number, description, website)
VALUES
('МойОфис', 'ООО «Новые облачные технологии»', '2023', 5, TRUE, '305', 'Офисный пакет', 'https://myoffice.ru'),
('Р7-Офис', 'ООО «Р7»', '2023', 5, TRUE, '5667', 'Офисный пакет', 'https://r7-office.ru'),
('LibreOffice', 'The Document Foundation', '7', 5, FALSE, NULL, 'Свободный офисный пакет', 'https://libreoffice.org'),
('Microsoft Office', 'Microsoft', '365', 5, FALSE, NULL, 'Офисный пакет Microsoft', 'https://microsoft.com');

-- =====================================================
-- СИСТЕМЫ ВИРТУАЛИЗАЦИИ (category_id = 6)
-- =====================================================
INSERT INTO software_catalog (name, vendor, version, category_id, is_russian, registry_number, description, website, fstec_certified, fstec_certificate_num, fstec_protection_class)
VALUES
('zVirt', 'ООО «Орион софт»', '3', 6, TRUE, '7765', 'Платформа виртуализации', 'https://orionsoft.ru', TRUE, '4244', 'до 2 класса защищённости'),
('Скала-Р', 'ООО «Скала Софтвер»', '2', 6, TRUE, '8121', 'Гиперконвергентная платформа', 'https://scala-r.ru', TRUE, '4501', 'до 2 класса защищённости'),
('РОСА Виртуализация', 'ООО «НТЦ ИТ РОСА»', '3', 6, TRUE, '6221', 'Платформа виртуализации', 'https://rosa.ru', TRUE, '4012', 'до 3 класса защищённости'),
('VMware vSphere', 'VMware', '8', 6, FALSE, NULL, 'Платформа виртуализации', 'https://vmware.com', FALSE, NULL, NULL),
('Microsoft Hyper-V', 'Microsoft', '2022', 6, FALSE, NULL, 'Гипервизор Microsoft', 'https://microsoft.com', FALSE, NULL, NULL),
('Proxmox VE', 'Proxmox', '8', 6, FALSE, NULL, 'Платформа виртуализации', 'https://proxmox.com', FALSE, NULL, NULL);

-- =====================================================
-- СРЕДСТВА МОНИТОРИНГА И SIEM (category_id = 7)
-- =====================================================
INSERT INTO software_catalog (name, vendor, version, category_id, is_russian, registry_number, description, website, fstec_certified, fstec_certificate_num)
VALUES
('MaxPatrol SIEM', 'АО «Позитив Текнолоджиз»', '7', 7, TRUE, '3887', 'SIEM-система', 'https://ptsecurity.com', TRUE, '3734'),
('KOMRAD Enterprise SIEM', 'ООО «НПО «Эшелон»', '4', 7, TRUE, '5585', 'SIEM-система', 'https://npo-echelon.ru', TRUE, '4015'),
('RuSIEM', 'ООО «РуСИЕМ»', '3', 7, TRUE, '7501', 'SIEM-система', 'https://rusiem.com', TRUE, '4287'),
('Kaspersky KUMA', 'АО «Лаборатория Касперского»', '2', 7, TRUE, '8845', 'SIEM-платформа', 'https://kaspersky.ru', TRUE, '4401'),
('Zabbix', 'Zabbix LLC', '6', 7, FALSE, NULL, 'Система мониторинга', 'https://zabbix.com', FALSE, NULL),
('Grafana', 'Grafana Labs', '10', 7, FALSE, NULL, 'Платформа визуализации', 'https://grafana.com', FALSE, NULL);

-- =====================================================
-- СКАНЕРЫ УЯЗВИМОСТЕЙ (category_id = 8)
-- =====================================================
INSERT INTO software_catalog (name, vendor, version, category_id, is_russian, registry_number, description, website, fstec_certified, fstec_certificate_num)
VALUES
('MaxPatrol 8', 'АО «Позитив Текнолоджиз»', '8', 8, TRUE, '3888', 'Сканер уязвимостей', 'https://ptsecurity.com', TRUE, '3375'),
('XSpider', 'АО «Позитив Текнолоджиз»', '7.8', 8, TRUE, '3889', 'Сканер уязвимостей', 'https://ptsecurity.com', TRUE, '2530'),
('RedCheck', 'АО «АЛТЭКС-СОФТ»', '3', 8, TRUE, '4451', 'Сканер уязвимостей', 'https://altx-soft.ru', TRUE, '3456'),
('ScanOVAL', 'ФСТЭК России', '1', 8, TRUE, NULL, 'Бесплатный сканер уязвимостей ФСТЭК', 'https://fstec.ru', TRUE, '4098');

-- =====================================================
-- КОНТЕЙНЕРИЗАЦИЯ И ОРКЕСТРАЦИЯ (category_id = 9)
-- =====================================================
INSERT INTO software_catalog (name, vendor, version, category_id, is_russian, registry_number, description, website, fstec_certified)
VALUES
('Deckhouse', 'ООО «Флант»', '1.50', 9, TRUE, '10215', 'Kubernetes-платформа', 'https://deckhouse.io', FALSE),
('Podman', 'Red Hat', '4', 9, FALSE, NULL, 'Контейнерный движок', 'https://podman.io', FALSE),
('Docker', 'Docker Inc', '24', 9, FALSE, NULL, 'Контейнерная платформа', 'https://docker.com', FALSE),
('Kubernetes', 'CNCF', '1.28', 9, FALSE, NULL, 'Оркестрация контейнеров', 'https://kubernetes.io', FALSE);

-- =====================================================
-- СРЕДСТВА РЕЗЕРВНОГО КОПИРОВАНИЯ (category_id = 10)
-- =====================================================
INSERT INTO software_catalog (name, vendor, version, category_id, is_russian, registry_number, description, website, fstec_certified, fstec_certificate_num)
VALUES
('Кибер Бэкап', 'ООО «Киберпротект»', '16', 10, TRUE, '6850', 'Система резервного копирования', 'https://cyberprotect.ru', TRUE, '4380'),
('RuBackup', 'ООО «РуБэкап»', '2', 10, TRUE, '8932', 'Система резервного копирования', 'https://rubackup.ru', TRUE, '4412'),
('Handy Backup', 'ООО «Новософт»', '8', 10, TRUE, '1852', 'ПО резервного копирования', 'https://handybackup.ru', FALSE, NULL),
('Veeam Backup', 'Veeam', '12', 10, FALSE, NULL, 'Система резервного копирования', 'https://veeam.com', FALSE, NULL);

-- =====================================================
-- СРЕДСТВА АУТЕНТИФИКАЦИИ (category_id = 11)
-- =====================================================
INSERT INTO software_catalog (name, vendor, version, category_id, is_russian, registry_number, description, website, fstec_certified, fstec_certificate_num, fsb_certified, fsb_certificate_num)
VALUES
('JaCarta', 'АО «Аладдин Р.Д.»', '2', 11, TRUE, '1234', 'Средства аутентификации и ЭП', 'https://aladdin-rd.ru', TRUE, '3449', TRUE, 'СФ/124-3501'),
('Рутокен', 'АО «Актив-софт»', '3', 11, TRUE, '987', 'USB-токены и смарт-карты', 'https://rutoken.ru', TRUE, '3021', TRUE, 'СФ/124-2877'),
('Indeed AM', 'ООО «Индид»', '8', 11, TRUE, '7123', 'Платформа управления доступом', 'https://indeed-id.com', TRUE, '4156', FALSE, NULL),
('Avanpost FAM', 'ООО «Аванпост»', '5', 11, TRUE, '6891', 'Федеративная аутентификация', 'https://avanpost.ru', TRUE, '4089', FALSE, NULL);

-- =====================================================
-- DLP-СИСТЕМЫ (category_id = 12)
-- =====================================================
INSERT INTO software_catalog (name, vendor, version, category_id, is_russian, registry_number, description, website, fstec_certified, fstec_certificate_num)
VALUES
('InfoWatch Traffic Monitor', 'АО «ИнфоВотч»', '7', 12, TRUE, '1567', 'DLP-система', 'https://infowatch.ru', TRUE, '3890'),
('Zecurion DLP', 'АО «Зекурион»', '11', 12, TRUE, '2341', 'DLP-система', 'https://zecurion.ru', TRUE, '3567'),
('SearchInform', 'ООО «СёрчИнформ»', '6', 12, TRUE, '3456', 'DLP и SIEM система', 'https://searchinform.ru', TRUE, '3678'),
('Стахановец', 'ООО «Стахановец»', '9', 12, TRUE, '5678', 'Система контроля сотрудников', 'https://stakhanovets.ru', TRUE, '3789');
