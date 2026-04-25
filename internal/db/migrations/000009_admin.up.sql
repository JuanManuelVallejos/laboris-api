ALTER TABLE user_roles DROP CONSTRAINT IF EXISTS user_roles_role_check;
ALTER TABLE user_roles ADD CONSTRAINT user_roles_role_check CHECK (role IN ('client', 'professional', 'admin'));

ALTER TABLE professionals ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'active'
  CHECK (status IN ('active', 'suspended'));

CREATE INDEX IF NOT EXISTS idx_professionals_status ON professionals(status);
