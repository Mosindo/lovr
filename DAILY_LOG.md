# Lovr - Daily Log

## 2026-03-10

### Fait Aujourd'hui
- Reprise projet sur les priorites `PROJECT_STATUS.md` sans ajout de nouvelle feature.
- Stabilisation UX chat mobile sur `apps/mobile/src/screens/ChatsScreen.tsx`:
  - gestion des erreurs de rafraichissement silencieux (`backgroundError`) au lieu de masquer les echecs polling;
  - action `Reload/Retry` contextualisee via une seule routine de relance;
  - ouverture de conversation rendue plus robuste (on conserve le contexte de chat, sans retour force a la liste en cas d'echec de load);
  - desactivation du bouton `Send` pendant `loading`/`sending` pour eviter les actions concurrentes.
- Observabilite backend minimale renforcee:
  - `request_start` middleware ajoute dans `services/api/cmd/api/main.go`;
  - log HTTP structure enrichi avec `user_id` et latence calculee depuis le contexte requete;
  - logs handlers (`Auth/Social/Chat`) homogenises avec `status` et `latency_ms`;
  - logs evenementiels handlers (`register/login/like/block/send`) enrichis avec `status` et `latency_ms`.
- Documentation statut projet mise a jour (`PROJECT_STATUS.md`).
- Alternative simple a Maestro ajoutee: script `scripts/qa-lite.ps1` (gate complete backend + mobile sans outillage device).
- Documentation README alignee sur le flux `QA lite`.
- CI ajoutee: workflow GitHub Actions `.github/workflows/qa-lite.yml` (Docker stack + healthcheck + `qa-lite` + artefacts rapports).

### Problemes Rencontres
- `go test ./...` et `go build ./cmd/api` ont depasse le timeout par defaut de l'outil, puis ont ete relances avec timeout etendus (resultat OK).

### Decisions Prises
- Prioriser uniquement les corrections de stabilisation/observabilite deja listees comme prioritaires.
- Ne pas introduire de nouvelle dependance ni de refactor structurel hors besoin direct.
- En cas de blocage outillage e2e UI device, utiliser `QA lite` comme gate de livraison continue.

### Impact Estime Sur L'avancement
- +2% Backend API (logs structures plus actionnables en exploitation).
- +2% Mobile App (gestion retry/error chat plus robuste sur reseau instable).
- +1% Tests & qualite (revalidation complete backend + smoke API execute).

### Prochaine Action
- Validation UX mobile multi-device (Android/iOS) avec run complet device reel.
- Execution/fiabilisation CI des tests UI/e2e (Maestro puis trajectoire Detox).

---

## 2026-03-02

### Fait Aujourd'hui
- Correction configuration projet: suppression de `.env.example` (usage `.env` uniquement) et alignement README sur `http://localhost:18080/health`.
- Refactor incremental backend termine pour modulariser l'API sans changer le comportement:
  - extraction Auth vers `internal/auth`, `internal/services`, `internal/handlers`
  - extraction Social (`discover`, `likes`, `matches`, `block`) vers `internal/services`, `internal/handlers`
  - extraction Chat (`chats`, `messages`) vers `internal/services`, `internal/handlers`
- `cmd/api/main.go` reduit au role de composition root (routing + wiring services/handlers + middleware).
- Ajustement des types utilises par les tests d'integration pour rester independants des types internes de `main.go`.
- Ajout de tests d'integration cibles sur la couche services (`internal/services/services_integration_test.go`) couvrant:
  - AuthService: register/login/me + erreurs
  - SocialService: discover/like/match/block
  - ChatService: match requis, send/list, blocage post-block
- Mobile: ajout du polling conversation/listing sur l'ecran chats (`ChatsScreen`) avec rafraichissement silencieux.
- Validation frontend: `npx tsc --noEmit` OK dans `apps/mobile`.
- Ajout d'un script smoke e2e mobile automatisable (`apps/mobile/scripts/e2e-smoke.js`) couvrant:
  - register/login
  - discover
  - like -> match
  - chat send/list
  - block + verification post-block
- Validation smoke e2e mobile: `npm run e2e:smoke` OK.
- Mise en place d'une base e2e UI device:
  - `apps/mobile/e2e/maestro/critical-flow.yaml` (flow critique)
  - `apps/mobile/scripts/e2e-ui-setup.js` (fixtures API avec compte deja matche)
  - `npm run e2e:ui:setup` pour generer `.e2e/maestro-env.ps1` et `.e2e/maestro-env.json`
- Ajout des `testID` sur ecrans mobile critiques (Auth/Discover/Matches/Chats/Account) pour stabiliser les selectors e2e.

### Problemes Rencontres
- Echec temporaire des tests d'integration apres deplacement de types (erreurs de compilation sur payloads): corrige par types dedies dans `main_integration_test.go`.
- `docker compose ps` non utilisable depuis ce shell (acces daemon refuse), mais verification API possible via port et smoke HTTP.

### Decisions Prises
- Continuer les refactors par lots verticaux (Auth -> Social -> Chat) pour limiter le risque de regression.
- Conserver strictement les contrats API existants (routes, statuts HTTP, messages d'erreur) pendant la modularisation.
- Garder les validations de qualite apres chaque lot: `gofmt ./...`, `go test ./...`, `go build ./cmd/api` + smoke HTTP.

### Impact Estime Sur L'avancement
- +14% sur la maintenabilite backend (separation handlers/services effective).
- +7% sur la couverture qualite backend (tests services en plus des tests API).
- +8% sur la reduction du risque de changement futur (main simplifie, logique metier isolee).
- +5% sur la fiabilite operationnelle (config `.env` clarifiee, doc healthcheck alignee).
- +6% sur l'experience chat mobile (rafraichissement conversation quasi temps reel sans spinner bloquant).
- +6% sur la testabilite mobile (parcours critique automatisable en script smoke).
- +6% sur la preparabilite e2e UI device (flow Maestro + setup fixtures + selectors stables).

### Prochaine Action
- Finaliser la documentation d'architecture backend (schema des modules `auth/social/chat` + conventions).
- Etendre les tests services sur cas limites supplementaires (pagination/limites, erreurs DB transientes).
- Executer le flow Maestro sur device/emulateur avec outillage installe (`maestro`, `adb`) et fiabiliser les assertions runtime.

---

## 2026-03-01

### Fait Aujourd'hui
- Mise en place des flows backend: auth, discover, likes/matches, block atomique, chat.
- Ajout et validation des migrations SQL jusqu'a `004_create_messages.sql`.
- Ajout des tests d'integration backend couvrant auth/social/chat et cas d'acces interdit.
- Stabilisation mobile: auth, discover, matches, block, chats (liste + conversation + envoi).
- Resolution des problemes reseau de test mobile (port host API dedie + URL LAN).

### Problemes Rencontres
- Conflit de port local `8080` avec un service Apache, causant des faux 404 HTML.
- Erreurs reseau mobile liees a `localhost` utilise sur telephone physique.
- Etat de chargement conversation chat bloquant suite a une boucle de refresh.

### Decisions Prises
- Utilisation du port host API `18080` pour eviter collisions locales.
- Timeout reseau frontend ajoute pour eviter spinner infini.
- Regles anti-recurrence formalisees dans `AGENTS.md` (reseau + readiness).
- Chat autorise strictement uniquement apres match reciproque et hors blocage.

### Impact Estime Sur L'avancement
- +18% sur le backend MVP (social + chat + tests integration).
- +12% sur la maturite mobile (flows testables bout en bout hors temps reel).
- +10% sur la fiabilite de run local (network/port/debug).

### Prochaine Action
- Ajouter polling conversation (ou websocket) pour experience chat temps reel.
- Poursuivre la modularisation backend (`handlers/services`) sans changer le comportement.
- Renforcer tests mobiles e2e sur les parcours critiques.

---

## Template Quotidien

## 2026-03-03

### Fait Aujourd'hui
- Priorite unique executee: mise en place de la couche `internal/repositories` pour `Auth/Social/Chat`.
- SQL extrait des services metier:
  - `internal/services/auth_service.go`
  - `internal/services/social_service.go`
  - `internal/services/chat_service.go`
- Repositories Postgres ajoutes:
  - `internal/repositories/auth_repo.go`
  - `internal/repositories/social_repo.go`
  - `internal/repositories/chat_repo.go`
- Wiring recable dans `cmd/api/main.go`: `repositories -> services -> handlers`.
- Comportement externe conserve (routes/payloads/codes HTTP inchanges).
- Lot suivant lance: priorite `PROJECT_STATUS` executee partiellement:
  - tentative d'execution e2e UI reelle via `npm run e2e:ui:run`
  - stabilisation UX chat sur `ChatsScreen` (loading/error/retry) sans nouvelle feature.
- Validation post-ajustements UX:
  - `npx tsc --noEmit` OK
  - `npm run e2e:smoke` OK
  - `go test ./...` OK
  - `go build ./cmd/api` OK
- Automatisation smoke MVP courte ajoutee:
  - `scripts/smoke-api.ps1` (API smoke automatise)
  - `scripts/smoke-all.ps1` (orchestration globale)
  - rapport automatique `./.smoke/report.txt` (PASS/FAIL)
- Execution reelle validee:
  - `powershell -ExecutionPolicy Bypass -File .\\scripts\\smoke-api.ps1` -> PASS
  - `powershell -ExecutionPolicy Bypass -File .\\scripts\\smoke-all.ps1` -> PASS
- Stabilisation UX chat complementaire:
  - correction d'un faux echec possible pendant `openChat` quand un fetch messages est deja en cours (`ChatsScreen`).
  - revalidation mobile: `npx tsc --noEmit` OK, `npm run e2e:smoke` OK.
- Outillage e2e UI prepare:
  - installation locale de `maestro` dans `C:\Users\mosin\.maestro\bin`
  - installation `adb` via `Google.PlatformTools`
  - relance `npm run e2e:ui:run` avec setup fixtures OK.
- Stabilisation UX chat renforcee (sans changement fonctionnel):
  - desactivation des actions `Reload`/`Retry` pendant `loading` ou `sending` pour eviter les relances concurrentes.
- Observabilite backend minimale ajoutee:
  - middleware `request_id` (`X-Request-ID`) + logs HTTP structures (method/path/status/latency/client_ip/request_id).
- Hardening backend additionnel:
  - logs structures d'erreurs handlers (`Auth/Social/Chat`) et evenements metier sensibles (register/login/like/block/send).
  - metriques HTTP minimales en memoire (count/latence moyenne/max) journalisees periodiquement.
- Pagination/limites robustes (sans rupture API):
  - `GET /discover?limit=` bornage service/repository (default 50, max 100).
  - `GET /chats/:userId/messages?limit=` bornage service/repository (default 200, max 500).
- Tests integration et non-regression ajoutes:
  - propagation `X-Request-ID` sur `/health`.
  - couverture `limit` pour discover et messages chat.
  - validations OK: `go test ./...`, `go build ./cmd/api`, `npx tsc --noEmit`, `npm run e2e:smoke`.
- Runs de reproductibilite executes:
  - `scripts/smoke-api.ps1` x2 -> PASS.
  - `npm run e2e:smoke` x2 -> PASS.
- E2E UI device:
  - environnement `maestro` + `adb` operationnel avec emulateurs detectes.
  - execution complete reste instable/longue en session locale (timeouts/interruptions), sans verdict final fiable.
- Orientation test UI:
  - decision de preparer une migration progressive vers Detox (Maestro conserve en fallback hardening).

### Problemes Rencontres
- Un echec de compilation temporaire sur mapping de types Chat (`repositories.ChatMessage` vs `services.ChatMessage`), corrige sans impact externe.
- Acces au cache build Go restreint dans l'environnement sandbox pour `go test`, necessitant execution avec permissions elevees.
- Execution e2e UI bloquee localement par outillage manquant: commande `maestro` absente du PATH.
- Execution e2e UI encore bloquee en run complet faute de device/emulateur connecte (`0 devices connected`).
- Even avec devices detectes, run UI long et fragile en local (contexte shell/permissions/sessions interrompues).

### Decisions Prises
- Garder une extraction 1:1 minimale (pas de refactor cosmetique).
- Laisser les signatures handlers/services externes stables.
- Valider apres chaque recablage domaine (Auth, puis Social, puis Chat).
- Conserver la priorite QA: pas de nouvelle feature tant que l'execution e2e UI reelle n'est pas validee.
- Ne pas toucher aux routes/payloads/codes HTTP pendant la stabilisation (uniquement UX chat + logs techniques).
- Maintenir `smoke-api` + `e2e:smoke` comme gate MVP, et traiter UI e2e device comme bonus hardening.

### Impact Estime Sur L'avancement
- +12% sur la maintenabilite backend (separation data access / logique metier finalisee).
- +8% sur la reduction du risque de regression lors des evolutions futures backend.

### Prochaine Action
- Priorite MVP en cours: finaliser stabilisation UX chat et observabilite backend (ajustements mineurs restants + revue tech lead).
- Bonus phase 1: reprendre execution e2e UI reelle sur device/emulateur lorsque la session run longue est disponible.
- Proposer plan minimal de migration Maestro -> Detox (sans changement produit) pour fiabiliser la QA UI.

---

## 2026-03-03 (Plan de Travail)

### Tache Prioritaire
- Mettre en place la couche repository dediee backend (SQL sorti des services) sans changer le comportement API.

### Objectif de la Journee
- Stabiliser le MVP en separant clairement acces donnees et logique metier sur `Auth/Social/Chat`.

### Plan (Matin / Apres-midi)
- Matin:
  - Definir interfaces repository minimales par domaine (`AuthRepository`, `SocialRepository`, `ChatRepository`).
  - Implementer repository Auth et recabler `AuthService`.
  - Implementer repository Social et recabler `SocialService`.
- Apres-midi:
  - Implementer repository Chat et recabler `ChatService`.
  - Ajuster le wiring `main.go` (repositories -> services -> handlers) sans changer routes/payloads.
  - Lancer validations backend/mobile et corriger toute regression.
  - Mettre a jour `PROJECT_STATUS.md` + `DAILY_LOG.md` en fin de journee.

### Commandes de Validation
- Backend:
  - `cd services/api`
  - `gofmt -w .`
  - `go test ./...`
  - `go build ./cmd/api`
- Mobile:
  - `cd apps/mobile`
  - `npx tsc --noEmit`
  - `npm run e2e:smoke`

### Definition de Done
- SQL retire des services et centralise en repositories dedies.
- Contrats API inchanges (comportements conserves).
- Build/tests backend OK + checks mobile OK.
- Statuts projet/journal mis a jour.

### Risques
- Regression de mapping erreurs service/repository.
- Rupture tests integration si signatures changent.
- Refactor trop large: rester strictement sur la stabilisation MVP.

## YYYY-MM-DD

### Fait Aujourd'hui
- ...

### Problemes Rencontres
- ...

### Decisions Prises
- ...

### Impact Estime Sur L'avancement
- ...%

### Prochaine Action
- ...
