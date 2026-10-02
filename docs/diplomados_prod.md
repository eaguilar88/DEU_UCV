# Diplomados (ECP) con el backend Go

`diplomados/` (repo `ecp_ucv`) es el frontend Next.js del portal de cursos y diplomados. En producción **no**
usa el mock `json-server`/`db.json`: todas las peticiones van al backend Go de este repositorio.

## Cómo funciona

El navegador solo habla con `/api/*`. La ruta `diplomados/src/app/api/[...path]/route.ts` reenvía cada petición a
`INTERNAL_API_URL` (el backend Go), por lo que no hace falta configurar CORS. El renderizado en servidor (SSR)
usa la misma variable.

## Desarrollo local

### 1. Node.js
    nvm install 18.20.4
    nvm use 18.20.4

### 2. Backend Go (Terminal 1, desde la raíz de DEU_UCV)
    cp .env.example .env     # solo la primera vez
    make start-backend
    make migrate-up

El backend queda en http://localhost:8081.

### 3. Frontend (Terminal 2, dentro de `diplomados/`)
    npm install
    cp .env.local.example .env.local   # INTERNAL_API_URL=http://localhost:8081
    npm run dev

La plataforma queda en http://localhost:3000.

## Producción

El servicio `diplomados` está definido en `docker-compose.prod.yml`:

- Se construye desde `./diplomados` (su `Dockerfile`) y depende de `backend`.
- `INTERNAL_API_URL=http://backend:8081` (red interna de Docker).
- Traefik lo publica en `formacion.extension.ucv.ve` con HTTPS (Let's Encrypt).

Comandos (desde la raíz de DEU_UCV):

    make start-prod
    make migrate-up-prod

## Ver también

- [Análisis de brechas: backend Go vs. mock Node](gap_analysis_diplomados_node_mock.md)
- [Guía de docker-compose](docker-compose-guia.md)
