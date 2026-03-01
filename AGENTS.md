# AGENTS.md — Lovr

Lovr est une application de rencontre "human-first".
Architecture pensée pour être scalable, sécurisée et maintenable long terme.

---

# 🎯 OBJECTIF PRODUIT

- Swipe simple
- Matching par centres d’intérêt
- Chat uniquement après match
- Blocage atomique
- Architecture compatible forte charge

---

# 🧱 STACK OFFICIELLE

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
- Pas d’ORM
- Pas de magie implicite

---

# 🏗 ARCHITECTURE BACKEND

Structure attendue :

cmd/api/main.go
internal/
  auth/
  handlers/
  db/
  middleware/
  models/

Règles :
- Aucun mélange net/http + Gin
- Un seul serveur HTTP
- Handlers fins
- Logique métier hors handlers
- Pas de logique SQL dans handlers

---

# 🔐 SÉCURITÉ OBLIGATOIRE

JWT :
- SigningMethodHMAC
- Alg() == HS256
- Expiration obligatoire
- Validation stricte des claims

Passwords :
- bcrypt
- Aucun stockage en clair

Entrées utilisateur :
- Validation stricte
- Trim + normalisation email

Secrets :
- Jamais en dur
- Variables attendues :
  - DATABASE_URL
  - JWT_SECRET
  - PORT (default 8080)

---

# ⚡ PERFORMANCE

- Connexion DB via pool unique
- Pas de connexion DB par requête
- Pas de requêtes N+1
- Index requis sur :
  - users.email
  - likes (from_user, to_user)

- Complexité attendue :
  - Matching O(n)
  - Pas de O(n²)

---

# 🧪 TESTS

Avant toute livraison :

- gofmt ./...
- go test ./...
- go build ./cmd/api

Si tests cassent → corriger avant toute suite.

---

# 🛠 RÈGLES D’EXÉCUTION POUR L’AGENT

1. Toujours proposer un plan (5–10 étapes).
2. Attendre validation.
3. Appliquer via patch incrémental.
4. Ne jamais overwrite complet d’un fichier sans validation explicite.
5. Fournir :
   - diff complet
   - fichiers modifiés complets
   - sorties build/test
   - nouvelles dépendances

---

# 🚫 INTERDIT

- Refactor massif sans demande
- Nouvelle dépendance non justifiée
- Suppression de fichiers
- Changement d’architecture implicite
- Mélange net/http et Gin

---

# 🧠 RÔLE DE L’AGENT

L’agent est un ingénieur exécutant.
Les décisions d’architecture appartiennent au fondateur.