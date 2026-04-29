-- +goose Up
-- +goose StatementBegin
CREATE TABLE moderation_reports (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  reporter_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  target_type VARCHAR(80) NOT NULL,
  target_id UUID NOT NULL,
  reason TEXT NOT NULL,
  status moderation_status NOT NULL DEFAULT 'open',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_moderation_reports_target ON moderation_reports(target_type, target_id);

CREATE INDEX idx_moderation_reports_status ON moderation_reports(status, created_at);

CREATE TABLE moderation_cases (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  report_id UUID REFERENCES moderation_reports(id) ON DELETE SET NULL,
  assigned_to UUID REFERENCES users(id) ON DELETE SET NULL,
  status moderation_status NOT NULL DEFAULT 'open',
  resolution TEXT,
  resolved_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_moderation_cases_status ON moderation_cases(status, created_at);

CREATE TRIGGER trg_moderation_cases_updated_at BEFORE UPDATE ON moderation_cases FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE audit_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  actor_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  action VARCHAR(150) NOT NULL,
  entity_type VARCHAR(100),
  entity_id UUID,
  old_values JSONB,
  new_values JSONB,
  ip_address INET,
  user_agent TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_logs_actor ON audit_logs(actor_user_id);

CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id);

CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);

CREATE TABLE idempotency_keys (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  key VARCHAR(255) NOT NULL,
  scope_hash TEXT NOT NULL,
  user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  request_hash TEXT,
  response_body JSONB,
  status_code INT,
  locked_until TIMESTAMPTZ,
  expires_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (scope_hash, key)
);

CREATE INDEX idx_idempotency_keys_expires ON idempotency_keys(expires_at);

CREATE TABLE event_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  event_name VARCHAR(150) NOT NULL,
  entity_type VARCHAR(100),
  entity_id UUID,
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  ip_address INET,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_event_logs_user ON event_logs(user_id);

CREATE INDEX idx_event_logs_event_name ON event_logs(event_name);

CREATE INDEX idx_event_logs_created_at ON event_logs(created_at);

CREATE TABLE search_recent_queries (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  session_id VARCHAR(255),
  query TEXT NOT NULL,
  filters JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_search_recent_user ON search_recent_queries(user_id, created_at DESC);

CREATE INDEX idx_search_recent_session ON search_recent_queries(session_id, created_at DESC);

CREATE TABLE search_click_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  session_id VARCHAR(255),
  query TEXT NOT NULL,
  product_id UUID REFERENCES products(id) ON DELETE SET NULL,
  position INT CHECK (position IS NULL OR position > 0),
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_search_click_events_product ON search_click_events(product_id, created_at);

CREATE INDEX idx_search_click_events_query ON search_click_events(query, created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS search_click_events;
DROP TABLE IF EXISTS search_recent_queries;
DROP TABLE IF EXISTS event_logs;
DROP TABLE IF EXISTS idempotency_keys;
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS moderation_cases;
DROP TABLE IF EXISTS moderation_reports;
-- +goose StatementEnd
