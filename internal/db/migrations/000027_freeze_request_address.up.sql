ALTER TABLE requests ADD COLUMN IF NOT EXISTS address_snapshot TEXT;

-- Backfill: las solicitudes que ya tenían un address_id conservan el texto
-- que tenía ese domicilio en este momento, para no perder el dato — de acá
-- en más, address_snapshot es la fuente de verdad y no vuelve a tocarse
-- aunque el domicilio original se edite o se borre más tarde.
UPDATE requests rq
SET address_snapshot = ad.address
FROM addresses ad
WHERE ad.id = rq.address_id AND rq.address_snapshot IS NULL;
