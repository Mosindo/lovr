# AGENTS.md — Lovr

## Objectif
App de rencontre "human-first" : swipe + matching par centres d'intérêt, chat après match, blocage atomique.

## Stack
- Mobile: React Native (Expo) + TypeScript
- API: Go
- DB: Postgres (docker-compose)

## Règles de travail
- Toujours proposer un plan court (5–10 étapes) avant une grosse modif.
- Changer petit et souvent (pas de refactor géant).
- Toujours donner les commandes pour vérifier (lint/test/run).
- Pas de secrets dans le repo. Ne jamais écrire de valeurs .env dans le code.