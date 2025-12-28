
// resolveOrCreatePlayer resolves the internal player record for an authenticated user.
//
// This function enforces the system’s identity boundary:
//
//   - External identity (Descope) is used ONLY for authentication
//   - Internal identity (players.id) is used for authorization and ownership
//
// Behavior:
//   - If a player row already exists for the given external Descope ID,
//     its internal primary key (players.id) is returned.
//   - If no such row exists, a new player row is created and its internal ID
//     is returned.
//
// Invariants:
//   - The external Descope ID is never used as a database primary key.
//   - The returned value is always an internal integer ID.
//   - This function must be called only after session validation succeeds.
//
// Parameters:
//   - ctx: request-scoped context (must not be nil)
//   - db: database connection
//   - externalPlayerID: Descope-assigned user identifier (string)
//
// Returns:
//   - int: internal player ID (players.id)
//   - error: non-nil if resolution or creation fails
//
// Callers MUST:
//   - Store the returned ID in request context
//   - Use the returned ID for all authorization and ownership checks
// func resolveOrCreatePlayer(ctx context.Context, db *sql.DB, externalPlayerID string) (string, error) {
func resolveOrCreatePlayer(
	ctx context.Context,
	externalPlayerID string,
) (int, error) {

	var playerID int

	err := db.QueryRowContext(ctx, `
		SELECT id FROM players WHERE player_id = $1
	`, externalPlayerID).Scan(&playerID)

	if err == sql.ErrNoRows {
		err = db.QueryRowContext(ctx, `
			INSERT INTO players (player_id)
			VALUES ($1)
			RETURNING id
		`, externalPlayerID).Scan(&playerID)
	}

	if err != nil {
		return 0, err
	}

	return playerID, nil
}
