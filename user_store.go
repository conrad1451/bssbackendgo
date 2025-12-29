package main // Must match the package name in main.go
import (
	"context"
	"database/sql"
)

// resolveOrCreateUser resolves the internal user record associated with a player.
//
// This function represents the second layer of identity resolution:
//
//   - players represent authenticated identities (one per Descope user)
//   - users represent application-level entities that own gameplay data
//
// Behavior:
//   - If a user row already exists for the given internal player ID,
//     its primary key (users.id) is returned.
//   - If no such row exists, a new user row is created and its ID is returned.
//
// Invariants:
//   - The input playerID MUST be an internal database ID (players.id).
//   - External identity values (e.g. Descope IDs) must never be passed here.
//   - The returned value is always an internal integer ID.
//   - At most one user row exists per player.
//
// Parameters:
//   - ctx: request-scoped context (must not be nil)
//   - db: database connection
//   - playerID: internal player ID (players.id)
//
// Returns:
//   - int: internal user ID (users.id)
//   - error: non-nil if resolution or creation fails
//
// Callers MUST:
//   - Call this only after resolveOrCreatePlayer succeeds
//   - Store the returned user ID in request context
//   - Use the returned ID for all gameplay ownership checks
func resolveOrCreateUser(
	ctx context.Context,
	db *sql.DB,
	playerID int,
) (int, string, error) {

	var userID int
	var username sql.NullString

	err := db.QueryRowContext(ctx, `
		SELECT user_id, user_name
		FROM users
		WHERE player_id = $1
	`, playerID).Scan(&userID, &username)

	if err == sql.ErrNoRows {
		// create user
		err = db.QueryRowContext(ctx, `
			INSERT INTO users (player_id)
			VALUES ($1)
			RETURNING user_id
		`, playerID).Scan(&userID)
		if err != nil {
			return 0, "", err
		}

		// username intentionally empty
		return userID, "", nil
	}

	if err != nil {
		return 0, "", err
	}

	if username.Valid {
		return userID, username.String, nil
	}

	return userID, "", nil
}
