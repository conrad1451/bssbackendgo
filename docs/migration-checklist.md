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
- [ ] `players.id` remains the only internal ownership identifier
- [ ] All foreign keys reference internal IDs only
- [ ] New columns do not duplicate identity meaning
- [ ] Existing rows can be migrated without ambiguity

---

## 2. Middleware

- [ ] Session validation still happens in `sessionValidationMiddleware`
- [ ] External IDs are resolved before handlers run
- [ ] Internal player ID is attached to request context
- [ ] Context keys use custom types (no string keys)
- [ ] Middleware does not leak auth concerns into handlers

---

## 3. Handlers

- [ ] Handlers do not parse JWTs
- [ ] Handlers do not call Descope directly
- [ ] Authorization checks use `players.id`
- [ ] No handler trusts client-provided IDs
- [ ] All errors return JSON responses

---

## 4. Authorization Rules

- [ ] Admin access is role-based, not route-based
- [ ] Ownership checks happen in SQL
- [ ] No frontend-enforced authorization logic
- [ ] Role names match Descope configuration

---

## 5. API Contracts

- [ ] Error responses are JSON, not text
- [ ] Frontend parsing expectations still match backend responses
- [ ] Status codes are unchanged or documented
- [ ] Breaking changes are versioned or coordinated

---

## 6. Documentation

- [ ] `docs/architecture.md` updated
- [ ] README architecture summary still accurate
- [ ] New invariants are documented
- [ ] Removed invariants are explicitly deleted

---

## 7. Deployment Safety

- [ ] Migration tested against production-like data
- [ ] Rollback plan exists
- [ ] No partial deploy leaves mixed identity logic live
