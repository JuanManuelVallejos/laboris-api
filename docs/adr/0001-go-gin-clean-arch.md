# ADR 0001 — Go + Gin + Clean Architecture

## Estado
Aceptado

## Contexto
laboris-api es la API REST del marketplace de servicios del hogar Laboris. Equipo de 3 personas, etapa de MVP. Se necesita un backend stateless, fácil de deployar en Render con Docker, y con una estructura que escale ordenada cuando se agreguen más entidades (Request, Job, Message, User).

## Decisión
Usar **Go** con **Gin** como framework HTTP, organizado en **Clean Architecture** con capas explícitas.

## Estructura de capas
```
cmd/api/          → entry point, wiring (DI manual)
config/           → carga de env vars
internal/
  domain/         → entidades + interfaces de repositorio
  usecase/        → lógica de negocio, orquesta repositorios
  handler/        → HTTP handlers (Gin), binding de request/response
  repository/     → implementaciones concretas (memory, postgres, etc.)
  middleware/     → auth, logging, recovery
```

## Alternativas consideradas

| Opción | Pros | Contras |
|--------|------|---------|
| **Go + Gin** | Rápido, tipado, bajo footprint, Docker liviano | Más verboso que frameworks opinados |
| Node.js + Express | Equipo familiarizado, ecosistema grande | Sin tipado fuerte, perf menor |
| Python + FastAPI | Rápido de prototipar | Más lento en runtime, menos adecuado para API pura |

## Consecuencias
- Cada nueva entidad sigue el mismo patrón: domain → usecase → handler → repository
- El DI es manual en `main.go` (sin framework de DI)
- Los repositorios son intercambiables: hoy in-memory, próximamente PostgreSQL
- Deploy: multi-stage Dockerfile en Render con `GIN_MODE=release`
