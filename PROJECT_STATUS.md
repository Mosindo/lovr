# Lovr - Project Status

## 1. Vision Produit
Lovr est une application de rencontre "human-first" orientee confiance et securite: decouverte simple, match reciproque, chat uniquement apres match, et blocage atomique.

## 2. Stack Technique
- Mobile: React Native (Expo), TypeScript
- API: Go, Gin
- DB: PostgreSQL
- Auth: JWT HS256
- Infra locale: Docker Compose (API + Postgres)

## 3. Avancement Global (estimation)

| Domaine | Avancement | Notes |
|---|---:|---|
| Produit MVP (core flows) | 70% | Flows essentiels disponibles, hardening restant |
| Backend API | 92% | Flows MVP presents + modularisation Auth/Social/Chat + couche repositories + logs structures renforces |
| Base de donnees & migrations | 72% | Schema MVP en place, modeles futurs a consolider |
| Mobile App | 76% | Auth + discover + matches + block + chat UI + polling conversation/listing + retry/error states chat renforces |
| Tests & qualite | 86% | Tests integration API + services + smoke e2e mobile + base e2e UI Maestro |
| DevOps / exploitation locale | 70% | Docker local stable, checklist runbook a renforcer |

### Smoke tests MVP (checklist courte)

- [x] Smoke test API automatisé (auth → discover → like/match → chat → block)
- [x] Smoke test mobile e2e (parcours critique complet)
- [x] Script orchestration "smoke_all"
- [x] Rapport automatique PASS/FAIL généré

## 4. Backend

### Checklist Fonctionnelle
- [x] `GET /health`
- [x] `POST /auth/register`
- [x] `POST /auth/login`
- [x] `GET /me` (JWT requis)
- [x] `GET /discover` (exclusion deja likes / blocks)
- [x] `POST /likes` (match reciproque)
- [x] `GET /matches`
- [x] `POST /block` (atomique)
- [x] `GET /chats`
- [x] `GET /chats/:userId/messages`
- [x] `POST /chats/:userId/messages`

### Checklist Technique
- [x] Validation JWT stricte HS256/HMAC
- [x] Password hashing bcrypt
- [x] Pool DB unique (pgxpool)
- [x] Migrations executees au demarrage
- [x] Verification interaction bloquee avant like/chat
- [x] Tests integration backend (auth/social/chat)
- [x] Tests integration couche services (Auth/Social/Chat)
- [x] Separation handlers/services effective (Auth/Social/Chat extraits de `main.go`)
- [x] Couche repository dediee (SQL extrait des services vers `internal/repositories`)
- [x] Observabilite (logs structures, metriques)
- [x] Pagination/limites robustes pour discover/chat

## 5. Base de Donnees

### Tables Prevues
- `users`
- `likes`
- `blocks`
- `messages`
- (future) `profiles`, `interests`, `user_interests`, `match_events`, `reports`

### Etat des Migrations
- [x] `001_create_users.sql`
- [x] `002_create_likes.sql`
- [x] `003_create_blocks.sql`
- [x] `004_create_messages.sql`

Etat: migrations MVP appliquees et coherentes avec les endpoints actuels.

## 6. Mobile

### Etat Actuel
- [x] Auth UI (register/login) + persistance token
- [x] Tabs: Discover / Matches / Chats / Account
- [x] Discover: load profils, like, block
- [x] Matches: listing + block
- [x] Chats: liste conversations, lecture, envoi
- [x] Gestion erreurs reseau de base (timeout + message)

### Taches Prevues
- [x] Polling auto des messages (mode polling intervalle sur ecran chat)
- [x] UX loading/error states plus robustes
- [ ] Ecrans profil + edition
- [ ] Validation UX mobile multi-device (Android/iOS)
- [x] Smoke e2e mobile parcours critique (script `npm run e2e:smoke`)
- [x] Base tests UI/e2e device (Maestro flow + setup fixtures + testIDs)
- [ ] Execution CI/device des tests UI/e2e (Detox/Maestro)

## 7. Problemes Techniques Ouverts
- Conflits possibles de port local (ex: `8080` pris par Apache/IIS/WSL)
- Configuration reseau mobile physique dependante de `EXPO_PUBLIC_API_URL`
- Couverture tests forte backend, faible mobile
- Instabilite d'execution e2e UI locale (Maestro/adb/HOME/permissions/sessions longues)
- Outillage e2e UI device dependant de prerequis externes (PATH, Java, CLI) selon environnement local

## 8. Prochaine Priorite
- Validation UX mobile multi-device (Android/iOS)
- Execution CI/device des tests UI/e2e (Detox/Maestro)
- Maintien d'une gate `QA lite` stable sans outillage device pour limiter les blocages de delivery

### Phase 1 – Bonus (Hardening technique)
- Execution reelle des tests e2e UI sur device/emulateur (Maestro/Detox)
- Trajectoire recommandee: migration progressive UI e2e vers Detox (Maestro conserve en fallback)

## 9. Risques Techniques
- Regression fonctionnelle lors du futur decoupage backend
- Incoherences UX mobile sous reseau instable
- Dette de structure si la croissance features continue sans modularisation
- Absence de monitoring/alerting en environnement non-local

## 10. Decisions Techniques Prises
- Gin unique (pas de melange `net/http`)
- JWT HS256 + claims stricts
- bcrypt obligatoire
- SQL explicite sans ORM
- Migrations SQL embarquees dans le binaire API
- Blocage atomique avec suppression des interactions bidirectionnelles
- Chat autorise uniquement apres match reciproque et hors blocage
- Port host API dedie (`18080`) pour eviter conflits locaux
- Modularisation backend par domaines (`auth`, `social`, `chat`) via handlers/services
- Standardisation projet: utilisation de `.env` uniquement (suppression `.env.example`)
- Couche repositories Postgres dediee (`internal/repositories`) pour sortir le SQL des services
- Gate qualite MVP: prioriser smoke API/mobile automatises; UI e2e device en hardening technique

## 11. Date de Derniere Mise a Jour
2026-03-10
