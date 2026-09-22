DO $$
BEGIN
  IF to_regclass('public.app_users') IS NOT NULL AND to_regclass('public."user"') IS NULL THEN
    ALTER TABLE app_users RENAME TO "user";
  END IF;
END $$;

CREATE TABLE IF NOT EXISTS "user" (
  id text PRIMARY KEY,
  google_subject text UNIQUE NOT NULL DEFAULT '',
  name text NOT NULL DEFAULT '',
  email text UNIQUE NOT NULL,
  email_verified boolean NOT NULL DEFAULT false,
  avatar_url text NOT NULL DEFAULT '',
  password_hash text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS wallets (
  id text PRIMARY KEY,
  owner_id text NOT NULL,
  name text NOT NULL,
  type text NOT NULL,
  currency text NOT NULL DEFAULT 'VND',
  opening_balance bigint NOT NULL DEFAULT 0,
  is_in_total boolean NOT NULL DEFAULT true,
  description text,
  target_amount bigint,
  target_date timestamptz,
  credit_limit bigint,
  last_statement_balance bigint,
  statement_day integer,
  payment_due_day integer,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT wallets_type_check CHECK (type IN ('basic', 'goal', 'credit')),
  CONSTRAINT wallets_currency_check CHECK (currency = 'VND')
);
CREATE INDEX IF NOT EXISTS wallets_owner_id_idx ON wallets(owner_id);

CREATE TABLE IF NOT EXISTS categories (
  id text PRIMARY KEY,
  owner_id text,
  parent_id text,
  kind text NOT NULL,
  name text NOT NULL,
  system_key text UNIQUE,
  is_system boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT categories_kind_check CHECK (kind IN ('expense', 'income', 'debt'))
);
CREATE INDEX IF NOT EXISTS categories_owner_id_idx ON categories(owner_id);
CREATE INDEX IF NOT EXISTS categories_parent_id_idx ON categories(parent_id);

CREATE TABLE IF NOT EXISTS category_wallets (
  category_id text NOT NULL,
  wallet_id text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (category_id, wallet_id),
  FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE,
  FOREIGN KEY (wallet_id) REFERENCES wallets(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS transactions (
  id text PRIMARY KEY,
  owner_id text NOT NULL,
  wallet_id text NOT NULL,
  category_id text,
  type text NOT NULL,
  amount bigint NOT NULL CHECK (amount > 0),
  occurred_at timestamptz NOT NULL,
  note text,
  included_in_reports boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT transactions_type_check CHECK (type IN ('income', 'expense')),
  FOREIGN KEY (wallet_id) REFERENCES wallets(id) ON DELETE CASCADE,
  FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE RESTRICT
);
CREATE INDEX IF NOT EXISTS transactions_owner_id_idx ON transactions(owner_id);
CREATE INDEX IF NOT EXISTS transactions_wallet_id_idx ON transactions(wallet_id);
CREATE INDEX IF NOT EXISTS transactions_category_id_idx ON transactions(category_id);
CREATE INDEX IF NOT EXISTS transactions_occurred_at_idx ON transactions(occurred_at);

INSERT INTO categories (id, kind, name, system_key, is_system) VALUES
('00000000-0000-4000-8000-000000000301','expense','Ăn uống','expense_food',true),
('00000000-0000-4000-8000-000000000302','expense','Hoá đơn & Tiện ích','expense_bills',true),
('00000000-0000-4000-8000-000000000303','expense','Mua sắm','expense_shopping',true),
('00000000-0000-4000-8000-000000000304','expense','Gia đình','expense_family',true),
('00000000-0000-4000-8000-000000000305','expense','Di chuyển','expense_transport',true),
('00000000-0000-4000-8000-000000000306','expense','Sức khỏe','expense_health',true),
('00000000-0000-4000-8000-000000000307','expense','Giáo dục','expense_education',true),
('00000000-0000-4000-8000-000000000308','expense','Giải trí','expense_entertainment',true),
('00000000-0000-4000-8000-000000000309','expense','Quà tặng & Quyên góp','expense_gifts',true),
('00000000-0000-4000-8000-000000000310','expense','Bảo hiểm','expense_insurance',true),
('00000000-0000-4000-8000-000000000311','expense','Đầu tư','expense_investment',true),
('00000000-0000-4000-8000-000000000312','expense','Du lịch','expense_travel',true),
('00000000-0000-4000-8000-000000000313','expense','Tiết kiệm','expense_savings',true),
('00000000-0000-4000-8000-000000000314','expense','Kinh doanh','expense_business',true),
('00000000-0000-4000-8000-000000000315','expense','Các chi phí chung','expense_general',true),
('00000000-0000-4000-8000-000000000316','income','Thu nhập','income_root',true),
('00000000-0000-4000-8000-000000000317','debt','Vay/Nợ','debt_root',true)
ON CONFLICT (system_key) DO UPDATE SET name = EXCLUDED.name, kind = EXCLUDED.kind, is_system = true;

INSERT INTO categories (id, parent_id, kind, name, system_key, is_system)
SELECT child.id, parent.id, child.kind, child.name, child.system_key, true
FROM (VALUES
('00000000-0000-4000-8000-000000000321','expense_food','expense','Ăn vặt','expense_food_snacks'),
('00000000-0000-4000-8000-000000000322','expense_food','expense','Cà phê','expense_food_coffee'),
('00000000-0000-4000-8000-000000000323','expense_food','expense','Cơm bữa','expense_food_meals'),
('00000000-0000-4000-8000-000000000324','expense_food','expense','Nhà hàng','expense_food_restaurant'),
('00000000-0000-4000-8000-000000000366','income_root','income','Lương','income_salary'),
('00000000-0000-4000-8000-000000000367','income_root','income','Thu nhập khác','income_other'),
('00000000-0000-4000-8000-000000000372','income_root','income','Thưởng','income_bonus'),
('00000000-0000-4000-8000-000000000374','debt_root','debt','Cho vay','debt_lend'),
('00000000-0000-4000-8000-000000000375','debt_root','debt','Trả nợ','debt_repay'),
('00000000-0000-4000-8000-000000000376','debt_root','debt','Đi vay','debt_loan'),
('00000000-0000-4000-8000-000000000377','debt_root','debt','Thu nợ','debt_collect')
) AS child(id, parent_key, kind, name, system_key)
JOIN categories parent ON parent.system_key = child.parent_key
ON CONFLICT (system_key) DO UPDATE SET parent_id = EXCLUDED.parent_id, name = EXCLUDED.name, kind = EXCLUDED.kind, is_system = true;

ALTER TABLE transactions
  DROP CONSTRAINT IF EXISTS transactions_wallet_id_fkey;

ALTER TABLE transactions
  ADD CONSTRAINT transactions_wallet_id_fkey
  FOREIGN KEY (wallet_id) REFERENCES wallets(id) ON DELETE CASCADE;

-- Owner-approved default category catalog. System rows are read-only in v1.
WITH roots(system_key, kind, name) AS (
  VALUES
    ('expense_food','expense','Ăn uống'),
    ('expense_bills','expense','Hoá đơn & Tiện ích'),
    ('expense_shopping','expense','Mua sắm'),
    ('expense_family','expense','Gia đình'),
    ('expense_transport','expense','Di chuyển'),
    ('expense_health','expense','Sức khoẻ'),
    ('expense_entertainment','expense','Giải trí'),
    ('expense_education','expense','Giáo dục'),
    ('expense_gifts','expense','Quà tặng & Quyên góp'),
    ('expense_insurance','expense','Bảo hiểm'),
    ('expense_investment','expense','Đầu tư'),
    ('expense_general','expense','Các chi phí khác'),
    ('expense_transfer_out','expense','Tiền chuyển đi'),
    ('expense_interest_paid','expense','Trả lãi'),
    ('expense_uncategorized','expense','Khoản chi chưa phân loại'),
    ('expense_withdrawal','expense','Rút tiền'),
    ('expense_accessories','expense','Phụ kiện'), ('expense_books','expense','Sách'),
    ('expense_business','expense','Kinh doanh'), ('expense_charity','expense','Từ thiện'),
    ('expense_funeral','expense','Tang lễ'), ('expense_games','expense','Trò chơi'),
    ('expense_home_repair','expense','Sửa chữa nhà cửa'), ('expense_wedding','expense','Cưới hỏi'),
    ('expense_movies','expense','Phim ảnh'), ('expense_parking','expense','Gửi xe'),
    ('expense_personal_care','expense','Chăm sóc cá nhân'), ('expense_fuel','expense','Xăng dầu'),
    ('expense_medicine','expense','Thuốc'), ('expense_restaurant','expense','Nhà hàng'),
    ('expense_sports','expense','Thể thao'), ('expense_taxi','expense','Taxi'),
    ('expense_travel','expense','Du lịch'), ('expense_savings','expense','Tiết Kiệm'),
    ('expense_children','expense','Con cái'), ('expense_clothes','expense','Quần áo'),
    ('expense_medical_treatment','expense','Khám chữa bệnh'), ('expense_electronics','expense','Thiết bị điện tử'),
    ('expense_cost','expense','Chi phí'), ('expense_shoes','expense','Giày dép'),
    ('expense_friends','expense','Bạn bè & Người yêu'),
    ('income_salary','income','Lương'), ('income_other','income','Thu nhập khác'),
    ('income_transfer_in','income','Tiền chuyển đến'), ('income_interest','income','Thu lãi'),
    ('income_uncategorized','income','Khoản thu chưa phân loại'), ('income_gift','income','Được tặng'),
    ('income_bonus','income','Thưởng'), ('income_sell_items','income','Bán đồ'),
    ('debt_lend','debt','Cho vay'), ('debt_repay','debt','Trả nợ'),
    ('debt_loan','debt','Đi vay'), ('debt_collect','debt','Thu nợ')
)
INSERT INTO categories (id, kind, name, system_key, is_system)
SELECT '00000000-0000-4000-8000-' || lpad((500 + row_number() OVER ())::text, 12, '0'), kind, name, system_key, true FROM roots
ON CONFLICT (system_key) DO UPDATE SET name = EXCLUDED.name, kind = EXCLUDED.kind, parent_id = NULL, is_system = true;

WITH children(system_key, parent_key, kind, name) AS (
  VALUES
    ('expense_food_snacks','expense_food','expense','Ăn vặt'),
    ('expense_food_coffee','expense_food','expense','Cà phê'),
    ('expense_food_meals','expense_food','expense','Cơm Bữa'),
    ('expense_bills_phone','expense_bills','expense','Hoá đơn điện thoại'),
    ('expense_bills_water','expense_bills','expense','Hoá đơn nước'),
    ('expense_bills_electricity','expense_bills','expense','Hoá đơn điện'),
    ('expense_bills_gas','expense_bills','expense','Hoá đơn gas'),
    ('expense_bills_tv','expense_bills','expense','Hoá đơn TV'),
    ('expense_bills_internet','expense_bills','expense','Hoá đơn internet'),
    ('expense_bills_rent','expense_bills','expense','Thuê nhà'),
    ('expense_bills_other','expense_bills','expense','Hoá đơn tiện ích khác'),
    ('expense_shopping_personal','expense_shopping','expense','Đồ dùng cá nhân'),
    ('expense_shopping_household','expense_shopping','expense','Đồ gia dụng'),
    ('expense_shopping_beauty','expense_shopping','expense','Làm đẹp'),
    ('expense_family_home_decor','expense_family','expense','Sửa & trang trí nhà'),
    ('expense_family_services','expense_family','expense','Dịch vụ gia đình'),
    ('expense_family_pets','expense_family','expense','Vật nuôi'),
    ('expense_transport_maintenance','expense_transport','expense','Bảo dưỡng xe'),
    ('expense_health_checkup','expense_health','expense','Khám sức khoẻ'),
    ('expense_health_fitness','expense_health','expense','Thể dục thể thao'),
    ('expense_entertainment_online','expense_entertainment','expense','Dịch vụ trực tuyến'),
    ('expense_entertainment_fun','expense_entertainment','expense','Vui - chơi')
)
INSERT INTO categories (id, parent_id, kind, name, system_key, is_system)
SELECT '00000000-0000-4000-8000-' || lpad((600 + row_number() OVER ())::text, 12, '0'), parent.id, child.kind, child.name, child.system_key, true
FROM children child JOIN categories parent ON parent.system_key = child.parent_key
ON CONFLICT (system_key) DO UPDATE SET parent_id = EXCLUDED.parent_id, name = EXCLUDED.name, kind = EXCLUDED.kind, is_system = true;

-- Remove retired system rows after reparenting any personal child to root.
UPDATE categories SET parent_id = NULL WHERE is_system = false AND parent_id IN (SELECT id FROM categories WHERE is_system = true AND system_key IN ('expense_food_restaurant','income_root','debt_root'));
DELETE FROM categories WHERE is_system = true AND system_key IN ('expense_food_restaurant','income_root','debt_root');

ALTER TABLE categories ADD COLUMN IF NOT EXISTS icon_key text NOT NULL DEFAULT 'tag';

UPDATE categories
SET is_system = COALESCE(system_key IN (
  'expense_general','expense_transfer_out','expense_interest_paid','expense_uncategorized','expense_withdrawal',
  'income_other','income_transfer_in','income_interest','income_uncategorized','income_gift',
  'debt_lend','debt_repay','debt_loan','debt_collect'
), false);

UPDATE categories
SET icon_key = system_key
WHERE owner_id IS NULL AND system_key IS NOT NULL AND icon_key = 'tag';

UPDATE categories AS personal
SET icon_key = template.icon_key
FROM categories AS template
WHERE personal.owner_id IS NOT NULL
  AND personal.icon_key = 'tag'
  AND template.owner_id IS NULL
  AND template.is_system = false
  AND template.name = personal.name
  AND template.kind = personal.kind
  AND template.icon_key <> 'tag';

-- Canonical category-icon seed. Every default category carries its icon key
-- from deployment; no frontend-only fallback is required to populate the DB.
WITH icon_seed(system_key, icon_key) AS (
  SELECT system_key, system_key
  FROM categories
  WHERE owner_id IS NULL AND system_key IS NOT NULL
)
UPDATE categories AS category
SET icon_key = icon_seed.icon_key
FROM icon_seed
WHERE category.system_key = icon_seed.system_key;

CREATE TABLE budgets (
 id text PRIMARY KEY,
 owner_id text NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
 name text NOT NULL CHECK (length(trim(name)) > 0),
 limit_amount bigint NOT NULL CHECK (limit_amount > 0 AND limit_amount <= 9007199254740991),
 wallet_id text REFERENCES wallets(id) ON DELETE CASCADE,
 category_id text REFERENCES categories(id) ON DELETE CASCADE,
 start_at timestamptz NOT NULL,
 end_at timestamptz NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(),
 updated_at timestamptz NOT NULL DEFAULT now(),
 CHECK (end_at > start_at)
);
CREATE INDEX budgets_owner_period ON budgets(owner_id, start_at, end_at);

CREATE TABLE ai_entry_sessions (
 id text PRIMARY KEY,
 owner_id text NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
 request_token text NOT NULL DEFAULT '',
 processing_until timestamptz,
 error text NOT NULL DEFAULT '',
 created_at timestamptz NOT NULL DEFAULT now(),
 updated_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(id, owner_id)
);
CREATE INDEX ai_entry_sessions_owner_recent ON ai_entry_sessions(owner_id, updated_at DESC);
CREATE TABLE ai_entry_messages (
 id text PRIMARY KEY,
 session_id text NOT NULL REFERENCES ai_entry_sessions(id) ON DELETE CASCADE,
 role text NOT NULL CHECK(role IN ('user','assistant')),
 content text NOT NULL,
 source_text text NOT NULL DEFAULT '',
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX ai_entry_messages_session ON ai_entry_messages(session_id,created_at);
CREATE TABLE ai_entry_proposals (
 id text PRIMARY KEY,
 session_id text NOT NULL,
 owner_id text NOT NULL,
 version integer NOT NULL DEFAULT 1 CHECK(version>0),
 status text NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','approved','rejected')),
 draft jsonb NOT NULL,
 questions jsonb NOT NULL DEFAULT '[]',
 transaction_id text UNIQUE,
 created_at timestamptz NOT NULL DEFAULT now(),
 updated_at timestamptz NOT NULL DEFAULT now(),
 FOREIGN KEY(session_id, owner_id) REFERENCES ai_entry_sessions(id, owner_id) ON DELETE CASCADE,
 CHECK((status='approved') = (transaction_id IS NOT NULL))
);
-- transaction_id intentionally remains a receipt after the ledger row is deleted;
-- replaying approval must never resurrect a deleted transaction.
CREATE INDEX ai_entry_proposals_session ON ai_entry_proposals(session_id,created_at);
CREATE TABLE ai_entry_requests (
 session_id text NOT NULL REFERENCES ai_entry_sessions(id) ON DELETE CASCADE,
 request_id text NOT NULL,
 hash text NOT NULL,
 token text NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY(session_id,request_id)
);

CREATE TABLE transaction_attachments (
 id text PRIMARY KEY,
 owner_id text NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
 session_id text NOT NULL,
 object_key text NOT NULL UNIQUE,
 filename text NOT NULL,
 mime_type text NOT NULL CHECK (mime_type IN ('image/jpeg','image/png','application/pdf')),
 size_bytes bigint NOT NULL CHECK (size_bytes > 0 AND size_bytes <= 5242880),
 sha256 text NOT NULL,
 ocr_status text NOT NULL DEFAULT 'pending' CHECK (ocr_status IN ('pending','completed','failed','deleting','delete_failed','deleted')),
 ocr_text text NOT NULL DEFAULT '',
 delete_after timestamptz NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(),
 updated_at timestamptz NOT NULL DEFAULT now(),
 FOREIGN KEY (session_id, owner_id) REFERENCES ai_entry_sessions(id, owner_id) ON DELETE CASCADE,
 UNIQUE (id, owner_id)
);
CREATE INDEX transaction_attachments_cleanup ON transaction_attachments(delete_after) WHERE ocr_status IN ('pending','completed','failed','deleting','delete_failed');

CREATE TABLE transaction_attachment_links (
 transaction_id text NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
 attachment_id text NOT NULL REFERENCES transaction_attachments(id) ON DELETE CASCADE,
 created_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY(transaction_id, attachment_id)
);

ALTER TABLE transaction_attachments ADD COLUMN IF NOT EXISTS updated_at timestamptz NOT NULL DEFAULT now();
ALTER TABLE transaction_attachments DROP CONSTRAINT IF EXISTS transaction_attachments_ocr_status_check;
ALTER TABLE transaction_attachments ADD CONSTRAINT transaction_attachments_ocr_status_check CHECK (ocr_status IN ('pending','completed','failed','deleting','delete_failed','deleted'));
DROP INDEX IF EXISTS transaction_attachments_cleanup;
CREATE INDEX transaction_attachments_cleanup ON transaction_attachments(delete_after, updated_at) WHERE ocr_status IN ('pending','completed','failed','deleting','delete_failed');

ALTER TABLE "user"
  ADD COLUMN timezone text NOT NULL DEFAULT 'Asia/Ho_Chi_Minh',
  ADD COLUMN timezone_confirmed boolean NOT NULL DEFAULT false;

-- Existing accounts retain the legacy app's Vietnam-local calendar interpretation.
UPDATE "user" SET timezone_confirmed = true;

-- Target dates are calendar labels; the legacy endpoint stored them as UTC midnight.
ALTER TABLE wallets
  ALTER COLUMN target_date TYPE date
  USING CASE WHEN target_date IS NULL THEN NULL ELSE (target_date AT TIME ZONE 'UTC')::date END;

DROP INDEX IF EXISTS budgets_owner_period;
ALTER TABLE budgets ADD COLUMN start_date date;
ALTER TABLE budgets ADD COLUMN end_date date;
UPDATE budgets
SET start_date = (start_at AT TIME ZONE 'Asia/Ho_Chi_Minh')::date,
    end_date = ((end_at AT TIME ZONE 'Asia/Ho_Chi_Minh') - interval '1 microsecond')::date;
ALTER TABLE budgets ALTER COLUMN start_date SET NOT NULL;
ALTER TABLE budgets ALTER COLUMN end_date SET NOT NULL;
ALTER TABLE budgets DROP COLUMN start_at;
ALTER TABLE budgets DROP COLUMN end_at;
ALTER TABLE budgets ADD CONSTRAINT budgets_date_order CHECK (end_date >= start_date);
CREATE INDEX budgets_owner_period ON budgets(owner_id, start_date, end_date);

CREATE TABLE jars (
  id text PRIMARY KEY,
  owner_id text NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(owner_id, id)
);

CREATE TABLE jar_months (
  owner_id text NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
  month date NOT NULL CHECK (extract(day FROM month) = 1),
  PRIMARY KEY(owner_id, month)
);

CREATE TABLE jar_month_configs (
  owner_id text NOT NULL,
  month date NOT NULL,
  jar_id text NOT NULL,
  name text NOT NULL CHECK (length(trim(name)) > 0),
  allocation_mode text NOT NULL CHECK (allocation_mode IN ('none', 'fixed', 'percent')),
  allocation_amount bigint,
  allocation_percent_bps integer,
  active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY(owner_id, month, jar_id),
  FOREIGN KEY(owner_id, month) REFERENCES jar_months(owner_id, month) ON DELETE CASCADE,
  FOREIGN KEY(owner_id, jar_id) REFERENCES jars(owner_id, id) ON DELETE CASCADE,
  CHECK (
    (allocation_mode = 'none' AND allocation_amount IS NULL AND allocation_percent_bps IS NULL) OR
    (allocation_mode = 'fixed' AND allocation_amount IS NOT NULL AND allocation_amount > 0 AND allocation_percent_bps IS NULL) OR
    (allocation_mode = 'percent' AND allocation_amount IS NULL AND allocation_percent_bps IS NOT NULL AND allocation_percent_bps BETWEEN 0 AND 10000)
  )
);
CREATE INDEX jar_month_configs_owner_month_active ON jar_month_configs(owner_id, month, active);

ALTER TABLE transactions ADD COLUMN jar_id text;
ALTER TABLE transactions ADD CONSTRAINT transactions_owner_jar_fk
  FOREIGN KEY(owner_id, jar_id) REFERENCES jars(owner_id, id);
CREATE INDEX transactions_owner_jar_idx ON transactions(owner_id, jar_id);

CREATE TABLE month_notes (
  owner_id text NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
  month date NOT NULL CHECK (extract(day FROM month) = 1),
  note text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY(owner_id, month)
);

ALTER TABLE transactions ADD COLUMN IF NOT EXISTS transfer_id text;
CREATE INDEX IF NOT EXISTS transactions_transfer_id_idx ON transactions(transfer_id);

CREATE TABLE changelogs (
  id text PRIMARY KEY,
  version text NOT NULL UNIQUE,
  title text NOT NULL,
  description text NOT NULL,
  published_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE feedback (
  id text PRIMARY KEY,
  user_id text NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
  type text NOT NULL CHECK (type IN ('bug', 'feature', 'improvement')),
  title text NOT NULL,
  description text NOT NULL,
  status text NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'triaged', 'in_progress', 'fixed', 'rejected')),
  fixed_at timestamptz,
  changelog_id text REFERENCES changelogs(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX feedback_user_created_idx ON feedback(user_id, created_at DESC);
CREATE INDEX feedback_status_created_idx ON feedback(status, created_at ASC);
CREATE INDEX feedback_changelog_idx ON feedback(changelog_id);

CREATE TABLE advisor_conversations (
    id text PRIMARY KEY,
    owner_id text NOT NULL UNIQUE REFERENCES "user"(id) ON DELETE CASCADE,
    generation bigint NOT NULL DEFAULT 1 CHECK (generation > 0),
    next_message_seq bigint NOT NULL DEFAULT 1 CHECK (next_message_seq > 0),
    summary jsonb NOT NULL DEFAULT '{}'::jsonb,
    summary_through_seq bigint NOT NULL DEFAULT 0 CHECK (summary_through_seq >= 0),
    summary_version bigint NOT NULL DEFAULT 0 CHECK (summary_version >= 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (id, owner_id)
);

CREATE TABLE advisor_runs (
    id text PRIMARY KEY,
    owner_id text NOT NULL,
    conversation_id text NOT NULL,
    generation bigint NOT NULL CHECK (generation > 0),
    client_request_id text NOT NULL,
    payload_hash text NOT NULL,
    credential_kind text NOT NULL CHECK (credential_kind IN ('session', 'user_api_key')),
    credential_id text NOT NULL,
    credential_expires_at timestamptz,
    status text NOT NULL CHECK (status IN ('queued', 'running', 'completed', 'failed', 'cancelled', 'interrupted', 'purged')),
    lease_token text,
    lease_until timestamptz,
    last_event_seq bigint NOT NULL DEFAULT 0 CHECK (last_event_seq >= 0),
    error_code text,
    created_at timestamptz NOT NULL DEFAULT now(),
    finished_at timestamptz,
    FOREIGN KEY (conversation_id, owner_id) REFERENCES advisor_conversations(id, owner_id) ON DELETE CASCADE,
    UNIQUE (owner_id, client_request_id)
);

CREATE UNIQUE INDEX advisor_one_active_run ON advisor_runs(conversation_id)
    WHERE status IN ('queued', 'running');
CREATE INDEX advisor_run_lease ON advisor_runs(lease_until)
    WHERE status = 'running';

CREATE TABLE advisor_messages (
    id text PRIMARY KEY,
    conversation_id text NOT NULL REFERENCES advisor_conversations(id) ON DELETE CASCADE,
    generation bigint NOT NULL,
    seq bigint NOT NULL CHECK (seq > 0),
    run_id text NOT NULL REFERENCES advisor_runs(id) ON DELETE CASCADE,
    role text NOT NULL CHECK (role IN ('user', 'assistant')),
    parts_version integer NOT NULL DEFAULT 1 CHECK (parts_version = 1),
    parts jsonb NOT NULL CHECK (jsonb_typeof(parts) = 'array'),
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (conversation_id, seq),
    UNIQUE (run_id, role)
);
CREATE INDEX advisor_messages_history ON advisor_messages(conversation_id, seq DESC);

CREATE TABLE advisor_events (
    run_id text NOT NULL REFERENCES advisor_runs(id) ON DELETE CASCADE,
    seq bigint NOT NULL CHECK (seq > 0),
    type text NOT NULL,
    payload jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (run_id, seq)
);

CREATE TABLE advisor_fact_bundles (
    id text PRIMARY KEY,
    owner_id text NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    run_id text NOT NULL REFERENCES advisor_runs(id) ON DELETE CASCADE,
    scope jsonb NOT NULL,
    as_of timestamptz NOT NULL,
    facts jsonb NOT NULL CHECK (jsonb_typeof(facts) = 'array'),
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX advisor_fact_bundles_run ON advisor_fact_bundles(run_id, created_at DESC);

CREATE TABLE user_api_keys (
    id text PRIMARY KEY,
    owner_id text NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    lookup_id text NOT NULL UNIQUE,
    name text NOT NULL CHECK (char_length(name) BETWEEN 1 AND 80),
    secret_hash text NOT NULL UNIQUE,
    scopes jsonb NOT NULL DEFAULT '[]'::jsonb,
    expires_at timestamptz,
    revoked_at timestamptz,
    last_used_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX user_api_keys_owner_active ON user_api_keys(owner_id, created_at DESC) WHERE revoked_at IS NULL;
