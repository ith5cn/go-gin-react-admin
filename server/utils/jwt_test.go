package utils

import (
	"errors"
	"testing"
	"time"

	"server/config"
	redisInit "server/setup/redis"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func testTokenStore(t *testing.T) *miniredis.Miniredis {
	t.Helper()
	t.Setenv("JWT_SECRET", "unit-test-secret-never-use-in-production")
	t.Setenv("JWT_LOGIN_MODE", "multi")
	store := miniredis.RunT(t)
	previous := redisInit.Redis.Client
	client := redis.NewClient(&redis.Options{Addr: store.Addr()})
	redisInit.Redis.Client = client
	t.Cleanup(func() { _ = client.Close(); redisInit.Redis.Client = previous })
	return store
}

func TestTokenRefreshAndRevocation(t *testing.T) {
	testTokenStore(t)
	pair, err := GenerateToken(7, "alice")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := ValidateToken(pair.AccessToken, TokenTypeAccess)
	if err != nil || claims.UserID != 7 {
		t.Fatalf("claims=%v err=%v", claims, err)
	}
	if _, err := RefreshToken(pair.AccessToken); !errors.Is(err, ErrInvalidTokenType) {
		t.Fatalf("access refreshed: %v", err)
	}
	if _, err := ValidateToken(pair.RefreshToken, TokenTypeAccess); !errors.Is(err, ErrInvalidTokenType) {
		t.Fatalf("refresh accepted as access: %v", err)
	}
	next, err := RefreshToken(pair.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	if next.AccessJTI == pair.AccessJTI {
		t.Fatal("refresh reused jti")
	}
	if err := RevokeToken(next.AccessToken); err != nil {
		t.Fatal(err)
	}
	if _, err := ValidateToken(next.AccessToken, TokenTypeAccess); !errors.Is(err, ErrTokenRevoked) {
		t.Fatalf("revoked token: %v", err)
	}
	if err := RevokeToken(pair.RefreshToken); err != nil {
		t.Fatal(err)
	}
	if _, err := RefreshToken(pair.RefreshToken); !errors.Is(err, ErrTokenRevoked) {
		t.Fatalf("revoked refresh: %v", err)
	}
}

func TestTokensFailClosed(t *testing.T) {
	testTokenStore(t)
	expired, _, err := signToken(7, "alice", TokenTypeAccess, -time.Minute, config.JwtConfig())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ValidateToken(expired, TokenTypeAccess); !errors.Is(err, ErrTokenExpired) {
		t.Fatalf("expired: %v", err)
	}
	pair, err := GenerateToken(7, "alice")
	if err != nil {
		t.Fatal(err)
	}
	redisInit.Redis.Client = nil
	if _, err := ValidateToken(pair.AccessToken, TokenTypeAccess); !errors.Is(err, ErrRedisNotInitialized) {
		t.Fatalf("redis unavailable: %v", err)
	}
}

func TestSingleLoginReplacesPreviousTokens(t *testing.T) {
	testTokenStore(t)
	t.Setenv("JWT_LOGIN_MODE", "single")
	first, err := GenerateToken(7, "alice")
	if err != nil {
		t.Fatal(err)
	}
	second, err := GenerateToken(7, "alice")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ValidateToken(first.AccessToken, TokenTypeAccess); !errors.Is(err, ErrTokenRevoked) {
		t.Fatalf("old access: %v", err)
	}
	if _, err := RefreshToken(first.RefreshToken); !errors.Is(err, ErrTokenRevoked) {
		t.Fatalf("old refresh: %v", err)
	}
	if _, err := ValidateToken(second.AccessToken, TokenTypeAccess); err != nil {
		t.Fatal(err)
	}
}
