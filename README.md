# SSO API

API escrita en Go con Postgres. Incluye recetas para correr en local (Go) o en contenedores Docker.

## Requisitos
- Docker + Docker Compose
- (Opcional) Go 1.25 si deseas ejecutarlo sin contenedores

## Variables de entorno
- `.env` (local con Go): `DB_URL`, `PORT`, `JWT_SECRET`.
- `.env.docker`: para la API en Docker. Se carga porque `make docker-run` ejecuta `docker compose --env-file .env.docker ...`, así Compose toma esas vars sin editar el YAML. Por defecto apunta a `host.docker.internal:5432`.

## Ejecutar con Docker
### Usando tu Postgres ya existente (5432 en el host)
1. Ajusta credenciales en `.env.docker` si es necesario (por defecto usa `host.docker.internal:5432`).
2. Levanta la API en segundo plano (usa `.env.docker` porque el comando incluye `--env-file .env.docker`):
   ```bash
   make docker-run
   # o: docker compose --env-file .env.docker up -d --build api
   ```
3. API disponible en http://localhost:8080.

### Detener contenedores
```bash
make docker-down
# o: docker compose down
```

## Ejecutar con Go (sin Docker)
1. Crea/ajusta `.env` con `DB_URL`, `PORT`, `JWT_SECRET`.
2. Corre la app:
   ```bash
   make run
   ```

## Comandos útiles
- `make test` — ejecuta tests.
- `make sqlc` — regenera código sqlc (si aplicable).
- `make docker-run` — build + levanta la API en Docker (requiere Postgres externo).
- `make docker-down` — detiene contenedores.

## Recordatorio rápido (Docker)
- Configura `.env.docker` con tus credenciales/host de Postgres.
- Ejecuta `make docker-run` (incluye `--env-file .env.docker`).
- API en `http://localhost:8080`.

## Notas sobre la imagen Docker
- Usa build multi-stage con binario estático comprimido con UPX y stage final `scratch`, tamaño aprox. 12 MB.
- No incluye Postgres; se conecta a una instancia externa vía `DB_URL`.
