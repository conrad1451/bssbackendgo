# bssbackendgo

A backend server for the bss game. Written in Go (GoLang).

## To run the API server:

go run ./cmd/api

## To run the migration tool:

go run ./cmd/migrate

## Architecture Overview

### 1. Authentication vs Authorization

- Descope is used for authenticating users
- This backend is used to validate session tokens
- Identities provided by client are not trusted
- Roles are derived server-side

### 2️. Identity Model

**External Identity (Descope)**
→ Used for authentication only

**Internal Identity (Database)**
→ Used for authorization & ownership

While `players.id` is used as the internal primary key, `players.player_id` stores the Descope-assigned player ID.
This prevents future regressions.

### 3. Request Lifecycle

Request
→ JWT validation (Descope)
→ Resolve / create player row
→ Store IDs in request context
→ Handler uses ctx[playerID]
→ SQL queries

→ Store internal player ID, external player ID, and admin role in request context

### 4. Design Decisions

“Descope IDs are never used as DB PKs”

“All ownership checks happen in SQL”

“Admin access is role-based, not route-based”

This shows why things are the way they are.

All requests pass through `sessionValidationMiddleware`
before reaching handlers.
