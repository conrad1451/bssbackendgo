# Architecture Overview

This document describes the core architectural decisions, trust boundaries,
and invariants of the **bssbackendgo** service.

Its purpose is to prevent regressions, clarify intent, and make the system
safe to extend.

---

## 1. Core Principle

**Authentication and authorization are separate concerns.**

- Authentication answers: _Who is the user?_
- Authorization answers: _What are they allowed to do?_

This service enforces that separation strictly.

---

## 2. Identity Model

### External Identity (Descope)

- Descope is the **source of truth for authentication**
- Descope issues session tokens (JWTs)
- Descope assigns a stable external user identifier (`token.ID`)
- External IDs are **never trusted for authorization**

Descope IDs are considered **untrusted input** until resolved internally.

---

### Internal Identity (Database)

The database defines the system’s **internal identity model**.

#### Players table

- `players.id`
  - Internal primary key
  - Used for ownership, joins, and authorization
- `players.player_id`
  - Stores the external Descope user ID
  - Used only for identity resolution

> **Invariant:**  
> Descope IDs are never used as database primary keys or foreign keys.

---

## 3. Identity Resolution Flow

All requests pass through `sessionValidationMiddleware`.

### Middleware responsibilities

1. Extract `Authorization: Bearer <token>`
2. Validate session with Descope
3. Resolve or create a player row
4. Store identity data in request context

### Context values set

- `contextKeyUserID` → Descope user ID (string)
- `contextKeyPlayerID` → Internal player ID (int)
- `contextKeyExternalPlayerID` → Descope player ID (string)
- `contextKeyIsAdmin` → boolean

> **Rule:**  
> Handlers must not re-parse JWTs or query Descope directly.

---

## 4. Request Lifecycle

```text
HTTP Request
→ sessionValidationMiddleware
   → validate JWT (Descope)
   → resolve / create player row
   → attach identity to context
→ handler
   → authorization checks
   → SQL queries scoped by players.id
→ response
```

## 5. Authorization Model

Admin status is derived from Descope roles

Admin access is role-based, not route-based

Ownership checks are enforced in SQL using internal IDs

Example pattern:

WHERE resource.owner_player_id = $1

Invariant:
All authorization checks must use players.id, never external IDs.

## 6. Security & Trust Boundaries

What the client can provide

Session token (JWT)

What the client cannot provide

Player IDs

Admin status

Ownership claims

All sensitive identity data is resolved server-side.

## 7. Design Decisions (Non-Negotiable)

Descope IDs are never used as DB PKs

Handlers do not parse JWTs

Middleware owns identity resolution

Authorization logic lives in SQL

Context keys use custom types to avoid collisions

Violating any of these rules introduces security and data integrity risks.

## 8. Common Anti-Patterns (Do Not Do This)

❌ Accepting player IDs from request bodies
❌ Using Descope IDs in SQL joins
❌ Skipping middleware for “internal” routes
❌ Performing ownership checks in frontend code
❌ Deriving admin access from client flags

## 9. Why This Exists

This architecture exists to:

Prevent identity confusion bugs

Make authorization auditable

Allow authentication providers to change without DB rewrites

Keep handlers simple and safe

If something feels harder because of this structure, that friction is
intentional.

## 10. Future Changes

Any change that affects:

Identity resolution

Authorization rules

Context values

Player creation logic

must update this document.

Architecture drift is considered a bug.
