## Architecture Enforcement Scripts

This repository includes a set of automated checks that enforce the architectural
rules described in [`docs/architecture.md`](docs/architecture.md).

These scripts exist to **prevent architectural regressions**, not just to document intent.

### Location

```text
/scripts/architecture/
```

### Script Overview

Each script maps **one-to-one** with a section in `docs/architecture.md`:

```text
01-auth-vs-authz.sh          → Authentication vs Authorization
02-identity-model.sh         → Identity Model (external vs internal)
03-identity-resolution.sh    → Player resolution & creation
04-request-lifecycle.sh      → Middleware enforcement
05-authorization-model.sh    → Authorization rules
06-trust-boundaries.sh       → Client trust boundaries
07-design-invariants.sh      → Non-negotiable invariants
08-anti-patterns.sh          → Explicitly forbidden patterns
run-architecture-checks.sh   → Runs all checks
```

### How These Scripts Are Used

- Run locally before committing architectural changes
- Executed in CI to block unsafe changes
- Act as executable documentation for system invariants

Example:

```bash
./scripts/architecture/run-architecture-checks.sh
```

If any script fails, the build should fail.

---

### Why These Scripts Exist

Architecture docs describe **what should be true**.
These scripts verify **that it is still true**.

They ensure:

- Descope IDs are never used as database primary keys
- All identity resolution happens in middleware
- Authorization is enforced server-side and in SQL
- Handlers never trust client-supplied identity
- Admin access is role-based, not route-based

If something “small” breaks a script, it usually means something **important** was violated.

---

### Making Changes

If you modify:

- Identity resolution
- Authorization rules
- Context values
- Player creation logic
- Trust boundaries

You must update **both**:

- `docs/architecture.md`
- The corresponding script in `scripts/architecture/`

Architecture drift is treated as a bug.
