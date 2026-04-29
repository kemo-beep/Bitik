-- +goose Up
-- +goose StatementBegin
CREATE TABLE user_sessions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  device_id VARCHAR(255),
  user_agent TEXT,
  ip_address INET,
  platform VARCHAR(50),
  push_token TEXT,
  last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  revoked_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_user_sessions_user_active ON user_sessions(user_id) WHERE revoked_at IS NULL;

CREATE TRIGGER trg_user_sessions_updated_at BEFORE UPDATE ON user_sessions FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE password_reset_tokens (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash TEXT NOT NULL UNIQUE,
  expires_at TIMESTAMPTZ NOT NULL,
  consumed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_password_reset_tokens_user ON password_reset_tokens(user_id);
CREATE INDEX idx_password_reset_tokens_expires ON password_reset_tokens(expires_at);

CREATE TABLE casbin_rule (
  id SERIAL PRIMARY KEY,
  ptype VARCHAR(100) NOT NULL,
  v0 VARCHAR(256),
  v1 VARCHAR(256),
  v2 VARCHAR(256),
  v3 VARCHAR(256),
  v4 VARCHAR(256),
  v5 VARCHAR(256)
);

CREATE INDEX idx_casbin_rule_ptype ON casbin_rule(ptype);

ALTER TABLE refresh_tokens
  ADD COLUMN session_id UUID REFERENCES user_sessions(id) ON DELETE SET NULL;

CREATE INDEX idx_refresh_tokens_session_id ON refresh_tokens(session_id);

-- Dev seed users: bcrypt hash for password "password" (cost 10).
UPDATE users
SET password_hash = '$2a$10$xbra1pjJj4jk3D1MLcDcuOyphbdWbzP8SNzLfScLQDbjse0X2yVfG'
WHERE email IN ('admin@bitik.local', 'seller@bitik.local', 'buyer@bitik.local');

INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES
  ('p', 'admin', '/api/v1/admin', '*'),
  ('p', 'admin', '/api/v1/admin/*', '*'),
  ('p', 'seller', '/api/v1/seller', '*'),
  ('p', 'seller', '/api/v1/seller/*', '*'),
  ('p', 'buyer', '/api/v1/buyer', '*'),
  ('p', 'buyer', '/api/v1/buyer/*', '*'),
  ('p', 'buyer', '/api/v1/users', '*'),
  ('p', 'buyer', '/api/v1/users/*', '*'),
  ('p', 'seller', '/api/v1/users', '*'),
  ('p', 'seller', '/api/v1/users/*', '*'),
  ('p', 'admin', '/api/v1/users', '*'),
  ('p', 'admin', '/api/v1/users/*', '*');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_refresh_tokens_session_id;
ALTER TABLE refresh_tokens DROP COLUMN IF EXISTS session_id;
DROP TABLE IF EXISTS casbin_rule;
DROP TABLE IF EXISTS password_reset_tokens;
DROP TRIGGER IF EXISTS trg_user_sessions_updated_at ON user_sessions;
DROP TABLE IF EXISTS user_sessions;
-- +goose StatementEnd
