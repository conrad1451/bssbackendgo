package main

// CHQ: Claude AI refactored this file
// user_store.go

import (
	"context"
	"database/sql"
)

func resolveOrCreateUser(ctx context.Context, descopeID string) (int, error) {
    var userID int

    err := db.QueryRowContext(ctx, `
        SELECT id FROM users WHERE descope_id = $1
    `, descopeID).Scan(&userID)

    if err == sql.ErrNoRows {
        err = db.QueryRowContext(ctx, `
            INSERT INTO users (descope_id)
            VALUES ($1)
            RETURNING id
        `, descopeID).Scan(&userID)
    }

    if err != nil {
        return 0, err
    }

    return userID, nil
}