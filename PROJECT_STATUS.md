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
| Backend API | 78% | Auth, discover, likes/matches, block, chat presentes |
| Base de donnees & migrations | 72% | Schema MVP en place, modeles futurs a consolider |
| Mobile App | 68% | Auth + discover + matches + block + chat UI disponibles |
| Tests & qualite | 65% | Tests integration backend en place, e2e mobile a ajouter |
| DevOps / exploitation locale | 70% | Docker local stable, checklist runbook a renforcer |

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
- [ ] Separation handlers/services/repositories (main.go encore volumineux)
- [ ] Observabilite (logs structures, metriques)
- [ ] Pagination/limites robustes pour discover/chat

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
- [ ] Polling auto des messages (ou websocket)
- [ ] UX loading/error states plus robustes
- [ ] Ecrans profil + edition
- [ ] Validation UX mobile multi-device (Android/iOS)
- [ ] Tests UI/e2e (detox/expo tests)

## 7. Problemes Techniques Ouverts
- Conflits possibles de port local (ex: `8080` pris par Apache/IIS/WSL)
- Configuration reseau mobile physique dependante de `EXPO_PUBLIC_API_URL`
- `main.go` concentre encore beaucoup de logique applicative
- Couverture tests forte backend, faible mobile

## 8. Prochaine Priorite
Stabilisation du chat en mobile (polling temps reel + UX conversation), puis extraction progressive de la logique backend hors `main.go` sans regression.

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

## 11. Date de Derniere Mise a Jour
2026-03-01
