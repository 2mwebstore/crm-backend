-- ─────────────────────────────────────────────────────────────────────────────
-- CRM Database Seed  (GORM AutoMigrate handles table creation)
-- ─────────────────────────────────────────────────────────────────────────────
USE `crm_db`;

-- ── Default admin ─────────────────────────────────────────────────────────────
-- Password: Admin@1234
INSERT IGNORE INTO users (name, email, password_hash, role, is_active, created_at, updated_at)
VALUES ('Super Admin','admin@crm.local',
  '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
  'admin',1,NOW(),NOW());

-- ── Levels ────────────────────────────────────────────────────────────────────
INSERT IGNORE INTO levels (name,description,color,sort_order,is_active,created_at,updated_at) VALUES
  ('Bronze',  'Entry-level clients',         '#cd7f32',1,1,NOW(),NOW()),
  ('Silver',  'Mid-tier clients',            '#c0c0c0',2,1,NOW(),NOW()),
  ('Gold',    'High-value clients',          '#ffd700',3,1,NOW(),NOW()),
  ('Platinum','Top-tier strategic accounts', '#e5e4e2',4,1,NOW(),NOW());

-- ── Contact Sources ───────────────────────────────────────────────────────────
INSERT IGNORE INTO contact_sources (name,description,icon,is_active,created_at,updated_at) VALUES
  ('Referral',      'Referred by existing client or partner','people',     1,NOW(),NOW()),
  ('Cold Call',     'Outbound cold call',                    'phone',      1,NOW(),NOW()),
  ('Website',       'Inbound via website or chat',           'globe',      1,NOW(),NOW()),
  ('Social Media',  'LinkedIn, Facebook, Instagram, etc.',   'share',      1,NOW(),NOW()),
  ('Event',         'Conference, trade show, networking',    'calendar',   1,NOW(),NOW()),
  ('Email Campaign','Marketing email campaign',              'mail',       1,NOW(),NOW()),
  ('Walk-in',       'Physical walk-in to office',            'door-open',  1,NOW(),NOW()),
  ('Partner',       'Referred by a business partner',        'handshake',  1,NOW(),NOW());

-- ── Bank Types ────────────────────────────────────────────────────────────────
INSERT IGNORE INTO bank_types (name,code,description,is_active,sort_order,created_at,updated_at) VALUES
  ('ABA Bank',            'ABA',     'Advanced Bank of Asia',               1,1,NOW(),NOW()),
  ('ACLEDA Bank',         'ACLEDA',  'ACLEDA Bank Plc.',                    1,2,NOW(),NOW()),
  ('Canadia Bank',        'CANADIA', 'Canadia Bank Plc.',                   1,3,NOW(),NOW()),
  ('Wing Bank',           'WING',    'Wing Bank (Cambodia) Plc.',           1,4,NOW(),NOW()),
  ('Prince Bank',         'PRINCE',  'Prince Bank Plc.',                    1,5,NOW(),NOW()),
  ('Cambodia Post Bank',  'CPB',     'Cambodia Post Bank Plc.',             1,6,NOW(),NOW()),
  ('Phillip Bank',        'PHILLIP', 'Phillip Bank Plc.',                   1,7,NOW(),NOW()),
  ('Other',               'OTHER',   'Other / unlisted bank',               1,99,NOW(),NOW());

-- ── Product Types ─────────────────────────────────────────────────────────────
INSERT IGNORE INTO product_types (name,code,description,icon,is_active,sort_order,created_at,updated_at) VALUES
  ('Personal Loan',   'PERSONAL_LOAN',   'Personal consumer loan',           'cash',        1,1,NOW(),NOW()),
  ('Business Loan',   'BUSINESS_LOAN',   'SME or corporate loan',            'briefcase',   1,2,NOW(),NOW()),
  ('Mortgage',        'MORTGAGE',        'Home or property mortgage',        'home',        1,3,NOW(),NOW()),
  ('Savings Account', 'SAVINGS',         'Savings or deposit account',       'piggy-bank',  1,4,NOW(),NOW()),
  ('Fixed Deposit',   'FIXED_DEPOSIT',   'Time deposit / fixed deposit',     'lock',        1,5,NOW(),NOW()),
  ('Insurance',       'INSURANCE',       'Life or general insurance',        'shield',      1,6,NOW(),NOW()),
  ('Investment',      'INVESTMENT',      'Investment or fund product',       'trending-up', 1,7,NOW(),NOW()),
  ('Credit Card',     'CREDIT_CARD',     'Credit card product',              'credit-card', 1,8,NOW(),NOW()),
  ('Remittance',      'REMITTANCE',      'Money transfer / remittance',      'send',        1,9,NOW(),NOW()),
  ('Other',           'OTHER',           'Other product or service',         'more',        1,99,NOW(),NOW());

-- ── Bonus Option Types ────────────────────────────────────────────────────────
INSERT IGNORE INTO bonus_option_types (name,code,description,calc_type,bonus_value,is_active,sort_order,created_at,updated_at) VALUES
  ('No Bonus',          'NO_BONUS',     'No bonus applied',                   'fixed',      0,     1,1,NOW(),NOW()),
  ('Flat $500',         'FLAT_500',     'Fixed $500 bonus',                   'fixed',      500,   1,2,NOW(),NOW()),
  ('Flat $1,000',       'FLAT_1000',    'Fixed $1,000 bonus',                 'fixed',      1000,  1,3,NOW(),NOW()),
  ('Flat $5,000',       'FLAT_5000',    'Fixed $5,000 bonus',                 'fixed',      5000,  1,4,NOW(),NOW()),
  ('5% Commission',     'PCT_5',        '5% of base value added as bonus',    'percentage', 5,     1,5,NOW(),NOW()),
  ('10% Commission',    'PCT_10',       '10% of base value added as bonus',   'percentage', 10,    1,6,NOW(),NOW()),
  ('15% Commission',    'PCT_15',       '15% of base value added as bonus',   'percentage', 15,    1,7,NOW(),NOW()),
  ('20% Commission',    'PCT_20',       '20% of base value added as bonus',   'percentage', 20,    1,8,NOW(),NOW());

-- ── Currency Types ────────────────────────────────────────────────────────────
INSERT IGNORE INTO currency_types (code,name,symbol,is_base,is_active,sort_order,created_at,updated_at) VALUES
  ('USD','US Dollar',       '$', 1,1,1,NOW(),NOW()),
  ('KHR','Cambodian Riel',  '៛',0,1,2,NOW(),NOW());

-- ── Exchange Rate (initial seed — update daily via API) ───────────────────────
INSERT IGNORE INTO exchange_rates (rate_date,usd_to_khr,khr_to_usd,note,created_by_id,created_at,updated_at)
VALUES (CURDATE(), 4100.0000, 0.0002439024, 'Initial seed rate', 1, NOW(), NOW());

-- ── Company Banks (seeded examples) ──────────────────────────────────────────
INSERT IGNORE INTO company_banks (bank_type_id, account_no, account_name, description, is_active, sort_order, created_at, updated_at)
SELECT bt.id, 'COMP-001-ABA', 'Company Main Account (ABA)', 'Primary operating account', 1, 1, NOW(), NOW()
FROM bank_types bt WHERE bt.code = 'ABA' LIMIT 1;

INSERT IGNORE INTO company_banks (bank_type_id, account_no, account_name, description, is_active, sort_order, created_at, updated_at)
SELECT bt.id, 'COMP-001-ACLEDA', 'Company Account (ACLEDA)', 'Secondary account', 1, 2, NOW(), NOW()
FROM bank_types bt WHERE bt.code = 'ACLEDA' LIMIT 1;

-- ── Sample deposit (requires client + product to exist) ───────────────────────
-- Uncomment after creating your first client and product via the API.
-- INSERT INTO deposits (transaction_no, date, client_id, client_product_id, client_bank_id, company_bank_id, amount, bal, to, os, play, currency, remark, created_by_id, created_at, updated_at)
-- VALUES ('DEP-20250627-DEMO', NOW(), 1, 1, 1, 1, 1000.00, 1000.00, 0, 1000.00, 0, 'USD', 'Demo deposit', 1, NOW(), NOW());
