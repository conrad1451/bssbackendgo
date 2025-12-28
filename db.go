var db *sql.DB

func resolveOrCreatePlayer(ctx context.Context, externalID string) (int, error)
func resolveOrCreateUser(ctx context.Context, playerID int) (int, string, error)
