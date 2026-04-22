# ADR 0003 — PostgreSQL en Supabase

## Estado
Aceptado

## Contexto
laboris-api necesita persistencia real para usuarios, roles, profesionales y reviews. Hasta ahora todo era in-memory. El equipo es de 3 personas en etapa de MVP.

## Decisión
Usar **PostgreSQL hosteado en Supabase** como base de datos principal, con **`golang-migrate`** para gestionar el schema y **`pgx/v5`** como driver.

## Alternativas consideradas

| Opción | Pros | Contras |
|--------|------|---------|
| **Supabase** | Free tier real (500MB, sin vencimiento), panel SQL, mismo provider que auth keys | Vendor adicional |
| Render PostgreSQL | Todo en un mismo proveedor | Free tier vence a los 90 días |
| PlanetScale (MySQL) | Serverless, branching | MySQL en vez de Postgres, cambio de driver |

## Estructura de migrations

```
internal/db/migrations/
  000001_create_users.up/down.sql
  000002_create_user_roles.up/down.sql
  000003_create_professionals.up/down.sql
  000004_create_reviews.up/down.sql
```

Migrations embebidas con `//go:embed` en `internal/db/db.go` — corren automáticamente al iniciar el servidor.

## Decisiones de diseño

- **Rating calculado**: no se almacena en `professionals.rating`, se calcula con `AVG(reviews.rating)` en cada query. Simple y siempre consistente para el MVP.
- **Multi-rol**: `user_roles` es una tabla separada con `PRIMARY KEY (user_id, role)` — permite que un usuario sea cliente y profesional sin cambiar el modelo.
- **Fallback in-memory**: si `DATABASE_URL` no está seteado, el servidor arranca con el repositorio en memoria (útil para desarrollo sin DB).

## Variables de entorno
```
DATABASE_URL=postgresql://postgres:[password]@db.[project-id].supabase.co:5432/postgres
```
→ Supabase Dashboard → Settings → Database → Connection string → URI
