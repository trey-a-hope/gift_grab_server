// Package main serves as the entry point for the Nakama server module.
// It initializes and registers RPC handlers, hooks, and leaderboards.
package main

import (
	"context"
	"database/sql"	
	"gift-grab-server/controllers/auth"
	"gift-grab-server/controllers/account"
	"gift-grab-server/controllers/friend"
	"gift-grab-server/controllers/group"
	"gift-grab-server/controllers/leaderboard"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
)

// main is the standard Go entrypoint, unused by Nakama which loads the shared library instead.
func main() {}

// InitModule is the main entry point required by Nakama to initialize this module.
// It registers custom RPC endpoints, before-authentication hooks, and sets up leaderboards.
func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	initStart := time.Now()

	// Register all RPC functions defined in this server
	if err := registerRPCS(initializer, logger); err != nil {
		return err
	}

	// Register a hook that intercepts custom authentication before execution
	if err := initializer.RegisterBeforeAuthenticateCustom(auth.BeforeAuthenticateCustom); err != nil {
		logger.Error("Unable to register: %v", err)
		return err
	}

	// Create or verify the existence of the monthly leaderboard
	if err := leaderboard.CreateMonthlyLeaderboard(ctx, nk, logger); err != nil {
		return err
	}

	logger.Info("Module loaded in %dms", time.Since(initStart).Milliseconds())
	return nil
}

// registerRPCS registers custom RPC handlers with Nakama's initializer.
func registerRPCS(initializer runtime.Initializer, logger runtime.Logger) error {
	if err := initializer.RegisterRpc("account_delete_id", account.DeleteId); err != nil {
		logger.Error("Unable to register: %v", err)
		return err
	}

	if err := initializer.RegisterRpc("get_friendship_state", friend.GetFriendshipState); err != nil {
		logger.Error("Unable to register: %v", err)
		return err
	}

	if err := initializer.RegisterRpc("get_group_by_id", group.GetGroupById); err != nil {
		logger.Error("Unable to register: %v", err)
		return err
	}

	if err := initializer.RegisterRpc("get_group_membership_state", group.GetGroupMembershipState); err != nil {
		logger.Error("Unable to register: %v", err)
		return err
	}

	return nil
}
