package checkpointrepo

import (
	"context"
	"database/sql"
	"time"
)

// Checkpoint represents a persisted game checkpoint.
type Checkpoint struct {
	ID        int
	PlayerID int
	Name      string
	Lat       float64
	Lng       float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CreateCheckpoint inserts a new checkpoint owned by the given player
// and returns the newly created checkpoint ID.
func CreateCheckpoint(ctx context.Context, db *sql.DB, playerID int, cp *Checkpoint) (int, error) {
	var id int
	err := db.QueryRowContext(ctx, `
		INSERT INTO checkpoints (player_id, name, lat, lng)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, playerID, cp.Name, cp.Lat, cp.Lng).Scan(&id)

	return id, err
}

// GetCheckpointByID returns a checkpoint by its ID.
func GetCheckpointByID(ctx context.Context, db *sql.DB, checkpointID int) (*Checkpoint, error) {
	var cp Checkpoint
	err := db.QueryRowContext(ctx, `
		SELECT id, player_id, name, lat, lng, created_at, updated_at
		FROM checkpoints
		WHERE id = $1
	`, checkpointID).Scan(
		&cp.ID,
		&cp.PlayerID,
		&cp.Name,
		&cp.Lat,
		&cp.Lng,
		&cp.CreatedAt,
		&cp.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &cp, nil
}

// GetCheckpointsByPlayer returns all checkpoints owned by a player.
func GetCheckpointsByPlayer(ctx context.Context, db *sql.DB, playerID int) ([]Checkpoint, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, player_id, name, lat, lng, created_at, updated_at
		FROM checkpoints
		WHERE player_id = $1
		ORDER BY created_at DESC
	`, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var checkpoints []Checkpoint
	for rows.Next() {
		var cp Checkpoint
		if err := rows.Scan(
			&cp.ID,
			&cp.PlayerID,
			&cp.Name,
			&cp.Lat,
			&cp.Lng,
			&cp.CreatedAt,
			&cp.UpdatedAt,
		); err != nil {
			return nil, err
		}
		checkpoints = append(checkpoints, cp)
	}

	return checkpoints, rows.Err()
}

// UpdateCheckpoint updates mutable fields of a checkpoint.
func UpdateCheckpoint(ctx context.Context, db *sql.DB, checkpointID int, cp *Checkpoint) error {
	_, err := db.ExecContext(ctx, `
		UPDATE checkpoints
		SET name = $1, lat = $2, lng = $3, updated_at = NOW()
		WHERE id = $4
	`, cp.Name, cp.Lat, cp.Lng, checkpointID)

	return err
}

// DeleteCheckpoint removes a checkpoint by ID.
func DeleteCheckpoint(ctx context.Context, db *sql.DB, checkpointID int) error {
	_, err := db.ExecContext(ctx, `
		DELETE FROM checkpoints
		WHERE id = $1
	`, checkpointID)

	return err
}



// // CHQ: Gemini AI created function
// // Example function to retrieve a user's checkpoints
// func GetUserCheckpoints(userID string) ([]Checkpoint, error) {
//     // 1. Prepare a slice to hold the checkpoints.
//     var checkpoints []Checkpoint

//     // 2. Query the database for all checkpoints belonging to the given user ID.
//     // The query uses a parameterized statement ($1) to prevent SQL injection.
//     rows, err := db.Query("SELECT checkpoint_id, title, data, created_at, updated_at FROM game_checkpoints WHERE user_id = $1", userID)
//     if err != nil {
//         return nil, fmt.Errorf("failed to query game_checkpoints for user %s: %w", userID, err)
//     }
//     defer rows.Close()

//     // 3. Iterate through the result set and scan each row into a Checkpoint struct.
//     for rows.Next() {
//         var cp Checkpoint
//         err := rows.Scan(&cp.ID, &cp.Title, &cp.Data, &cp.CreatedAt, &cp.UpdatedAt)
//         if err != nil {
//             return nil, fmt.Errorf("failed to scan checkpoint row: %w", err)
//         }
//         checkpoints = append(checkpoints, cp)
//     }

//     // 4. Check for any errors that occurred during the iteration.
//     if err = rows.Err(); err != nil {
//         return nil, fmt.Errorf("error during row iteration: %w", err)
//     }

//     // 5. Return the slice of checkpoints.
//     return checkpoints, nil
// }