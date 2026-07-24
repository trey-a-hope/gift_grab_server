package account

import (
	"context"
	"crypto/rsa"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
)

var (
	ErrJwtVerification = runtime.NewError("jwt verification failed", 3)

	jwksMu        sync.RWMutex
	jwksKeys      = map[string]*rsa.PublicKey{}
	jwksURLCached string
	jwksFetchedAt time.Time
)

const jwksCacheTTL = 1 * time.Hour

type Claims struct {
	Id       string `json:"id"`
	Username string `json:"username"`
}

type jwk struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type jwksResponse struct {
	Keys []jwk `json:"keys"`
}

func fetchJWKS(jwksURL string) (map[string]*rsa.PublicKey, error) {
	resp, err := http.Get(jwksURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected JWKS status code: %d", resp.StatusCode)
	}

	var parsed jwksResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("failed to decode JWKS response: %w", err)
	}

	keys := make(map[string]*rsa.PublicKey, len(parsed.Keys))
	for _, k := range parsed.Keys {
		if k.Kty != "RSA" {
			continue
		}
		pubKey, err := jwkToRSAPublicKey(k)
		if err != nil {
			continue
		}
		keys[k.Kid] = pubKey
	}

	return keys, nil
}

func jwkToRSAPublicKey(k jwk) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}

	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: int(new(big.Int).SetBytes(eBytes).Int64()),
	}, nil
}

// getKey returns the RSA public key for the given kid, refreshing the JWKS cache as needed.
func getKey(jwksURL, kid string) (*rsa.PublicKey, error) {
	jwksMu.RLock()
	stale := time.Since(jwksFetchedAt) > jwksCacheTTL || jwksURLCached != jwksURL
	key, found := jwksKeys[kid]
	jwksMu.RUnlock()

	if found && !stale {
		return key, nil
	}

	keys, err := fetchJWKS(jwksURL)
	if err != nil {
		if found {
			return key, nil // fall back to stale cache rather than failing outright
		}
		return nil, err
	}

	jwksMu.Lock()
	jwksKeys = keys
	jwksURLCached = jwksURL
	jwksFetchedAt = time.Now()
	jwksMu.Unlock()

	key, found = keys[kid]
	if !found {
		return nil, fmt.Errorf("no matching key found for kid %q", kid)
	}

	return key, nil
}

func BeforeAuthenticateCustom(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, in *api.AuthenticateCustomRequest) (*api.AuthenticateCustomRequest, error) {
	env, ok := ctx.Value(runtime.RUNTIME_CTX_ENV).(map[string]string)
	if !ok {
		errMessage := "failed to get env from context"
		logger.Error(errMessage)
		return nil, fmt.Errorf("%w: %s", ErrJwtVerification, errMessage)
	}

	jwksURL := env["CLERK_JWKS_URL"]
	if jwksURL == "" {
		logger.Error("CLERK_JWKS_URL environment variable is not set")
		return nil, ErrJwtVerification
	}

	var claims Claims
	err := VerifyAndParseJwt(jwksURL, in.Account.Id, &claims)
	if err != nil {
		logger.Error("error verifying and parsing jwt: %s", err.Error())
		return nil, ErrJwtVerification
	}

	in.Account.Id = claims.Id
	in.Username = claims.Username

	return in, nil
}

func VerifyAndParseJwt(jwksURL string, tokenString string, claims *Claims) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("token header missing kid")
		}

		return getKey(jwksURL, kid)
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

	if sub, ok := mapClaims["sub"].(string); ok {
		claims.Id = sub
	} else if id, ok := mapClaims["id"].(string); ok {
		claims.Id = id
	}

	if username, ok := mapClaims["username"].(string); ok {
		claims.Username = username
	}

	return nil
}