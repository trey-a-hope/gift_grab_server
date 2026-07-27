// Package leaderboard initializes and manages leaderboard configurations.
package leaderboard

import (
	"context"

	"github.com/heroiclabs/nakama-common/runtime"
)

// CreateMonthlyLeaderboard creates or fetches a monthly leaderboard with descending score order.
// The leaderboard resets on the 1st of every month at midnight UTC.
func CreateMonthlyLeaderboard(ctx context.Context, nk runtime.NakamaModule, logger runtime.Logger) error {
	const id = "monthly_leaderboard"
	const authoritative = false
	const sortOrder = "desc"
	const operator = "best"
	const resetSchedule = "0 0 1 * *"
	var metadata = make(map[string]interface{})

	return nk.LeaderboardCreate(ctx, id, authoritative, sortOrder, operator, resetSchedule, metadata)
}
