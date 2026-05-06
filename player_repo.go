package main

// CHQ: Claude AI refactored this file
// player_repo.go

import (
	"context"
	"fmt"
)

func getPlayerCount(ctx context.Context, userID int) (int, error) {
    var count int
    err := db.QueryRowContext(ctx, `
        SELECT COUNT(*) FROM players WHERE user_id = $1
    `, userID).Scan(&count)
    return count, err
}

func createPlayer(ctx context.Context, userID int, playername string) (int, error) {
    count, err := getPlayerCount(ctx, userID)
    if err != nil {
        return 0, err
    }
    if count >= 10 {
        return 0, fmt.Errorf("player limit reached")
    }

    var playerID int
    err = db.QueryRowContext(ctx, `
        INSERT INTO players (user_id, playername)
        VALUES ($1, $2)
        RETURNING id
    `, userID, playername).Scan(&playerID)

    return playerID, err
}

func getPlayersByUser(ctx context.Context, userID int) ([]Player, error) {
    rows, err := db.QueryContext(ctx, `
        SELECT id, user_id, playername, created_at
        FROM players
        WHERE user_id = $1
        ORDER BY created_at ASC
    `, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var players []Player
    for rows.Next() {
        var p Player
        if err := rows.Scan(&p.ID, &p.UserID, &p.Playername, &p.CreatedAt); err != nil {
            return nil, err
        }
        players = append(players, p)
    }

    return players, rows.Err()
}