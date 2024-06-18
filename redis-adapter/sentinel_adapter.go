package redis_adapter

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/weloe/token-go/persist"
)

var _ persist.Adapter = (*SentinelAdapter)(nil)

var _ persist.BatchAdapter = (*SentinelAdapter)(nil)

type SentinelAdapter struct {
	*RedisAdapter
}

func (r *SentinelAdapter) GetCountsFilteredKey(filterKeyPrefix string) (int, error) {
	keys, err := r.client.Keys(context.Background(), filterKeyPrefix).Result()
	if err != nil {
		return 0, err
	}
	return len(keys), nil
}

// NewSentinelAdapter adapter for sentinel mode
func NewSentinelAdapter(masterName string, addrs []string, username string, password string, db int) *SentinelAdapter {
	return NewSentinelAdapterByOptions(&redis.FailoverOptions{
		MasterName:    masterName,
		SentinelAddrs: addrs,
		Username:      username,
		Password:      password,
		DB:            db,
	})
}

func NewSentinelAdapterByOptions(options *redis.FailoverOptions) *SentinelAdapter {
	return &SentinelAdapter{&RedisAdapter{client: redis.NewFailoverClient(options), serializer: persist.NewJsonSerializer()}}
}
