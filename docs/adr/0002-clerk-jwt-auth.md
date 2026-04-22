# ADR 0002 — Verificación de JWT de Clerk en Go

## Estado
Aceptado

## Contexto
laboris-api necesita proteger endpoints que crean o leen datos de usuarios (requests, pedidos, perfil). El frontend usa Clerk como auth provider y envía JWTs en el header `Authorization: Bearer`. El backend debe verificar esos tokens sin depender de un servicio externo en cada request.

## Decisión
Usar **`github.com/clerk/clerk-sdk-go/v2`** para verificar los JWTs de Clerk en un middleware de Gin.

## Implementación
```
internal/middleware/auth.go
  → verifica el JWT con clerk-sdk-go
  → extrae el subject (userId de Clerk)
  → lo inyecta en el Gin context: c.Set("userId", subject)
  → retorna 401 si el token es inválido, expirado o ausente

internal/handler/router.go
  → grupo público: /ping, GET /professionals, GET /professionals/:id
  → grupo protegido (con middleware): POST /requests, GET /requests, etc.
```

## Alternativas consideradas

| Opción | Pros | Contras |
|--------|------|---------|
| **clerk-sdk-go** | SDK oficial, verificación local (sin round-trip), soporte de JWKS | Dependencia de Clerk |
| Verificación manual de JWT (RS256) | Sin dependencias externas | Implementar JWKS fetching, rotación de keys, más código de seguridad |
| Llamar a Clerk API por request | Simple | Latencia extra en cada request, punto de fallo externo |

## Consecuencias
- `CLERK_SECRET_KEY` se agrega a config y env vars de Render
- Los handlers protegidos leen `c.GetString("userId")` del context
- Si se migra de Clerk: cambiar solo el middleware de verificación
