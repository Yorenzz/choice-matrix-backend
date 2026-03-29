package auth

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RefreshStore struct {
	client *redis.Client
}

func NewRefreshStore(client *redis.Client) *RefreshStore {
	return &RefreshStore{client: client}
}

func (s *RefreshStore) key(sessionID string) string {
	return fmt.Sprintf("refresh_session:%s", sessionID)
}

func (s *RefreshStore) Save(ctx context.Context, sessionID string, userID uint, ttl time.Duration) error {
	return s.client.Set(ctx, s.key(sessionID), strconv.FormatUint(uint64(userID), 10), ttl).Err()
}

func (s *RefreshStore) GetUserID(ctx context.Context, sessionID string) (uint, error) {
	value, err := s.client.Get(ctx, s.key(sessionID)).Result()
	if err != nil {
		return 0, err
	}

	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, err
	}

	return uint(parsed), nil
}

func (s *RefreshStore) Delete(ctx context.Context, sessionID string) error {
	return s.client.Del(ctx, s.key(sessionID)).Err()
}
