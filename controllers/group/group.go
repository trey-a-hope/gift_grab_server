// Package group contains controllers and types for querying and managing Nakama groups.
package group

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

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

// GetGroupMembershipState fetches the membership state of the current user in a group.
func GetGroupMembershipState(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	userId, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)

	var request GroupRequest

	if err := json.Unmarshal([]byte(payload), &request); err != nil {
		logger.Error("Failed to unmarshal payload: %v", err)
		return "", errors.New("invalid JSON payload")
	}

	groupId := request.GroupId

	if groupId == "" {
		return "", errors.New("group_id cannot be empty")
	}

	// Query membership state from group_edge
	// source_id can be group_id and destination_id can be user_id, or vice-versa
	query := `SELECT state FROM group_edge WHERE (source_id = $1 AND destination_id = $2) OR (source_id = $2 AND destination_id = $1) LIMIT 1`

	var state int
	err := db.QueryRow(query, userId, groupId).Scan(&state)
	if err != nil {
		if err == sql.ErrNoRows {
			return "-1", nil
		}
		logger.Error("Database query error: %v", err)
		return "", err
	}

	return fmt.Sprintf("%d", state), nil
}
