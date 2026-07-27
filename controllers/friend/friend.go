// Package friend handles retrieving and updating friendship/connection states between users.
package friend

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/heroiclabs/nakama-common/runtime"
)

// GetFriendshipState queries the Nakama database to find the friendship edge state between the calling user and a target destination user.
// Returns the status code as a string (or "-1" if no relationship exists).
func GetFriendshipState(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	// Extract the authenticated user ID from context
	source_id, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)

	var request struct {
		DestinationId string `json:"destination_id"`
	}

	// Parse JSON input payload
	if err := json.Unmarshal([]byte(payload), &request); err != nil {
		logger.Error("Failed to unmarshal payload: %v", err)
		return "", errors.New("invalid JSON payload")
	}

	destination_id := request.DestinationId

	// Validate required fields
	if destination_id == "" {
		return "", errors.New("destination_id cannot be empty")
	}

	// Query Nakama's user_edge table to fetch friendship state
	query := `SELECT state FROM user_edge WHERE (destination_id = $2 AND source_id = $1) LIMIT 1`

	var state int

	err := db.QueryRow(query, source_id, destination_id).Scan(&state)

	if err != nil {
		// If no row exists, represent this as state -1
		if err == sql.ErrNoRows {
			return "-1", nil
		}
		logger.Error("Database query error: %v", err)
		return "", err
	}

	return fmt.Sprintf("%d", state), nil
}
