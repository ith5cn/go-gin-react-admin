package system

import (
	"errors"
	"testing"

	redisInit "server/setup/redis"
	"server/utils"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

func TestLoginCredentials(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, password         string
		status                 int
		missing, deleted, want bool
	}{
		{"success", "correct", 1, false, false, true},
		{"wrong password", "wrong", 1, false, false, false},
		{"disabled", "correct", 2, false, false, false},
		{"missing", "correct", 1, true, false, false},
		{"deleted", "correct", 1, false, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, mock := userTestDB(t)
			t.Setenv("JWT_SECRET", "unit-test-secret")
			store := miniredis.RunT(t)
			client := redis.NewClient(&redis.Options{Addr: store.Addr()})
			previous := redisInit.Redis.Client
			redisInit.Redis.Client = client
			t.Cleanup(func() { client.Close(); redisInit.Redis.Client = previous })
			rows := sqlmock.NewRows([]string{"id", "username", "password", "status"})
			if !tc.missing && !tc.deleted {
				rows.AddRow(7, "alice", string(hash), tc.status)
			}
			mock.ExpectQuery("SELECT .*ai_system_user.*username = \\? AND delete_time IS NULL").WithArgs("alice", 1).WillReturnRows(rows)
			pair, err := loginInternal("alice", tc.password)
			if !tc.want {
				if !errors.Is(err, ErrLoginFailed) {
					t.Fatalf("error=%v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if _, err := utils.ValidateToken(pair.AccessToken, utils.TokenTypeAccess); err != nil {
				t.Fatal(err)
			}
		})
	}
}
