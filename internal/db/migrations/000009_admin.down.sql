ALTER TABLE professionals DROP COLUMN IF EXISTS status;
DROP INDEX IF EXISTS idx_professionals_status;

ALTER TABLE user_roles DROP CONSTRAINT IF EXISTS user_roles_role_check;
ALTER TABLE user_roles ADD CONSTRAINT user_roles_role_check CHECK (role IN ('client', 'professional'));
