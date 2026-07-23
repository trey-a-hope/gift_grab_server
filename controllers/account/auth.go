package account

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
)

var (
	ErrJwtVerification = runtime.NewError("jwt verification failed", 3)
)

type Claims struct {
	Id       string `json:"id"`
	Username string `json:"username"`
}

func BeforeAuthenticateCustom(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, in *api.AuthenticateCustomRequest) (*api.AuthenticateCustomRequest, error) {
	// Get the JWT secret key from the runtime context
	env, ok := ctx.Value(runtime.RUNTIME_CTX_ENV).(map[string]string)
	if !ok {
		logger.Error("failed to get env from context")
		return nil, ErrJwtVerification
	}

	secretKey := env["JWT_SECRET_KEY"]
	if secretKey == "" {
		logger.Error("JWT_SECRET_KEY environment variable is not set")
		return nil, ErrJwtVerification
	}

	var claims Claims
	err := VerifyAndParseJwt(secretKey, in.Account.Id, &claims)
	if err != nil {
		logger.Error("error verifying and parsing jwt: %s", err.Error())
		return nil, ErrJwtVerification
	}

	// Update the incoming authenticate request with the user ID and username from the JWT claims
	in.Account.Id = claims.Id
	in.Username = claims.Username

	return in, nil
}

func VerifyAndParseJwt(secretKey string, tokenString string, claims *Claims) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Support both RS256 (Clerk standard) and HMAC (shared secret)
		if _, ok := token.Method.(*jwt.SigningMethodRSA); ok {
			return jwt.ParseRSAPublicKeyFromPEM([]byte(secretKey))
		}
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); ok {
			return []byte(secretKey), nil
		}
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	})

	if err != nil {
		return err
	}

	if !token.Valid {
		return errors.New("invalid token")
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("invalid claims type")
	}

	// Clerk and standard JWTs use the "sub" claim to represent the user ID
	if sub, ok := mapClaims["sub"].(string); ok {
		claims.Id = sub
	} else if id, ok := mapClaims["id"].(string); ok {
		claims.Id = id
	}

	// Optional username or other fields
	if username, ok := mapClaims["username"].(string); ok {
		claims.Username = username
	}

	return nil
}
