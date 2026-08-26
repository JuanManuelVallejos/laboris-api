-- El dato de zona no se recupera al revertir (se perdió al aplicar el up).
ALTER TABLE professionals ADD COLUMN IF NOT EXISTS zone TEXT NOT NULL DEFAULT '';
ALTER TABLE professionals DROP COLUMN IF EXISTS home_address;
ALTER TABLE professionals DROP COLUMN IF EXISTS home_lat;
ALTER TABLE professionals DROP COLUMN IF EXISTS home_lng;
ALTER TABLE professionals DROP COLUMN IF EXISTS radius_km;

ALTER TABLE users DROP COLUMN IF EXISTS home_address;
ALTER TABLE users DROP COLUMN IF EXISTS home_lat;
ALTER TABLE users DROP COLUMN IF EXISTS home_lng;
