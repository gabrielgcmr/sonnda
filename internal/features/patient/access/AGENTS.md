<!-- internal/features/patient/access/AGENTS.md -->
# Patient access

- This feature owns the association between accounts and patients: active grants, revocation, relationship metadata, and accessible-patient queries.
- Access answers whether an account is currently linked to a patient. It does not decide which actions the account may perform.
- Keep action-level permissions and professional role policies out of this feature; those belong to a future `authz` feature.
- Keep `RelationshipType` as relationship metadata. Do not derive permissions from it or expose it in new list responses without an explicit domain decision.
- Keep access entities and validation rules in `domain/`, application services and repository contracts in the feature root, HTTP in `http/`, and the pgx/sqlc adapter in `postgres/`.
- The `/v1/me/patients` listing belongs to this feature and must not expose `RelationshipType` until the domain defines its public meaning.
- Shared database clients and generated sqlc code remain in `internal/infrastructure`.
- Access request types are dormant domain code until their workflow and authorization rules are defined; do not expose request endpoints during the structural migration.
- Keep the shared patient access checker in this feature and express its input as account and patient identifiers.
