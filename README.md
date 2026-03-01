# Monorepo: Mobile + API + Infra

Structure:
- `apps/mobile` : Expo + TypeScript, 2 ecrans (Discover / Matches)
- `services/api` : API Go (Gin) avec health + auth JWT
- `infra/docker-compose.yml` : Postgres + API

## Prerequis
- Node.js 18+
- npm
- Go 1.22+
- Docker + Docker Compose

## 1) Lancer DB + API
Depuis la racine du repo:

```bash
docker compose -f infra/docker-compose.yml up --build
```

Assurez-vous de definir `JWT_SECRET` avant le lancement.

PowerShell:

```powershell
$env:JWT_SECRET='change-me-in-dev'
docker compose -f infra/docker-compose.yml up --build
```

Bash:

```bash
export JWT_SECRET='change-me-in-dev'
docker compose -f infra/docker-compose.yml up --build
```

Verifier la sante de l'API:

```bash
curl http://localhost:8080/health
```

## 2) Endpoints auth MVP
- `POST /auth/register`
- `POST /auth/login`
- `GET /me` (header `Authorization: Bearer <token>`)

## 3) Lancer le mobile Expo
Dans un autre terminal:

```bash
cd apps/mobile
npm install
npm run start
```

Puis ouvrir l'app via Expo Go / emulateur.

## Scripts utiles
API en local (sans Docker):

```bash
cd services/api
export DATABASE_URL='postgresql://app:app@localhost:5432/app?sslmode=disable'
export JWT_SECRET='change-me-in-dev'
go run ./cmd/api
```

PowerShell:

```powershell
cd services/api
$env:DATABASE_URL='postgresql://app:app@localhost:5432/app?sslmode=disable'
$env:JWT_SECRET='change-me-in-dev'
go run ./cmd/api
```
