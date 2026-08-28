-- Coordenadas congeladas del domicilio elegido, en el mismo espíritu que
-- address_snapshot: se cargan al crear la solicitud y no vuelven a tocarse,
-- aunque el domicilio guardado original se edite o se borre después. Se usan
-- solo para el círculo aproximado de ~300m que ve el profesional antes de
-- que la dirección exacta se revele.
ALTER TABLE requests ADD COLUMN IF NOT EXISTS address_lat DOUBLE PRECISION;
ALTER TABLE requests ADD COLUMN IF NOT EXISTS address_lng DOUBLE PRECISION;
