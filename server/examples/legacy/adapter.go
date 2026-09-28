// Package legacy demonstrates a PHP compatibility boundary. It registers no public routes.
package legacy

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"server/pkg/query"
	"server/utils"

	"gorm.io/gorm"
)

// SessionVerifier must validate expiry/revocation server-side and map the legacy identity to a local user.
// Implement against the old system's session store or authenticated introspection API, never a client ID.
type SessionVerifier interface {
	VerifySession(context.Context, string) (uint, error)
}

// Bridge explicitly receives its database and trusted PHP session verifier.
type Bridge struct {
	DB       *gorm.DB
	Sessions SessionVerifier
}

// Exchange issues normal Redis-backed JWTs only for a verified, active local account.
// A real endpoint must also have rate limits and the old system's CSRF protection if it uses cookies.
func (b Bridge) Exchange(ctx context.Context, session string) (*utils.TokenPair, error) {
	if b.DB == nil || b.Sessions == nil || session == "" {
		return nil, errors.New("legacy session unavailable")
	}
	id, err := b.Sessions.VerifySession(ctx, session)
	if err != nil || id == 0 {
		return nil, errors.New("invalid legacy session")
	}
	var user struct {
		ID       uint
		Username string
	}
	if err := b.DB.WithContext(ctx).Table("ai_system_user").Select("id, username").
		Where("id = ? AND status = 1 AND delete_time IS NULL", id).Take(&user).Error; err != nil {
		return nil, err
	}
	return utils.GenerateToken(user.ID, user.Username)
}

// ListRequest represents a common PHP paging contract; expose it only in an explicit compatibility route.
type ListRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// Query maps legacy input into the shared pagination contract.
func (r ListRequest) Query() map[string]string {
	p := query.ParsePage(map[string]string{"page": strconv.Itoa(r.Page), "size": strconv.Itoa(r.PageSize)})
	return map[string]string{"page": strconv.Itoa(p.Page), "size": strconv.Itoa(p.Size)}
}

// LegacyUser keeps string IDs and Unix timestamps at the old API boundary.
type LegacyUser struct {
	ID        json.Number `json:"user_id"`
	CreatedAt int64       `json:"created_at"`
	Status    string      `json:"status"`
}

// UserFields is a new service DTO, separate from the PHP response representation.
type UserFields struct {
	ID        uint64
	CreatedAt time.Time
	Enabled   bool
}

// Normalize validates legacy fields without round-tripping large identifiers through float64.
func (u LegacyUser) Normalize() (UserFields, error) {
	id, err := strconv.ParseUint(string(u.ID), 10, 64)
	if err != nil || id == 0 {
		return UserFields{}, errors.New("invalid user_id")
	}
	status := strings.TrimSpace(u.Status)
	if status != "1" && status != "2" {
		return UserFields{}, errors.New("invalid legacy status")
	}
	if u.CreatedAt < 0 {
		return UserFields{}, errors.New("invalid created_at")
	}
	return UserFields{ID: id, CreatedAt: time.Unix(u.CreatedAt, 0).UTC(), Enabled: status == "1"}, nil
}
