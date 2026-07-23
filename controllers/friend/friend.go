package friend

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/heroiclabs/nakama-common/runtime"
)

func GetFriendshipState(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	source_id, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)

	var request struct {
		DestinationId string `json:"destination_id"`
	}

	if err := json.Unmarshal([]byte(payload), &request); err != nil {
		logger.Error("Failed to unmarshal payload: %v", err)
		return "", errors.New("invalid JSON payload")
	}

	destination_id := request.DestinationId

	if destination_id == "" {
		return "", errors.New("destination_id cannot be empty")
	}

	query := `SELECT state FROM user_edge WHERE (destination_id = $2 AND source_id = $1) LIMIT 1`

	var state int

	err := db.QueryRow(query, source_id, destination_id).Scan(&state)

	if err != nil {
		if err == sql.ErrNoRows {
			return "-1", nil
		}
		logger.Error("Database query error: %v", err)
		return "", err
	}

	return fmt.Sprintf("%d", state), nil
}
