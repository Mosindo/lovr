# Monorepo: Mobile + API + Infra

Structure:
- `apps/mobile` : Expo + TypeScript, 2 ecrans (Discover / Matches)
- `services/api` : API Go (Gin) avec health + auth JWT
- `infra/docker-compose.yml` : Postgres + API

## Etat Actuel (2026-03-02)
- Backend MVP fonctionnel: auth, discover, likes/matches, block atomique, chat.
- API backend modularisee par domaines:
  - `internal/auth` (JWT + middleware)
  - `internal/services` (logique metier)
  - `internal/handlers` (couche HTTP)
- `cmd/api/main.go` conserve un role de composition (wiring + routes).
- Configuration projet standardisee sur `.env` (pas de `.env.example`).

## Roadmap Immediate
- Mobile: polling conversation (ou websocket) pour chat quasi temps reel.
- Qualite: ajouter tests unitaires services backend + e2e mobile parcours critiques.
- Backend: introduire progressivement une couche repository (SQL explicite conserve).

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
curl http://localhost:18080/health
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

Smoke e2e mobile (parcours critique via API client mobile):

```bash
cd apps/mobile
npm run e2e:smoke
```

Variable optionnelle:
- `MOBILE_E2E_API_URL` (sinon fallback `EXPO_PUBLIC_API_URL`, puis `http://localhost:18080`)

QA lite (sans Maestro/adb):

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\qa-lite.ps1 -ApiBaseUrl http://localhost:18080
```

Ce pipeline execute:
- `gofmt` backend
- `go test ./...`
- `go build ./cmd/api`
- `npx tsc --noEmit` mobile
- `smoke-all` (API + mobile critique)

Rapport genere:
- `.smoke/qa-lite-report.txt`

## CI QA Lite (GitHub Actions)
Workflow:
- `.github/workflows/qa-lite.yml`

Declencheurs:
- `push` sur `main`
- `pull_request` vers `main`

Configuration GitHub requise:
- Aucune variable/secrets obligatoire pour ce workflow.
- Verifier que GitHub Actions est active sur le repository.
- Si votre branche par defaut n'est pas `main`, adapter `qa-lite.yml`.

Duree attendue:
- Typique: `5-12 minutes`
- Au-dela de `15 minutes`: verifier les logs (etapes `npm ci`, `docker compose up --build`, `Wait API health`)

Setup e2e UI device (fixtures match pre-creees):

```bash
cd apps/mobile
npm run e2e:ui:setup
```

Flow Maestro (si `maestro` est installe localement):

PowerShell:

```powershell
cd apps/mobile
npm run e2e:ui:run
```

Parametre optionnel du runner:
- `-AppId` (ex: `host.exp.exponent` pour Expo Go)
