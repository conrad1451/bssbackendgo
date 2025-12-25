> This checklist is required for all identity or authorization changes.
> Some items are enforced automatically via architecture scripts.
> Others require explicit human review and sign-off.

# Identity & Authorization Migration Checklist

Run this checklist for any change involving:

- authentication
- authorization
- identity fields
- players / users tables
- middleware
- session handling

---

## 1. Database Layer

- [ ] No external auth IDs used as primary keys
      🛠 01_identity_model.sh
- [ ] `players.id` remains the only internal ownership identifier
      🛠 01_identity_model.sh
- [ ] All foreign keys reference internal IDs only
      🛠 01_identity_model.sh
- [ ] New columns do not duplicate identity meaning
      🧠 (requires schema intent review)
- [ ] Existing rows can be migrated without ambiguity
      🧠 (data reasoning, cannot be automated)

---

## 2. Middleware

- [ ] Session validation still happens in `sessionValidationMiddleware`
      🛠 02_identity_resolution.sh
- [ ] External IDs are resolved before handlers run
      🛠 02_identity_resolution.sh
- [ ] Internal player ID is attached to request context
      🛠 02_identity_resolution.sh
- [ ] Context keys use custom types (no string keys)
      🛠 06_context_safety.sh
- [ ] Middleware does not leak auth concerns into handlers
      ⚠️ Script can detect obvious violations, intent still reviewed

---

## 3. Handlers

- [ ] Handlers do not parse JWTs
      🛠 02_identity_resolution.sh
- [ ] Handlers do not call Descope directly
      🛠 02_identity_resolution.sh
- [ ] Authorization checks use `players.id`
      🛠 04_authorization_rules.sh
- [ ] No handler trusts client-provided IDs
      🛠 04_authorization_rules.sh
- [ ] All errors return JSON responses
      🛠 05_error_handling.sh

---

## 4. Authorization Rules

- [ ] Admin access is role-based, not route-based
      ⚠️ Script can detect patterns, role semantics require review
- [ ] Ownership checks happen in SQL
      🛠 04_authorization_rules.sh
- [ ] No frontend-enforced authorization logic
      🧠 (outside backend repo scope)
- [ ] Role names match Descope configuration
      🧠 (external system validation)

---

## 5. API Contracts

- [ ] Error responses are JSON, not text
      🛠 05_error_handling.sh
- [ ] Frontend parsing expectations still match backend responses
      🧠 (cross-repo coordination)
- [ ] Status codes are unchanged or documented
      🧠 (requires release awareness)
- [ ] Breaking changes are versioned or coordinated
      🧠 (process decision)

---

## 6. Documentation

- [ ] `docs/architecture.md` updated
      🛠 07_docs_guard.sh (existence + basic diff checks)
- [ ] README architecture summary still accurate
      🧠 (narrative correctness)
- [ ] New invariants are documented
      🧠 (human responsibility)
- [ ] Removed invariants are explicitly deleted
      🧠 (intent review)

---

## 7. Deployment Safety

- [ ] Migration tested against production-like data
      🧠
- [ ] Rollback plan exists
      🧠
- [ ] No partial deploy leaves mixed identity logic live
      🧠

---

Legend:
🛠 Automatically enforced by architecture scripts
🧠 Requires human review and explicit sign-off
⚠️ Partially enforced — scripts detect violations, intent still reviewed
