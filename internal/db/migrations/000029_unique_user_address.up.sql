-- Elimina duplicados (user_id, address) ya creados por la migración
-- automática no atómica de AddressUseCase.List (dos llamadas concurrentes a
-- listMyAddresses podían insertar cada una su propio "Casa") — conserva la
-- fila más vieja de cada grupo antes de poder agregar el constraint.
DELETE FROM addresses a
USING addresses b
WHERE a.user_id = b.user_id
  AND a.address = b.address
  AND (a.created_at, a.id) > (b.created_at, b.id);

ALTER TABLE addresses ADD CONSTRAINT addresses_user_address_unique UNIQUE (user_id, address);
