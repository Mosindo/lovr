# AGENTS.md - Lovr

Lovr est une application de rencontre "human-first".
L'architecture doit rester scalable, securisee et maintenable long terme.

---

# OBJECTIF PRODUIT

- Swipe simple
- Matching par centres d'interet
- Chat uniquement apres match
- Blocage atomique
- Architecture compatible forte charge

---

# STACK OFFICIELLE

## Mobile
- React Native (Expo)
- TypeScript
- Navigation simple
- API REST

## API
- Go
- Framework: Gin
- DB: Postgres
- Driver: pgxpool
- Auth: JWT (HS256)
- Pas d'ORM
- Pas de magie implicite

---

# ARCHITECTURE BACKEND

Structure cible:

`cmd/api/main.go`  
`internal/`
- `auth/`
- `handlers/`
- `db/`
- `middleware/`
- `models/`

Regles:
- Aucun melange `net/http` + Gin
- Un seul serveur HTTP
- Handlers fins
- Logique metier hors handlers
- Pas de logique SQL dans handlers

---

# SECURITE OBLIGATOIRE

JWT:
- SigningMethodHMAC
- Alg() == HS256
- Expiration obligatoire
- Validation stricte des claims

Passwords:
- bcrypt
- Aucun stockage en clair

Entrees utilisateur:
- Validation stricte
- Trim + normalisation email

Secrets:
- Jamais en dur
- Variables attendues:
  - `DATABASE_URL`
  - `JWT_SECRET`
  - `PORT` (default 8080)

---

# PERFORMANCE

- Connexion DB via pool unique
- Pas de connexion DB par requete
- Pas de requetes N+1
- Index requis sur:
  - `users.email`
  - `likes (from_user_id, to_user_id)`
  - `blocks (blocker_user_id, blocked_user_id)`

Complexite attendue:
- Matching O(n)
- Pas de O(n^2)

---

# TESTS

Avant toute livraison backend:

- `gofmt ./...`
- `go test ./...`
- `go build ./cmd/api`

Si tests cassent: corriger avant toute suite.

Tests integration a maintenir:
- Health (`/health`)
- Auth (`/auth/register`, `/auth/login`, `/me`)
- Discover (`/discover`)
- Likes (`/likes`)
- Matches (`/matches`)
- Block (`/block`) incluant effets atomiques

---

# RESEAU ET MOBILE (ANTI-RECURRENCE)

Regles obligatoires pour eviter les incidents observes:

- Ne jamais utiliser `localhost` pour les tests sur telephone physique.
- Utiliser `EXPO_PUBLIC_API_URL=http://<LAN_IP>:<PORT_HOST_API>`.
- Verifier qu'aucun service local (Apache/IIS/WSL) ne capte le port API avant annonce "ready":
  - `docker compose ps`
  - `netstat -ano | findstr :<PORT_HOST_API>`
  - `curl http://<LAN_IP>:<PORT_HOST_API>/health`
- Si conflit detecte, changer le port host Docker (ex: `18080:8080`) et aligner:
  - `infra/docker-compose.yml`
  - fallback mobile API
  - script de demarrage mobile
  - README
- L'annonce "frontend pret" n'est valide qu'apres smoke test reel:
  - register
  - login
  - discover load
  - like -> match
  - block (depuis discover ou matches)
  - verification post-block (plus visible, plus match, likes refuses)

---

# REGLES D'EXECUTION POUR L'AGENT

1. Toujours proposer un plan (5-10 etapes).
2. Attendre validation.
3. Appliquer via patch incremental.
4. Ne jamais overwrite complet d'un fichier sans validation explicite.
5. Fournir:
   - diff complet
   - fichiers modifies complets (si demande)
   - sorties build/test
   - nouvelles dependances

---

# INTERDIT

- Refactor massif sans demande
- Nouvelle dependance non justifiee
- Suppression de fichiers sans validation
- Changement d'architecture implicite
- Melange `net/http` et Gin

---

# DEFINITION OF DONE

Une tache est consideree terminee uniquement si:
- code compile
- tests passent
- scenario fonctionnel critique est verifie
- doc d'execution est coherente avec l'etat reel du projet

---

# ROLE DE L'AGENT

L'agent est un ingenieur executant.
Les decisions d'architecture appartiennent au fondateur.
