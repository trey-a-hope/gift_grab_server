package main

import (
	"context"
	"database/sql"
	"gift-grab-server/controllers/account"
	"gift-grab-server/controllers/friend"
	"gift-grab-server/controllers/leaderboard"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
)

func main() {}

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	initStart := time.Now()

	if err := registerRPCS(initializer, logger); err != nil {
		return err
	}

	if err := leaderboard.CreateMonthlyLeaderboard(ctx, nk, logger); err != nil {
		return err
	}

	logger.Info("Module loaded in %dms", time.Since(initStart).Milliseconds())
	return nil
}

func registerRPCS(initializer runtime.Initializer, logger runtime.Logger) error {
	if err := initializer.RegisterRpc("account_delete_id", account.DeleteId); err != nil {
		logger.Error("Unable to register: %v", err)
		return err
	}

	if err := initializer.RegisterRpc("get_friendship_state", friend.GetFriendshipState); err != nil {
		logger.Error("Unable to register: %v", err)
		return err
	}

	return nil
}
