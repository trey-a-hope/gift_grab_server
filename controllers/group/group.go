package group

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/heroiclabs/nakama-common/runtime"
)

// GetGroupById fetches a group from Nakama by its ID.
func GetGroupById(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	var request GroupRequest

	if err := json.Unmarshal([]byte(payload), &request); err != nil {
		logger.Error("Failed to unmarshal payload: %v", err)
		return "", errors.New("invalid JSON payload")
	}

	if request.GroupId == "" {
		return "", errors.New("group_id cannot be empty")
	}

	// Fetch the group by ID using Nakama Module API
	groups, err := nk.GroupsGetId(ctx, []string{request.GroupId})
	if err != nil {
		logger.Error("Failed to fetch group: %v", err)
		return "", err
	}

	// Return error if no groups were returned matching the ID
	if len(groups) == 0 {
		return "", errors.New("group not found")
	}

	// Format and marshal response
	response := NewGroupResponse(groups[0])
	responseBytes, err := json.Marshal(response)
	if err != nil {
		logger.Error("Failed to marshal group response: %v", err)
		return "", err
	}

	return string(responseBytes), nil
}
