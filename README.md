# Monorepo: Mobile + API + Infra

Structure:
- `apps/mobile` : Expo + TypeScript, 2 ecrans (Discover / Matches)
- `services/api` : API Go minimale avec `GET /health`
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

Verifier la sante de l'API:

```bash
curl http://localhost:8080/health
```

## 2) Lancer le mobile Expo
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
go run ./cmd/api
```
