package legacy

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"server/internal/testutil"
)

type rejectedSession struct{}

func (rejectedSession) VerifySession(context.Context, string) (uint, error) {
	return 0, errors.New("expired")
}

func TestInvalidSessionNeverQueriesUsers(t *testing.T) {
	db, _ := testutil.DB(t)
	if _, err := (Bridge{DB: db, Sessions: rejectedSession{}}).Exchange(context.Background(), "opaque-session"); err == nil {
		t.Fatal("expired session accepted")
	}
}

func TestLegacyFieldsPreserveLargeIDs(t *testing.T) {
	var input LegacyUser
	if err := json.Unmarshal([]byte(`{"user_id":"9007199254740993","created_at":1700000000,"status":"1"}`), &input); err != nil {
		t.Fatal(err)
	}
	user, err := input.Normalize()
	if err != nil {
		t.Fatal(err)
	}
	if user.ID != 9007199254740993 || !user.Enabled || user.CreatedAt.Unix() != 1700000000 {
		t.Fatalf("fields=%#v", user)
	}
	input.Status = "unexpected"
	if _, err := input.Normalize(); err == nil {
		t.Fatal("invalid status accepted")
	}
}
